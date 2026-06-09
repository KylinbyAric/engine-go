package dao

import (
	"context"
	"errors"
	"time"

	"github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfTaskDao struct {
	db *gorm.DB
}

func NewWfTaskDao(db *gorm.DB) *WfTaskDao {
	if db == nil {
		db = models.DB()
	}
	return &WfTaskDao{db: db}
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
	var t models.WfTask
	err := d.db.WithContext(ctx).
		Where("uid = ? AND graph_id = ? AND instance_id = ? AND node_id = ? AND is_delete = 0",
			uid, graphID, instanceID, nodeID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *WfTaskDao) ListByInstanceID(ctx context.Context, instanceID string) ([]*models.WfTask, error) {
	var list []*models.WfTask
	err := d.db.WithContext(ctx).
		Where("instance_id = ? AND is_delete = 0", instanceID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (d *WfTaskDao) Update(ctx context.Context, t *models.WfTask) error {
	if t.ID == 0 {
		return errors.New("wf_task: id is required for update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", t.ID).
		Updates(t).Error
}

func (d *WfTaskDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(fields).Error
}

func (d *WfTaskDao) UpdateStatus(ctx context.Context, id int64, status models.WfTaskStatus, msg string) error {
	fields := map[string]any{
		"status":     status,
		"status_msg": msg,
	}
	if status == models.WfTaskStatusDone || status == models.WfTaskStatusStop {
		fields["end_time"] = time.Now()
	}
	return d.UpdateFields(ctx, id, fields)
}

func (d *WfTaskDao) IncrRunCount(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", id).
		UpdateColumn("run_count", gorm.Expr("run_count + 1")).Error
}

func (d *WfTaskDao) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).
		Model(&models.WfTask{}).
		Where("id = ? AND is_delete = 0", id).
		Update("is_delete", 1).Error
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
