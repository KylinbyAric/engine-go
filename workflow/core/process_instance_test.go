package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainaction "github.com/engine-go/domain/workflow/node/action"
	"github.com/engine-go/workflow/repository/dao"
	"github.com/engine-go/workflow/repository/models"
)

// 一条最小可执行图：in → action(DemoAction)。
// 注意：图遍历用的是节点 nextnodes 字段，所以连接关系写在 nextnodes 而不是 successors。
const minimalExecutableGraph = `{
  "nodes": [
    {"node_id":"start","name":"开始","type":"in","nextnodes":["act"]},
    {"node_id":"act","name":"动作","type":"action","action_type":"DemoAction"}
  ]
}`

func TestMain(m *testing.M) {
	if err := chdirToProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "[test setup] %v\n", err)
		os.Exit(1)
	}
	if err := models.Init(); err != nil {
		// DB 不可达就跳过所有 case（CI 没本地 mysql 时不阻塞）
		fmt.Fprintf(os.Stderr, "[test setup] skip: models.Init failed: %v\n", err)
		os.Exit(0)
	}
	domainaction.Init()
	os.Exit(m.Run())
}

// chdirToProjectRoot 让 conf.LoadConfig 能用相对路径找到 conf/dev/app.toml
func chdirToProjectRoot() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return os.Chdir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return errors.New("project root with go.mod not found")
		}
		dir = parent
	}
}

// insertTestGraph 灌一条测试图；返回 cleanup 用 defer。
// 硬删（Unscoped）确保测试可重复跑；同时清掉同 graph_id 可能残留的 wf_user_instance。
func insertTestGraph(t *testing.T, graphID, graphType, graphJSON string) func() {
	t.Helper()
	ctx := context.Background()
	g := &models.WfGraph{
		GraphID:     graphID,
		Name:        "test-" + graphID,
		Description: "unit test",
		Graph:       graphJSON,
		Version:     1,
		Type:        graphType,
		RecordID:    0,
		Status:      models.WfGraphStatusActive,
	}
	g.CreateBy = "test"
	g.UpdateBy = "test"
	if err := dao.NewWfGraphDao(nil).Create(ctx, g); err != nil {
		t.Fatalf("insert test graph: %v", err)
	}
	return func() {
		db := models.DB().WithContext(ctx)
		db.Unscoped().Where("graph_id = ?", graphID).Delete(&models.WfGraph{})
		db.Unscoped().Where("graph_id = ?", graphID).Delete(&models.WfUserInstance{})
		db.Unscoped().Where("graph_id = ?", graphID).Delete(&models.WfTask{})
	}
}

func TestStartProc_NilParam(t *testing.T) {
	_, err := StartProc(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil param, got nil")
	}
	if !strings.Contains(err.Error(), "ExeParam") {
		t.Errorf("expected ExeParam error, got %v", err)
	}
}

func TestStartProc_NilExeParam(t *testing.T) {
	_, err := StartProc(context.Background(), &WorkFlowParam{})
	if err == nil {
		t.Fatal("expected error for nil ExeParam, got nil")
	}
}

func TestStartProc_GraphNotFound(t *testing.T) {
	_, err := StartProc(context.Background(), &WorkFlowParam{
		ExeParam: &ExeParam{
			UID:     1,
			GraphId: "test-not-exist-xxx",
			Type:    models.WfGraphTypeSvcPipe,
		},
	})
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
	if !strings.Contains(err.Error(), "不存在") {
		t.Errorf("expected '不存在' in error, got %v", err)
	}
}

func TestStartProc_TypeMismatch(t *testing.T) {
	cleanup := insertTestGraph(t, "test-type-mismatch", models.WfGraphTypeSvcPipe, minimalExecutableGraph)
	defer cleanup()

	_, err := StartProc(context.Background(), &WorkFlowParam{
		ExeParam: &ExeParam{
			UID:     1,
			GraphId: "test-type-mismatch",
			Type:    models.WfGraphTypeFlowPipe, // 故意传错类型
		},
	})
	if err == nil {
		t.Fatal("expected type mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "图类型不匹配") {
		t.Errorf("expected '图类型不匹配', got %v", err)
	}
}

func TestStartProc_SvcPipe(t *testing.T) {
	cleanup := insertTestGraph(t, "test-svc-pipe", models.WfGraphTypeSvcPipe, minimalExecutableGraph)
	defer cleanup()

	f, err := StartProc(context.Background(), &WorkFlowParam{
		ExeParam: &ExeParam{
			UID:     1001,
			GraphId: "test-svc-pipe",
			Type:    models.WfGraphTypeSvcPipe,
		},
		DataInfo: map[string]any{"hello": "world"},
	})
	if err != nil {
		t.Fatalf("StartProc returned error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil UserFlow")
	}
	// svc_pipe 不留痕：InstanceID 为空，wf_user_instance 表里查不到
	if f.InstanceID != "" {
		t.Errorf("svc_pipe should not have InstanceID, got %q", f.InstanceID)
	}
	count := int64(0)
	models.DB().Model(&models.WfUserInstance{}).
		Where("graph_id = ?", "test-svc-pipe").Count(&count)
	if count != 0 {
		t.Errorf("svc_pipe should not write wf_user_instance, found %d rows", count)
	}
}

func TestStartProc_FlowPipe(t *testing.T) {
	cleanup := insertTestGraph(t, "test-flow-pipe", models.WfGraphTypeFlowPipe, minimalExecutableGraph)
	defer cleanup()
	ctx := context.Background()

	f, err := StartProc(ctx, &WorkFlowParam{
		ExeParam: &ExeParam{
			UID:     2002,
			GraphId: "test-flow-pipe",
			Type:    models.WfGraphTypeFlowPipe,
		},
		DataInfo: map[string]any{"x": 1},
	})
	if err != nil {
		t.Fatalf("StartProc returned error: %v", err)
	}
	if f == nil || f.InstanceID == "" {
		t.Fatalf("expected non-empty InstanceID for flow_pipe, got %+v", f)
	}

	// 读回来验状态：应该是 Success，且 data_info 已存
	ui, err := dao.NewWfUserInstanceDao(nil).GetByInstanceID(ctx, f.InstanceID)
	if err != nil {
		t.Fatalf("GetByInstanceID: %v", err)
	}
	if ui == nil {
		t.Fatal("instance not persisted")
	}
	if ui.Status != models.WfUserInstanceStatusSuccess {
		t.Errorf("expected status=Success(%d), got %d (msg=%s)",
			models.WfUserInstanceStatusSuccess, ui.Status, ui.StatusMsg)
	}
	if ui.DataInfo == "" {
		t.Error("expected DataInfo to be persisted, got empty")
	}
	if !strings.Contains(ui.DataInfo, `"x"`) {
		t.Errorf("DataInfo should contain marshaled DataInfo, got %s", ui.DataInfo)
	}
}
