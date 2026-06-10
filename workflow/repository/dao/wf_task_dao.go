package dao

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfTaskDao struct {
	db *gorm.DB
}

var (
	wfTaskDao     *WfTaskDao
	wfTaskDaoOnce sync.Once
)

// NewWfTaskDao 返回进程内单例。首次调用决定底层 *gorm.DB。
// 需要事务隔离请用 WithTx(tx)。
func NewWfTaskDao(db *gorm.DB) *WfTaskDao {
	wfTaskDaoOnce.Do(func() {
		if db == nil {
			db = models.DB()
		}
		wfTaskDao = &WfTaskDao{db: db}
	})
	return wfTaskDao
}

func (d *WfTaskDao) WithTx(tx *gorm.DB) *WfTaskDao {
	return &WfTaskDao{db: tx}
}

func (d *WfTaskDao) Create(ctx context.Context, t *models.WfTask) error {
	return d.db.WithContext(ctx).Create(t).Error
}

func (d *WfTaskDao) GetByID(ctx context.Context, id int64) (*models.WfTask, error) {
	var t models.WfTask
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_delete = 0", id).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *WfTaskDao) GetByUniqueKey(ctx context.Context, uid int64, graphID, instanceID, nodeID string) (*models.WfTask, error) {
	return d.GetOne(ctx, &WfTaskWhere{
		UID:        &uid,
		GraphID:    graphID,
		InstanceID: instanceID,
		NodeID:     nodeID,
	})
}

func (d *WfTaskDao) ListByInstanceID(ctx context.Context, instanceID string) ([]*models.WfTask, error) {
	var list []*models.WfTask
	err := d.db.WithContext(ctx).
		Where("instance_id = ? AND is_delete = 0", instanceID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// GetOne 按任意条件取一条。
func (d *WfTaskDao) GetOne(ctx context.Context, where *WfTaskWhere) (*models.WfTask, error) {
	tx := d.db.WithContext(ctx).Model(&models.WfTask{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return nil, errors.New("wf_task: GetOne requires at least one condition")
	}
	var t models.WfTask
	if err := tx.First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// Update 用整 struct 更新（按主键），零值字段不写入。
func (d *WfTaskDao) Update(ctx context.Context, t *models.WfTask) error {
	if t.ID == 0 {
		return errors.New("wf_task: id is required for Update; use UpdateWhere for condition-based update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", t.ID).
		Updates(t).Error
}

// UpdateWhere 按任意条件更新指定列，返回受影响行数。
func (d *WfTaskDao) UpdateWhere(ctx context.Context, where *WfTaskWhere, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	tx := d.db.WithContext(ctx).Model(&models.WfTask{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return 0, errors.New("wf_task: UpdateWhere requires at least one condition")
	}
	res := tx.Updates(fields)
	return res.RowsAffected, res.Error
}

// UpdateFields 按主键 ID 更新指定列。
func (d *WfTaskDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	_, err := d.UpdateWhere(ctx, &WfTaskWhere{ID: &id}, fields)
	return err
}

// UpdateFieldsByInstanceID 按 instance_id 批量更新该实例下的任务。
func (d *WfTaskDao) UpdateFieldsByInstanceID(ctx context.Context, instanceID string, fields map[string]any) (int64, error) {
	return d.UpdateWhere(ctx, &WfTaskWhere{InstanceID: instanceID}, fields)
}

// UpdateStatus 推进任务状态；终态自动写 end_time。
// where 传入定位条件（id / 唯一键 / instance_id 任选）；返回受影响行数。
func (d *WfTaskDao) UpdateStatus(ctx context.Context, where *WfTaskWhere, status models.WfTaskStatus, msg string) (int64, error) {
	fields := map[string]any{
		"status":     status,
		"status_msg": msg,
	}
	if status == models.WfTaskStatusSuccess || status == models.WfTaskStatusFail {
		fields["end_time"] = time.Now()
	}
	return d.UpdateWhere(ctx, where, fields)
}

// IncrRunCount 原子自增执行次数（按 ID）。
func (d *WfTaskDao) IncrRunCount(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", id).
		UpdateColumn("run_count", gorm.Expr("run_count + 1")).Error
}

// Delete 软删（按主键 ID）。
func (d *WfTaskDao) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", id).
		Update("is_delete", 1).Error
}

// DeleteWhere 软删（按条件）；返回受影响行数。
func (d *WfTaskDao) DeleteWhere(ctx context.Context, where *WfTaskWhere) (int64, error) {
	return d.UpdateWhere(ctx, where, map[string]any{"is_delete": 1})
}

// WfTaskWhere 通用条件结构：被 UpdateWhere/DeleteWhere/GetOne 共用。
type WfTaskWhere struct {
	ID         *int64
	UID        *int64
	GraphID    string
	InstanceID string
	NodeID     string
	NodeType   string
	Status     *models.WfTaskStatus
}

func (w *WfTaskWhere) apply(tx *gorm.DB) (*gorm.DB, bool) {
	if w == nil {
		return tx, false
	}
	has := false
	if w.ID != nil {
		tx = tx.Where("id = ?", *w.ID)
		has = true
	}
	if w.UID != nil {
		tx = tx.Where("uid = ?", *w.UID)
		has = true
	}
	if w.GraphID != "" {
		tx = tx.Where("graph_id = ?", w.GraphID)
		has = true
	}
	if w.InstanceID != "" {
		tx = tx.Where("instance_id = ?", w.InstanceID)
		has = true
	}
	if w.NodeID != "" {
		tx = tx.Where("node_id = ?", w.NodeID)
		has = true
	}
	if w.NodeType != "" {
		tx = tx.Where("node_type = ?", w.NodeType)
		has = true
	}
	if w.Status != nil {
		tx = tx.Where("status = ?", *w.Status)
		has = true
	}
	return tx, has
}

type WfTaskQuery struct {
	UID        *int64
	GraphID    string
	InstanceID string
	NodeType   string
	Status     *models.WfTaskStatus
	OrderBy    string
	Offset     int
	Limit      int
}

func (d *WfTaskDao) List(ctx context.Context, q *WfTaskQuery) ([]*models.WfTask, int64, error) {
	if q == nil {
		q = &WfTaskQuery{}
	}
	tx := d.db.WithContext(ctx).Model(&models.WfTask{}).Where("is_delete = 0")
	if q.UID != nil {
		tx = tx.Where("uid = ?", *q.UID)
	}
	if q.GraphID != "" {
		tx = tx.Where("graph_id = ?", q.GraphID)
	}
	if q.InstanceID != "" {
		tx = tx.Where("instance_id = ?", q.InstanceID)
	}
	if q.NodeType != "" {
		tx = tx.Where("node_type = ?", q.NodeType)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := q.OrderBy
	if orderBy == "" {
		orderBy = "id DESC"
	}
	tx = tx.Order(orderBy)
	if q.Limit > 0 {
		tx = tx.Limit(q.Limit)
	}
	if q.Offset > 0 {
		tx = tx.Offset(q.Offset)
	}

	var list []*models.WfTask
	if err := tx.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
