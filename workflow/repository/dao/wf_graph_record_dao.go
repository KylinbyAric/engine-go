package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfGraphRecordDao struct {
	db *gorm.DB
}

var (
	wfGraphRecordDao     *WfGraphRecordDao
	wfGraphRecordDaoOnce sync.Once
)

// NewWfGraphRecordDao 返回进程内单例。首次调用决定底层 *gorm.DB。
// 需要事务隔离请用 WithTx(tx)。
func NewWfGraphRecordDao(db *gorm.DB) *WfGraphRecordDao {
	wfGraphRecordDaoOnce.Do(func() {
		if db == nil {
			db = models.DB()
		}
		wfGraphRecordDao = &WfGraphRecordDao{db: db}
	})
	return wfGraphRecordDao
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
	return d.GetOne(ctx, &WfGraphRecordWhere{GraphID: graphID, Version: &version})
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

// GetOne 按任意条件取一条。
func (d *WfGraphRecordDao) GetOne(ctx context.Context, where *WfGraphRecordWhere) (*models.WfGraphRecord, error) {
	tx := d.db.WithContext(ctx).Model(&models.WfGraphRecord{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return nil, errors.New("wf_graph_record: GetOne requires at least one condition")
	}
	var r models.WfGraphRecord
	if err := tx.First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// Update 用整 struct 更新（按主键），零值字段不写入。
func (d *WfGraphRecordDao) Update(ctx context.Context, r *models.WfGraphRecord) error {
	if r.ID == 0 {
		return errors.New("wf_graph_record: id is required for Update; use UpdateWhere for condition-based update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfGraphRecord{}).
		Where("id = ? AND is_delete = 0", r.ID).
		Updates(r).Error
}

// UpdateWhere 按任意条件更新；返回受影响行数。
func (d *WfGraphRecordDao) UpdateWhere(ctx context.Context, where *WfGraphRecordWhere, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	tx := d.db.WithContext(ctx).Model(&models.WfGraphRecord{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return 0, errors.New("wf_graph_record: UpdateWhere requires at least one condition")
	}
	res := tx.Updates(fields)
	return res.RowsAffected, res.Error
}

func (d *WfGraphRecordDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	_, err := d.UpdateWhere(ctx, &WfGraphRecordWhere{ID: &id}, fields)
	return err
}

// UpdateFieldsByGraphIDVersion 按 (graph_id, version) 改字段（与唯一业务键对应）。
func (d *WfGraphRecordDao) UpdateFieldsByGraphIDVersion(ctx context.Context, graphID string, version int, fields map[string]any) (int64, error) {
	return d.UpdateWhere(ctx, &WfGraphRecordWhere{GraphID: graphID, Version: &version}, fields)
}

// Delete 软删（按主键 ID），同时把状态改成 Deleted。
func (d *WfGraphRecordDao) Delete(ctx context.Context, id int64, updateBy string) error {
	_, err := d.UpdateWhere(ctx, &WfGraphRecordWhere{ID: &id}, map[string]any{
		"is_delete": 1,
		"status":    models.WfGraphStatusDeleted,
		"update_by": updateBy,
	})
	return err
}

// DeleteWhere 软删（按条件）。
func (d *WfGraphRecordDao) DeleteWhere(ctx context.Context, where *WfGraphRecordWhere, updateBy string) (int64, error) {
	return d.UpdateWhere(ctx, where, map[string]any{
		"is_delete": 1,
		"status":    models.WfGraphStatusDeleted,
		"update_by": updateBy,
	})
}

// WfGraphRecordWhere 通用条件结构。
type WfGraphRecordWhere struct {
	ID       *int64
	GraphID  string
	Version  *int
	Type     string
	Status   *models.WfGraphStatus
	CreateBy string
}

func (w *WfGraphRecordWhere) apply(tx *gorm.DB) (*gorm.DB, bool) {
	if w == nil {
		return tx, false
	}
	has := false
	if w.ID != nil {
		tx = tx.Where("id = ?", *w.ID)
		has = true
	}
	if w.GraphID != "" {
		tx = tx.Where("graph_id = ?", w.GraphID)
		has = true
	}
	if w.Version != nil {
		tx = tx.Where("version = ?", *w.Version)
		has = true
	}
	if w.Type != "" {
		tx = tx.Where("type = ?", w.Type)
		has = true
	}
	if w.Status != nil {
		tx = tx.Where("status = ?", *w.Status)
		has = true
	}
	if w.CreateBy != "" {
		tx = tx.Where("create_by = ?", w.CreateBy)
		has = true
	}
	return tx, has
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
