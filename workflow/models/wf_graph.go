package models

import "time"

type WfGraphStatus int

const (
	WfGraphStatusDraft   WfGraphStatus = 1
	WfGraphStatusActive  WfGraphStatus = 2
	WfGraphStatusOffline WfGraphStatus = 3
	WfGraphStatusDeleted WfGraphStatus = 4
)

const (
	WfGraphTypeSvcPipe  = "svc_pipe"
	WfGraphTypeFlowPipe = "flow_pipe"
)

type WfGraph struct {
	ID          int64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GraphID     string        `gorm:"column:graph_id;size:64" json:"graph_id"`
	Name        string        `gorm:"column:name;size:255;not null" json:"name"`
	Description string        `gorm:"column:description;size:512;not null" json:"description"`
	Graph       string        `gorm:"column:graph;type:text;not null" json:"graph"`
	Version     int           `gorm:"column:version;not null" json:"version"`
	Type        string        `gorm:"column:type;size:64;not null" json:"type"`
	RecordID    int64         `gorm:"column:record_id;not null" json:"record_id"`
	Status      WfGraphStatus `gorm:"column:status;not null" json:"status"`
	IsDelete    int8          `gorm:"column:is_delete;not null;default:0" json:"is_delete"`
	CreateBy    string        `gorm:"column:create_by;size:255;not null" json:"create_by"`
	UpdateBy    string        `gorm:"column:update_by;size:255;not null" json:"update_by"`
	CreatedAt   time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (WfGraph) TableName() string {
	return "wf_graph"
}
