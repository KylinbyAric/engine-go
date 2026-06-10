package core

import (
	"context"

	"github.com/engine-go/workflow/repository/models"
)

// RunSvcPipe 执行服务编排(无实例化操作)
// 1）功能介绍：1、根据流程id查询工作流配置 2、基于配置执行流程 3、同步返回执行结果
// 2）使用场景:无异步节点、结果同步返回的场景使用
// 3）参数说明：param- uid 、GraphId、DataInfo 必传
func RunSvcPipe(ctx context.Context, param *WorkFlowParam) error {
	param.ExeParam.Type = models.WfGraphTypeSvcPipe
	_, err := StartProc(ctx, param)
	if err != nil {
		return err
	}
	return nil
}
