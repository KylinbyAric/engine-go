package graph

import "github.com/engine-go/workflow/core/node"

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
