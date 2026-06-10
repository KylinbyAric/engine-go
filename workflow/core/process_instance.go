package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/engine-go/workflow/common"
	log "github.com/engine-go/workflow/common/log"
	"github.com/engine-go/workflow/core/graph"
	"github.com/engine-go/workflow/core/node"
	"github.com/engine-go/workflow/repository/dao"
	"github.com/engine-go/workflow/repository/models"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type nodeRunInfo struct {
	CurRunNodes    []node.NodeProcessor // 待执行的节点列表，所以的依赖节点都已经执行成功
	flow           *UserFlow            // 用户工作流实例，UID/InstanceID/GraphId/G 等均从此获取
	wfParam        *WorkFlowParam       // 持有引用，节点执行中对 DataInfo 的修改自动同步回调用方
	CurrStepID     []string             // 执行到的终点节点id
	isStoreMidData bool                 // 是否保存过程数据
}

// StartProc 流程实例化，且执行
func StartProc(ctx context.Context, param *WorkFlowParam) (*UserFlow, error) {
	f, err := CreateUserWorkflow(ctx, param)
	if err != nil {
		log.Errorf("msg=流程实例化失败||param=%v||err=%v", param, err)
		return nil, err
	}
	return f, runInstance(ctx, f, param)
}

func runInstance(ctx context.Context, f *UserFlow, param *WorkFlowParam) error {
	// 获取当前执行节点
	curNodes, err := getCurrentNode(f)
	if err != nil {
		return err
	}
	// 2.节点执行
	nInfo := newNodeRunInfo(curNodes, f, param)

	if nInfo.isStoreMidData {
		uiDao := dao.NewWfUserInstanceDao(nil)
		if err := uiDao.UpdateStatus(ctx, f.UID, f.InstanceID, models.WfUserInstanceStatusRunning); err != nil {
			return err
		}
		if err := nInfo.Run(ctx); err != nil {
			if err2 := uiDao.UpdateStatus(ctx, f.UID, f.InstanceID, models.WfUserInstanceStatusError); err2 != nil {
				log.Errorf("msg=回写错误状态失败||param=%v||UID=%v||InstanceID=%v||run_err=%v||status_err=%v",
					param, f.UID, f.InstanceID, err, err2)
			}
			return err
		}
		// 业务已执行成功：先持久化业务数据，再回写 Success。
		// 若数据写入失败，需要业务侧重试 → 保留 error；
		// 若数据已落但 Success 状态写入失败，只 log 不中断 — 业务结果真实，状态由后续心跳/查询补偿。
		if err := uiDao.SaveUserWorkFlowData(ctx, f.UID, f.InstanceID, param.DataInfo); err != nil {
			log.Errorf("msg=保存业务数据失败||param=%v||runInfo=%v||err=%v", param, nInfo, err)
			return err
		}
		if err := uiDao.UpdateStatus(ctx, f.UID, f.InstanceID, models.WfUserInstanceStatusSuccess); err != nil {
			log.Errorf("msg=业务已完成但回写Success状态失败(需补偿)||param=%v||UID=%v||InstanceID=%v||err=%v",
				param, f.UID, f.InstanceID, err)
		}
		return nil
	}

	err = nInfo.Run(ctx) // 执行
	if err != nil {
		log.Errorf("msg=流程执行失败||param=%v||runInfo=%v||err=%v", param, nInfo, err)
		return err
	}
	log.Infof("msg=流程执行完成||param=%v||runInfo=%v", param, nInfo)
	return nil
}

func (n *nodeRunInfo) Run(ctx context.Context) error {
	if len(n.CurRunNodes) == 0 {
		log.Infof("msg=无待执行节点||param=%v", common.StructToString(n))
		return nil
	}

	for {
		curNode, isExist := n.getNextExecuted()
		if !isExist {
			log.Infof("msg=无待执行节点||param=%v", common.StructToString(n))
			return nil
		}
		task, err := n.createTask(ctx, curNode)
		if err != nil {
			log.Errorf("msg=创建任务失败||node=%v||err=%v", curNode, err)
			return err
		}
		// 2.2 执行节点
		p := &node.NodeParam{
			RunParam:   n.wfParam.DataInfo,
			UID:        n.flow.UID,
			InstanceID: n.flow.InstanceID,
			GraphID:    n.flow.GraphId,
			TaskId: func() string {
				if task != nil {
					return cast.ToString(task.ID)
				}
				return ""
			}(),
		}
		sta, nexNodeId, err := curNode.Process(ctx, p)
		n.wfParam.DataInfo = p.RunParam
		log.Debugf("msg=节点执行结果||uid=%v||GraphID=%v||node_id=%v||node_type=%v||nex_node_id=%v||node_status=%v||err=%v", n.flow.UID, n.flow.GraphId, curNode.GetNodeID(), curNode.GetType(), nexNodeId, sta, err)
		p.LastNodeErr = err // 保存节点错误信息，用于下个节点判断使用
		curNode.SetStatus(sta)
		if err != nil {
			return err
		}
		n.updateTask(ctx, task, curNode) // 2.3 更新节点状态
		n.addNextPool(nexNodeId)         // 执行成功，则执行下个节点

		if sta == node.NodeStatusRunning { // 等待执行结果。异步节点，会等待结果
			n.runningHandle(ctx, curNode)
		}
		if sta == node.NodeStatusFail { // 执行失败，业务逻辑失败
		}
		if sta == node.NodeStatusErr { // 执行系统异常
			n.errHandle(ctx, curNode)
		}
	}
}

