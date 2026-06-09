package models

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
	BaseModel
}

func (WfGraph) TableName() string {
	return "wf_graph"
}
