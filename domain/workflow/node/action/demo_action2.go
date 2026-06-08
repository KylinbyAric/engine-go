package action

import (
	"context"
	"fmt"

	"github.com/engine-go/workflow/core/node"
	"github.com/engine-go/workflow/core/node/action"
)

type DemoAction2 struct {
	action.ActionNode
}

func NewDemoAction2() action.ActionNodeHandle {
	return &DemoAction2{
		ActionNode: action.ActionNode{
			ActType:     "DemoAction2",
			RequestMode: node.RequestModeSync,
		},
	}
}

func (d *DemoAction2) ProcessAction(ctx context.Context, nodeParam *node.NodeParam) (*action.ActionResult, error) {
	fmt.Printf("DemoAction: %v\n||actType=%v", nodeParam, d.GetActType())
	return &action.ActionResult{
		Success: true,
	}, nil
}