func (n *nodeRunInfo) createTask(ctx context.Context, curNode node.NodeProcessor) (*models.WfTask, error) {
	if !(n.isStoreMidData && curNode.GetReqMode() == node.RequestModeAsync) { // 不保存请求、 同步节点 不保存过程数据
		return nil, nil
	}
	t, err := NodeToTaskModel(ctx, n.flow.UID, n.flow.GraphId, n.flow.InstanceID, curNode, n.wfParam.DataInfo)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// NodeToTaskModel node节点创建任务
func NodeToTaskModel(ctx context.Context, uid int64, graphId, instanceId string, curNode node.NodeProcessor, param any) (*models.WfTask, error) {
	t := &models.WfTask{
		InstanceID:  instanceId,
		UID:         uid,
		GraphID:     graphId,
		NodeID:      curNode.GetNodeID(),
		NodeType:    string(curNode.GetType()),
		Status:      models.WfTaskStatusInit,
		StatusMsg:   models.WfTaskStatusInit.String(),
		BeforeParam: common.StructToString(param),
	}
	taskDao := dao.NewWfTaskDao(nil)
	err := taskDao.Create(ctx, t)
	if err == nil {
		return t, nil
	}
	// 唯一键 (uid, graph_id, instance_id, node_id) 冲突 → 取已存在的那条
	if !isDuplicatedKey(err) {
		return nil, err
	}
	existing, gerr := taskDao.GetByUniqueKey(ctx, uid, graphId, instanceId, t.NodeID)
	if gerr != nil {
		return nil, gerr
	}
	if existing == nil {
		return nil, errors.New("wf_task: duplicate key but record not found")
	}
	return existing, nil
}

// isDuplicatedKey 优先用 GORM 的 sentinel；某些场景（gorm 版本 / 驱动差异）会回退到字符串匹配。
func isDuplicatedKey(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	// MySQL: "Error 1062: Duplicate entry ..."
	return false
}

// 取出获取执行节点
func (n *nodeRunInfo) getNextExecuted() (node.NodeProcessor, bool) {
	if len(n.CurRunNodes) == 0 {
		return nil, false
	}
	cur := n.CurRunNodes[len(n.CurRunNodes)-1]
	n.CurRunNodes = n.CurRunNodes[:len(n.CurRunNodes)-1]
	return cur, true
}

// addNextPool 添加执行节点
func (n *nodeRunInfo) addNextPool(list []string) {
	for _, nodeId := range list {
		if curNode, ok := n.flow.G.NodeMap[nodeId]; ok {
			if slice.Every(curNode.GetDependents(), func(index int, item string) bool {
				if denode, ok2 := n.flow.G.NodeMap[item]; ok2 {
					return denode.GetStatus() == node.NodeStatusSucc || denode.GetStatus() == node.NodeStatusFail
				}
				return false
			}) {
				n.CurrStepID = append(n.CurrStepID, nodeId)
			}
		}
	}
}

// runningHandle 异步节点处理
func (n *nodeRunInfo) runningHandle(ctx context.Context, curNode node.NodeProcessor) {

}

// errHandle 异常处理
func (n *nodeRunInfo) errHandle(ctx context.Context, exeNode any) {

}

func newNodeRunInfo(nodes []node.NodeProcessor, f *UserFlow, param *WorkFlowParam) *nodeRunInfo {
	return &nodeRunInfo{
		CurRunNodes:    nodes,
		flow:           f,
		wfParam:        param,
		isStoreMidData: f.G.Type != models.WfGraphTypeSvcPipe, // 服务编排不留痕
	}
}

// 更新执行结果
func (n *nodeRunInfo) updateTask(ctx context.Context, t *models.WfTask, curNode node.NodeProcessor) {
	if t == nil {
		return
	}
	t.Status = curNode.GetStatus().ToTaskStatus()
	t.AfterParam = common.StructToString(n.wfParam.DataInfo)
	t.RunCount++
	err := dao.NewWfTaskDao(nil).Update(ctx, t)
	if err != nil {
		log.Errorf("msg=更新任务状态失败||task_id=%v||AfterParam=%v||err=%v", t.ID, t.AfterParam, err)
	}
	return
}

func getCurrentNode(f *UserFlow) ([]node.NodeProcessor, error) {
	if len(f.CurNodeIDs) == 0 {
		inNodes := slice.Filter(f.G.Nodes, func(_ int, cNode node.NodeProcessor) bool {
			return cNode.GetType() == node.NodeIn
		})
		if len(inNodes) == 0 {
			return nil, errors.New("未有可执行节点")
		}
		return inNodes, nil
	}
	list := make([]node.NodeProcessor, 0, len(f.CurNodeIDs))
	for _, v := range f.CurNodeIDs {
		n, ok := f.G.NodeMap[v]
		if !ok {
			return nil, errors.New("节点丢失")
		}
		list = append(list, n)
	}
	return list, nil
}

// CreateUserWorkflow 创建用户工作流实例
func CreateUserWorkflow(ctx context.Context, param *WorkFlowParam) (*UserFlow, error) {
	if param == nil || param.ExeParam == nil {
		return nil, errors.New("WorkFlowParam.ExeParam is required")
	}
	// 1.获取的工作流图
	g, err := dao.NewWfGraphDao(nil).GetByGraphID(ctx, param.ExeParam.GraphId)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("graph_id 不存在: %s", param.ExeParam.GraphId)
	}
	if g.Type != param.ExeParam.Type {
		return nil, fmt.Errorf("图类型不匹配: 期望 %s, 实际 %s", param.ExeParam.Type, g.Type)
	}
	toGraph, err := graph.ModelToGraph(g)
	if err != nil {
		return nil, err
	}
	// 2. 创建用户实例
	f, err := buildUserWorkFlow(toGraph, param)
	if err != nil {
		log.Errorf("msg=构建表信息失败||param=%v||err=%v", param, err)
		return nil, err
	}

	if g.Type != models.WfGraphTypeSvcPipe { // 服务编排不留痕
		err = dao.NewWfUserInstanceDao(nil).Create(ctx, f)
		if err != nil {
			log.Errorf("msg=写用户实例表失败||param=%v||instance_flow=%v||err=%v", param, f, err)
			return nil, err
		}
	}
	uf, err := ModelToUserFlow(f)
	if err != nil {
		log.Errorf("msg=协议转换失败||param=%v||instance_flow=%v||err=%v", common.StructToString(param), common.StructToString(f), err)
		return nil, err
	}
	log.Infof("msg=创建用户流程实例成功||uid=%v||flow_type=%v||ExeParam=%v||user_flow=%v", param.ExeParam.UID, param.ExeParam.Type, common.StructToString(param.ExeParam), common.StructToString(uf))
	return uf, nil
}

