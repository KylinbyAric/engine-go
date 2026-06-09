package models

import "time"

// BaseModel 1. 定义基础模板结构体（全局复用）
type BaseModel struct {
	CreateBy  string    `gorm:"column:create_by;<-:create" json:"create_by"` // 创建后不可改变
	UpdateBy  string    `gorm:"column:update_by" json:"update_by"`           // 更新人
	IsDelete  int8      `gorm:"column:is_delete" json:"is_delete"`           // 删除标识 0-未删除 1-已删除
	CreatedAt time.Time `gorm:"column:create_at;<-:false" json:"created_at"` // 禁止写入
	UpdatedAt time.Time `gorm:"column:update_at;<-:false" json:"updated_at"` // 禁止写入
}

const (
	RECORD_NORMAL  = 0 // 正常状态
	RECORD_DELETED = 1 // 已删除
)
