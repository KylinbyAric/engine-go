package dao

import (
	"context"
	"errors"

	"github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfGraphRecordDao struct {
	db *gorm.DB
}

func NewWfGraphRecordDao(db *gorm.DB) *WfGraphRecordDao {
	if db == nil {
		db = models.DB()
	}
	return &WfGraphRecordDao{db: db}
}

func (d *WfGraphRecordDao) WithTx(tx *gorm.DB) *WfGraphRecordDao {
	return &WfGraphRecordDao{db: tx}
}

func (d *WfGraphRecordDao) Create(ctx context.Context, r *models.WfGraphRecord) error {
	return d.db.WithContext(ctx).Create(r).Error
}

func (d *WfGraphRecordDao) GetByID(ctx context.Context, id int64) (*models.WfGraphRecord, error) {
	var r models.WfGraphRecord
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_delete = 0", id).
		First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *WfGraphRecordDao) GetByGraphIDVersion(ctx context.Context, graphID string, version int) (*models.WfGraphRecord, error) {
	var r models.WfGraphRecord
	err := d.db.WithContext(ctx).
		Where("graph_id = ? AND version = ? AND is_delete = 0", graphID, version).
		First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *WfGraphRecordDao) ListByGraphID(ctx context.Context, graphID string, limit int) ([]*models.WfGraphRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	var list []*models.WfGraphRecord
	err := d.db.WithContext(ctx).
		Where("graph_id = ? AND is_delete = 0", graphID).
		Order("version DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (d *WfGraphRecordDao) Update(ctx context.Context, r *models.WfGraphRecord) error {
	if r.ID == 0 {
		return errors.New("wf_graph_record: id is required for update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfGraphRecord{}).
		Where("id = ? AND is_delete = 0", r.ID).
		Updates(r).Error
}

func (d *WfGraphRecordDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).
		Model(&models.WfGraphRecord{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(fields).Error
}

func (d *WfGraphRecordDao) Delete(ctx context.Context, id int64, updateBy string) error {
	return d.db.WithContext(ctx).
		Model(&models.WfGraphRecord{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(map[string]any{
			"is_delete": 1,
			"status":    models.WfGraphStatusDeleted,
			"update_by": updateBy,
		}).Error
}

type WfGraphRecordQuery struct {
	GraphID  string
	Type     string
	Status   *models.WfGraphStatus
	CreateBy string
	OrderBy  string
	Offset   int
	Limit    int
}

func (d *WfGraphRecordDao) List(ctx context.Context, q *WfGraphRecordQuery) ([]*models.WfGraphRecord, int64, error) {
	if q == nil {
		q = &WfGraphRecordQuery{}
	}
	tx := d.db.WithContext(ctx).Model(&models.WfGraphRecord{}).Where("is_delete = 0")
	if q.GraphID != "" {
		tx = tx.Where("graph_id = ?", q.GraphID)
	}
	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.CreateBy != "" {
		tx = tx.Where("create_by = ?", q.CreateBy)
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

	var list []*models.WfGraphRecord
	if err := tx.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
