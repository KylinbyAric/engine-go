package main

import (
	"github.com/engine-go/domain/workflow/node/action"
	"github.com/engine-go/workflow/models"
)

func init() {
	// 初始化 action 节点
	action.Init()
	// 初始化 MySQL（加载 conf/<env>/app.toml + gorm + ping）
	if err := models.Init(); err != nil {
		panic("init mysql: %v")
	}
}
