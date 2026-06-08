package action

import (
	"context"

	"github.com/engine-go/workflow/core/node"
)

// ActionNodeHandle 动作节点接口
type ActionNodeHandle interface {
	ProcessAction(ctx context.Context, nodeParam *node.NodeParam) (*ActionResult, error) // 执行动作接口，编排调用
	GetActType() string                                                                  // 动作key
	GetReqMode() node.RequestMode                                                        // 请求模式，同步/异步
	//CallbackHandle(ctx context.Context, envParam any, respData *DriveVary, m *RunMeta) (*ActionResult, error) // 回调处理
	//CheckParam(ctx context.Context, param any) (bool, error)                                                                // 参数检查
}

// ActionResult 动作节点执行结果。
// Success 表示是否命中成功主分支；
// RouteKeys 用于命中 ActionNode 上配置的 route_edges；
// NextNodes 用于动态追加后继节点，保留旧引擎的扩展能力。
type ActionResult struct {
	Success   bool     `json:"success"`    // 是否命中成功主分支
	NextNodes []string `json:"next_nodes"` // 动态追加的后继节点
}