// 构建用户流程实例
func buildUserWorkFlow(g *graph.Graph, param *WorkFlowParam) (*models.WfUserInstance, error) {
	var (
		istanceId string
	)
	if g.Type != models.WfGraphTypeSvcPipe {
		istanceId = uuid.New().String()
	}

	f := &models.WfUserInstance{
		InstanceID: istanceId,
		UID:        param.ExeParam.UID,
		Name:       fmt.Sprintf("s-%v-%v", g.Name, time.Now().Format(time.DateOnly)),
		GraphID:    g.GraphId,
		Graph:      common.StructToString(g),
		Status:     models.WfUserInstanceStatusInit,
		DataInfo:   common.StructToString(param.DataInfo),
	}
	return f, nil
}

// ModelToUserFlow db model转用户工作流实例
func ModelToUserFlow(p *models.WfUserInstance) (*UserFlow, error) {
	g, err := graph.ParseGraph(p.Graph)
	if err != nil {
		return nil, err
	}
	var d any
	err = sonic.UnmarshalString(p.DataInfo, &d)
	if err != nil {
		return nil, err
	}
	f := &UserFlow{
		UID:         p.UID,
		GraphId:     p.GraphID,
		InstanceID:  p.InstanceID,
		Status:      int(p.Status),
		CurrGraphID: p.GraphID,
		G:           g,
		DataInfo:    d,
	}
	return f, nil
}
