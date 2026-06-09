package models

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type WfGraphDao struct {
	db *gorm.DB
}

func NewWfGraphDao(db *gorm.DB) *WfGraphDao {
	if db == nil {
		db = DB()
	}
	return &WfGraphDao{db: db}
}

func (d *WfGraphDao) WithTx(tx *gorm.DB) *WfGraphDao {
	return &WfGraphDao{db: tx}
}

func (d *WfGraphDao) Create(ctx context.Context, g *WfGraph) error {
	return d.db.WithContext(ctx).Create(g).Error
}

func (d *WfGraphDao) GetByID(ctx context.Context, id int64) (*WfGraph, error) {
	var g WfGraph
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_delete = 0", id).
		First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (d *WfGraphDao) GetByGraphID(ctx context.Context, graphID string) (*WfGraph, error) {
	var g WfGraph
	err := d.db.WithContext(ctx).
		Where("graph_id = ? AND is_delete = 0", graphID).
		First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (d *WfGraphDao) Update(ctx context.Context, g *WfGraph) error {
	if g.ID == 0 {
		return errors.New("wf_graph: id is required for update")
	}
	return d.db.WithContext(ctx).
		Model(&WfGraph{}).
		Where("id = ? AND is_delete = 0", g.ID).
		Updates(g).Error
}

func (d *WfGraphDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).
		Model(&WfGraph{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(fields).Error
}

func (d *WfGraphDao) UpdateStatus(ctx context.Context, id int64, status WfGraphStatus, updateBy string) error {
	return d.UpdateFields(ctx, id, map[string]any{
		"status":    status,
		"update_by": updateBy,
	})
}

func (d *WfGraphDao) Delete(ctx context.Context, id int64, updateBy string) error {
	return d.db.WithContext(ctx).
		Model(&WfGraph{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(map[string]any{
			"is_delete": 1,
			"status":    WfGraphStatusDeleted,
			"update_by": updateBy,
		}).Error
}

type WfGraphQuery struct {
	GraphID  string
	Name     string
	Type     string
	Status   *WfGraphStatus
	RecordID *int64
	CreateBy string

	OrderBy string
	Offset  int
	Limit   int
}

func (d *WfGraphDao) List(ctx context.Context, q *WfGraphQuery) ([]*WfGraph, int64, error) {
	if q == nil {
		q = &WfGraphQuery{}
	}
	tx := d.db.WithContext(ctx).Model(&WfGraph{}).Where("is_delete = 0")
	if q.GraphID != "" {
		tx = tx.Where("graph_id = ?", q.GraphID)
	}
	if q.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.RecordID != nil {
		tx = tx.Where("record_id = ?", *q.RecordID)
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

	var list []*WfGraph
	if err := tx.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
