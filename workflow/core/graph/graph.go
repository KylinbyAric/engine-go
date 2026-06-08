package graph

import (
	"errors"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/engine-go/workflow/core/node"
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
