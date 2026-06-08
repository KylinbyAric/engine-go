package action

import (
	"context"
	"fmt"

	"github.com/engine-go/workflow/core/node"
)

// ActionNode 动作节点
type ActionNode struct {
	node.Node
	ActionType    node.ActionNodeType `json:"action_type"`     // 动作节点类型，
	SuccessNodeId []string            `json:"success_node_id"` // 执行成功的节点id
	FailNodeId    []string            `json:"fail_node_id"`    // 执行失败的节点id 动作节点会有成功和失败节点
	//RouteEdges    []RouteEdge       `json:"route_edges"`     // 按路由键命中的后继节点 todo 没看懂干啥的
	Meta        string `json:"meta"` // 节点运行时数据 由具体的actionNode反序列化成结构体
	ActType     string
	RequestMode node.RequestMode
}

func (a *ActionNode) ProcessAction(ctx context.Context, nodeParam *node.NodeParam) (*ActionResult, error) {
	fmt.Printf("BaseAction: %v\n||actType=%v", nodeParam, a.GetActType())
	return &ActionResult{
		Success: true,
	}, nil
}

func (a *ActionNode) Process(ctx context.Context, nodeParam *node.NodeParam) (node.NodeStatus, []string, error) {
	h, err := a.getHandler()
	if err != nil {
		return node.NodeStatusErr, nil, err
	}
	res, err := h.ProcessAction(ctx, nodeParam)
	if err != nil {
		return node.NodeStatusErr, nil, err
	}
	if res.Success {
		return node.NodeStatusSucc, a.SuccessNodeId, nil
	}
	return node.NodeStatusFail, a.FailNodeId, nil
}

func (a *ActionNode) GetActType() string {
	return a.ActType
}

func (a *ActionNode) GetReqMode() node.RequestMode {
	return a.RequestMode
}

var (
	actionNodeTypeHandleMap = map[string]ActionNodeHandle{}
)

func (a *ActionNode) getHandler() (ActionNodeHandle, error) {
	h, ok := actionNodeTypeHandleMap[string(a.ActionType)]
	if !ok {
		return nil, fmt.Errorf("未知的动作类型:%v", a.ActionType)
	}
	return h, nil
}

// RegisterActionNode 注册动作节点
func RegisterActionNode(a ActionNodeHandle) {
	actionNodeTypeHandleMap[a.GetActType()] = a
}
