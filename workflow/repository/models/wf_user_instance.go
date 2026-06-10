package models

type WfUserInstanceStatus int

const (
	WfUserInstanceStatusInit    WfUserInstanceStatus = 1 // 初始态
	WfUserInstanceStatusRunning WfUserInstanceStatus = 2 // 开始执行
	WfUserInstanceStatusSuccess WfUserInstanceStatus = 3 // 执行成功
	WfUserInstanceStatusFailed  WfUserInstanceStatus = 4 // 执行失败
	WfUserInstanceStatusError   WfUserInstanceStatus = 5 // 异常
)

func (t WfUserInstanceStatus) String() string {
	switch t {
	case WfUserInstanceStatusInit:
		return "待执行"
	case WfUserInstanceStatusRunning:
		return "执行中"
	case WfUserInstanceStatusSuccess:
		return "执行成功"
	case WfUserInstanceStatusFailed:
		return "执行失败"
	case WfUserInstanceStatusError:
		return "执行异常"
	default:
		return "未知"
	}
}

// WfUserInstance 用户工作流实例
type WfUserInstance struct {
	ID         int64                `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UID        int64                `gorm:"column:uid;not null" json:"uid"`
	InstanceID string               `gorm:"column:instance_id;size:255" json:"instance_id"`
	Name       string               `gorm:"column:name;size:255;not null" json:"name"`
	GraphID    string               `gorm:"column:graph_id;type:text;not null" json:"graph_id"`
	Graph      string               `gorm:"column:graph;type:text;not null" json:"graph"`
	Status     WfUserInstanceStatus `gorm:"column:status;not null" json:"status"`
	StatusMsg  string               `gorm:"column:status_msg;size:255;not null" json:"status_msg"`
	CurrStepID string               `gorm:"column:curr_step_id;size:255;not null;default:''" json:"curr_step_id"`
	DataInfo   string               `gorm:"column:data_info;type:text;not null" json:"data_info"`
	BaseModel
}

func (WfUserInstance) TableName() string {
	return "wf_user_instance"
}
