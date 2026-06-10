package core

import "github.com/engine-go/workflow/core/graph"

// WorkFlowParam 工作流请求参数
type WorkFlowParam struct {
	ExeParam *ExeParam `json:"uid"`                 // 流程执行参数，无业务数据
	DataInfo any       `json:"data_info,omitempty"` // 业务数据
}
type ExeParam struct {
	UID            int64  `json:"uid"`                        // 用户 id
	GraphId        string `json:"graph_id"`                   // 流程图 id
	FlowInstanceId string `json:"flow_instance_id,omitempty"` // 工作流实例 id（首次为空）
	NodeID         string `json:"node_id"`                    // 触发的节点
	TaskId         int64  `json:"task_id"`                    // 任务重试的时候需要传
	Type           string `json:"type"`                       // 是否保存过程数据
}

// UserFlow 用户工作流参数
type UserFlow struct {
	UID         int64        //  用户 id
	GraphId     string       // 工作流图id
	InstanceID  string       // 实例id
	Status      int          // 状态，进展，用户维度
	CurrGraphID string       // 当前命中的流程id
	CurNodeIDs  []string     // 当前进行到的节点id
	G           *graph.Graph // 工作流图信息
	DataInfo    any          // 被编排的节点使用
}
