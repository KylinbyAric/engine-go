package node

import "context"

// NodeProcessor 节点执行器
type NodeProcessor interface {
	Process(ctx context.Context, p *NodeParam) (NodeStatus, []string, error) // Process 处理节点的核心方法
	GetNodeID() string                                                       // GetNodeID 获取节点ID
	GetType() NodeType                                                       // GetType 获取节点类型
	GetName() string                                                         // GetName 获取节点名称
	GetReqMode() RequestMode                                                 // 请求模式，同步/异步
	GetStatus() NodeStatus                                                   // GetStatus 获取节点状态
	SetStatus(status NodeStatus)                                             // SetStatus 设置节点状态
	GetNextNodes() []string                                                  // 获取继任节点
	GetDependents() []string                                                 // GetDependents 获取依赖节点 todo 暂时还没弄明白是干嘛的
	SetDependents(list []string)
}

// NodeParam 节点执行参数
type NodeParam struct {
	RunParam    any    // 用户请求参数
	UID         int64  // 用户id
	InstanceID  string // 工作流实例id
	GraphID     string // 工作流图id
	TaskId      string // 任务id，异步节点时会传
	LastNodeErr error  // 上个节点的执行错误
}
