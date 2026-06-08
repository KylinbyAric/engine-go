package node

import "context"

// Node 节点定义
type Node struct {
	NodeID     string      `json:"node_id,omitempty"` // 节点id
	Name       string      `json:"name,omitempty"`
	Type       NodeType    `json:"type"` // 节点类型
	Desc       string      `json:"desc"`
	NextNodes  []string    `json:"nextnodes,omitempty"`  // 下个节点id todo 为什么不是直接nodeProcessor?
	Dependents []string    `json:"dependents,omitempty"` // 依赖节点id，如果节点为空，则表示入口节点
	Status     NodeStatus  `json:"status"`               // 状态，进展
	ReqMode    RequestMode `json:"req_mode"`             // 状态，进展
}

// Process 处理节点的核心方法
func (n *Node) Process(ctx context.Context, p *NodeParam) (NodeStatus, []string, error) {
	return NodeStatusSucc, n.NextNodes, nil
}

// GetNodeID 获取节点ID
func (n *Node) GetNodeID() string {
	return n.NodeID
}

// GetType 获取节点类型
func (n *Node) GetType() NodeType {
	return n.Type
}

// GetName 获取节点名称
func (n *Node) GetName() string {
	return n.Name
}

// GetReqMode 请求模式，同步/异步
func (n *Node) GetReqMode() RequestMode {
	return RequestModeSync
}

// GetStatus 获取节点状态
func (n *Node) GetStatus() NodeStatus {
	return n.Status
}

// SetStatus 设置节点状态
func (n *Node) SetStatus(status NodeStatus) {
	n.Status = status
}

// GetNextNodes 获取继任节点
func (n *Node) GetNextNodes() []string {
	return n.NextNodes
}

// GetDependents 获取依赖节点
func (n *Node) GetDependents() []string {
	return n.Dependents
}

// SetDependents 设置依赖节点
func (n *Node) SetDependents(list []string) {
	n.Dependents = list
}
