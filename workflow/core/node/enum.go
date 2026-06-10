package node

import "github.com/engine-go/workflow/repository/models"

type NodeType string

const (
	NodeAction    NodeType = "action"    // 动作节点
	NodeCondition NodeType = "condition" // 条件节点
	NodeState     NodeType = "state"     // 状态节点
	NodeIn        NodeType = "in"        // 入口节点

)

type NodeStatus int

const (
	NodeStatusToBeExe NodeStatus = 1 // 待执行
	NodeStatusRunning NodeStatus = 2 // 开始执行，等待结果
	NodeStatusSucc    NodeStatus = 3 // 执行成功
	NodeStatusFail    NodeStatus = 4 // 执行失败
	NodeStatusErr     NodeStatus = 5 // 执行错误
)

func (s NodeStatus) ToTaskStatus() models.WfTaskStatus {
	switch s {
	case NodeStatusToBeExe:
		return models.WfTaskStatusInit
	case NodeStatusRunning:
		return models.WfTaskStatusInit
	case NodeStatusSucc:
		return models.WfTaskStatusSuccess
	case NodeStatusFail:
		return models.WfTaskStatusFail
	case NodeStatusErr:
		return models.WfTaskStatusErr
	default:
		return -1
	}
}

type RequestMode string

const (
	RequestModeSync  RequestMode = "sync"  // 同步节点
	RequestModeAsync RequestMode = "async" // 异步节点
)

type (
	ActionNodeType string // 动作节点的类型，一个工作流里可能会调用多个类型的节点，比如规则引擎节点，因此需要用type字段
)
