package graph

import (
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/engine-go/workflow/common"
	"github.com/engine-go/workflow/core/node"
	"github.com/engine-go/workflow/core/node/action"
	"github.com/engine-go/workflow/core/node/condition"
	"github.com/engine-go/workflow/core/node/state"
	"github.com/engine-go/workflow/repository/models"
	errors2 "github.com/pkg/errors"
	"github.com/spf13/cast"
)

// Graph 工作流图定义
type Graph struct {
	GraphId   string                        `json:"graph_id,omitempty"`
	Name      string                        `json:"name,omitempty"`
	Type      string                        `json:"tyep,omitempty"` // 类型：svc_pipe:服务编排 flow_pipe:流程编排
	Desc      string                        `json:"desc,omitempty"`
	NodesJson []map[string]any              `json:"nodes,omitempty"` // 节点原始数据
	Nodes     []node.NodeProcessor          `json:"-"`               // 节点编译后
	NodeMap   map[string]node.NodeProcessor `json:"-"`               // 数据
}

func (g *Graph) hasCycle() (bool, error) {
	if len(g.Nodes) == 0 {
		return false, nil
	}
	inNodes := slice.Filter(g.Nodes, func(_ int, n node.NodeProcessor) bool { // 查找入口节点
		return n.GetType() == node.NodeIn
	})
	if len(inNodes) != 1 {
		return false, errors.New("未有可执行节点")
	}
	var visited = map[string]node.NodeProcessor{}
	var dfs func(curNode node.NodeProcessor) bool
	dfs = func(curNode node.NodeProcessor) bool {
		// 出现重复节点，说明有环
		if _, ok := visited[curNode.GetNodeID()]; ok {
			return true
		}
		// 标记当前节点
		visited[curNode.GetNodeID()] = curNode

		//
		if len(curNode.GetNextNodes()) == 0 {
			return false
		}

		defer delete(visited, curNode.GetNodeID())

		for _, nextNode := range curNode.GetNextNodes() {
			next, ok := g.NodeMap[nextNode]
			if !ok {
				continue // 实际这里应该抛错，不应该有不存在的节点
			}
			if dfs(next) {
				return true
			}
		}
		return false
	}

	return dfs(inNodes[0]), nil

}

// buildMap 构建节点map
func (g *Graph) buildMap() error {
	nodeMap := make(map[string]node.NodeProcessor)
	nodes := make([]node.NodeProcessor, 0, len(g.NodesJson))
	for _, v := range g.NodesJson {
		var temp node.NodeProcessor
		switch node.NodeType(cast.ToString(v["type"])) {
		case node.NodeAction:
			temp = new(action.ActionNode)
		case node.NodeCondition:
			temp = new(condition.ConditionNode)
		case node.NodeState:
			temp = new(state.StateNode)
		case node.NodeIn:
			temp = new(node.Node)
		default:
			return fmt.Errorf("未知的节点类型:[%v]", v["type"])
		}
		err := sonic.UnmarshalString(common.StructToString(v), temp)
		if err != nil {
			return err
		}
		nodes = append(nodes, temp)
		nodeMap[temp.GetNodeID()] = temp
	}
	g.NodeMap = nodeMap
	g.Nodes = nodes
	return nil
}

// buildDependency 构建依赖关系
func (g *Graph) buildDependency() {
	dependMap := make(map[string][]string, len(g.NodeMap))
	for nodeId, nodeProcessor := range g.NodeMap {
		for _, sucId := range nodeProcessor.GetNextNodes() {
			dependMap[sucId] = append(dependMap[sucId], nodeId)
		}
	}
	for nodeId, nodeProcessor := range g.NodeMap {
		nodeProcessor.SetDependents(dependMap[nodeId])
	}
	return
}

// checkNodeCount 检查图中的节点是否和总节点数量一致
func (g *Graph) checkNodeCount() error {
	curNodes := slice.Filter(g.Nodes, func(_ int, item node.NodeProcessor) bool { // 查找开始几点
		return item.GetType() == node.NodeIn
	})
	if len(curNodes) != 1 {
		return fmt.Errorf("入口节点数量:%v不等于1", len(curNodes))
	}

	var dfs func(curNode node.NodeProcessor) error
	var count int
	visitMap := make(map[string]node.NodeProcessor)

	dfs = func(curNode node.NodeProcessor) error {
		if _, ok := visitMap[curNode.GetNodeID()]; ok { // 有环，已经遍历过的剔除
			return nil
		}
		count++
		visitMap[curNode.GetNodeID()] = curNode
		for _, v := range curNode.GetNextNodes() {
			next, ok := g.NodeMap[v]
			if !ok {
				return fmt.Errorf("节点id:%v未知", v)
			}
			err := dfs(next)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err := dfs(curNodes[0])
	if err != nil {
		return err
	}
	if count != len(g.Nodes) {
		return fmt.Errorf("可以遍历节点数量 %v 和实际 %v 不一致", count, len(g.Nodes))
	}
	return nil
}

// ParseGraph 解析图配置
func ParseGraph(d string) (*Graph, error) {
	g := &Graph{}
	err := sonic.UnmarshalString(d, &g)
	if err != nil {
		return nil, err
	}
	err = g.buildMap()
	if err != nil {
		return nil, err
	}
	g.buildDependency()
	err = g.checkNodeCount()
	if err != nil {
		return g, err
	}
	return g, nil
}

func ModelToGraph(d *models.WfGraph) (*Graph, error) {
	g, err := ParseGraph(d.Graph)
	if err != nil {
		return nil, errors2.Wrap(err, "ParseGraph 失败")
	}
	g.GraphId = d.GraphID
	g.Name = d.Name
	g.Type = d.Type
	return g, nil
}
