package main

import (
	"context"
	"fmt"
	"log"

	"github.com/engine-go/workflow/models"
)

func main() {
	dao := models.NewWfGraphDao(nil)
	_, total, err := dao.List(context.Background(), &models.WfGraphQuery{Limit: 1})
	if err != nil {
		log.Fatalf("list wf_graph: %v", err)
	}
	fmt.Printf("engine-go started, wf_graph rows=%d\n", total)
}
