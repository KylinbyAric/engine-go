package action

import (
	"context"
	"fmt"

	"github.com/engine-go/workflow/core/node"
	"github.com/engine-go/workflow/core/node/action"
)

func Init() {
	actions := []action.ActionNodeHandle{
		NewDemoAction(),
		NewDemoAction2(),
	}
	for _, processor := range actions {
		action.RegisterActionNode(processor)
	}

}

type DemoAction struct {
	action.ActionNode
}

func NewDemoAction() action.ActionNodeHandle {
	return &DemoAction{
		ActionNode: action.ActionNode{
			ActType:     "DemoAction",
			RequestMode: node.RequestModeSync,
		},
	}
}

func (d *DemoAction) ProcessAction(ctx context.Context, nodeParam *node.NodeParam) (*action.ActionResult, error) {
	fmt.Printf("DemoAction: %v\n||actType=%v", nodeParam, d.GetActType())
	return &action.ActionResult{
		Success: true,
	}, nil
}
