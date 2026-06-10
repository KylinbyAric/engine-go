package models

import "time"

type WfTaskStatus int

const (
	WfTaskStatusInit    WfTaskStatus = 1 // 执行中
	WfTaskStatusSuccess WfTaskStatus = 2 // 成功完成
	WfTaskStatusFail    WfTaskStatus = 3 // 失败
	WfTaskStatusErr     WfTaskStatus = 4 // 异常
)

func (t WfTaskStatus) String() string {
	switch t {
	case WfTaskStatusInit:
		return "执行中"
	case WfTaskStatusSuccess:
		return "成功"
	case WfTaskStatusFail:
		return "失败"
	case WfTaskStatusErr:
		return "异常"
	default:
		return "未知"
	}
}

// WfTask 工作流任务执行记录
// 注意：本表无 create_by / update_by，因此不嵌入 BaseModel，公共字段手写。
type WfTask struct {
	ID               int64        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID              int64        `gorm:"column:uid;not null" json:"uid"`
	GraphID          string       `gorm:"column:graph_id;size:255;not null" json:"graph_id"`
	InstanceID       string       `gorm:"column:instance_id;size:255" json:"instance_id"`
	NodeID           string       `gorm:"column:node_id;type:text;not null" json:"node_id"`
	NodeType         string       `gorm:"column:node_type;size:255" json:"node_type"`
	Status           WfTaskStatus `gorm:"column:status;not null" json:"status"`
	StatusMsg        string       `gorm:"column:status_msg;size:255;not null" json:"status_msg"`
	NodeParam        string       `gorm:"column:node_param;size:512;not null" json:"node_param"`
	BeforeParam      string       `gorm:"column:before_param;type:text" json:"before_param"`
	AfterParam       string       `gorm:"column:after_param;type:text" json:"after_param"`
	TaskResponse     string       `gorm:"column:task_response;type:text;not null" json:"task_response"`
	CallbackResponse string       `gorm:"column:callback_response;type:text;not null" json:"callback_response"`
	RunCount         int          `gorm:"column:run_count;default:0" json:"run_count"`
	StartTime        *time.Time   `gorm:"column:start_time" json:"start_time,omitempty"`
	EndTime          *time.Time   `gorm:"column:end_time" json:"end_time,omitempty"`
	IsDelete         int8         `gorm:"column:is_delete;default:0" json:"is_delete"`
	CreatedAt        time.Time    `gorm:"column:create_at;<-:false" json:"created_at"`
	UpdatedAt        time.Time    `gorm:"column:update_at;<-:false" json:"updated_at"`
}

func (WfTask) TableName() string {
	return "wf_task"
}
