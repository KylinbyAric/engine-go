package models

// WfGraphRecord 工作流图配置的历史快照
type WfGraphRecord struct {
	ID          int64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GraphID     string        `gorm:"column:graph_id;size:64" json:"graph_id"`
	Name        string        `gorm:"column:name;size:255;not null" json:"name"`
	Description string        `gorm:"column:description;size:512;not null" json:"description"`
	Graph       string        `gorm:"column:graph;type:text;not null" json:"graph"`
	Version     int           `gorm:"column:version;not null" json:"version"`
	Type        string        `gorm:"column:type;size:64;not null" json:"type"`
	Status      WfGraphStatus `gorm:"column:status;not null" json:"status"`
	BaseModel
}

func (WfGraphRecord) TableName() string {
	return "wf_graph_record"
}
