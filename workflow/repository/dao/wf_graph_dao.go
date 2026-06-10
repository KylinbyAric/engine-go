package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfGraphDao struct {
	db *gorm.DB
}

var (
	wfGraphDao     *WfGraphDao
	wfGraphDaoOnce sync.Once
)

// NewWfGraphDao 返回进程内单例。首次调用决定底层 *gorm.DB。
// 需要事务隔离请用 WithTx(tx)。
func NewWfGraphDao(db *gorm.DB) *WfGraphDao {
	wfGraphDaoOnce.Do(func() {
		if db == nil {
			db = models.DB()
		}
		wfGraphDao = &WfGraphDao{db: db}
	})
	return wfGraphDao
}

func (d *WfGraphDao) WithTx(tx *gorm.DB) *WfGraphDao {
	return &WfGraphDao{db: tx}
}

func (d *WfGraphDao) Create(ctx context.Context, g *models.WfGraph) error {
	return d.db.WithContext(ctx).Create(g).Error
}

func (d *WfGraphDao) GetByID(ctx context.Context, id int64) (*models.WfGraph, error) {
	var g models.WfGraph
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

func (d *WfGraphDao) GetByGraphID(ctx context.Context, graphID string) (*models.WfGraph, error) {
	return d.GetOne(ctx, &WfGraphWhere{GraphID: graphID})
}

// GetOne 按任意条件取一条。
func (d *WfGraphDao) GetOne(ctx context.Context, where *WfGraphWhere) (*models.WfGraph, error) {
	tx := d.db.WithContext(ctx).Model(&models.WfGraph{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return nil, errors.New("wf_graph: GetOne requires at least one condition")
	}
	var g models.WfGraph
	if err := tx.First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// Update 用整 struct 更新（按主键），零值字段不写入。
func (d *WfGraphDao) Update(ctx context.Context, g *models.WfGraph) error {
	if g.ID == 0 {
		return errors.New("wf_graph: id is required for Update; use UpdateWhere for condition-based update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfGraph{}).
		Where("id = ? AND is_delete = 0", g.ID).
		Updates(g).Error
}

// UpdateWhere 按任意条件更新；返回受影响行数。
func (d *WfGraphDao) UpdateWhere(ctx context.Context, where *WfGraphWhere, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	tx := d.db.WithContext(ctx).Model(&models.WfGraph{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return 0, errors.New("wf_graph: UpdateWhere requires at least one condition")
	}
	res := tx.Updates(fields)
	return res.RowsAffected, res.Error
}

func (d *WfGraphDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	_, err := d.UpdateWhere(ctx, &WfGraphWhere{ID: &id}, fields)
	return err
}

// UpdateFieldsByGraphID 按 graph_id 更新。
func (d *WfGraphDao) UpdateFieldsByGraphID(ctx context.Context, graphID string, fields map[string]any) (int64, error) {
	return d.UpdateWhere(ctx, &WfGraphWhere{GraphID: graphID}, fields)
}

func (d *WfGraphDao) UpdateStatus(ctx context.Context, id int64, status models.WfGraphStatus, updateBy string) error {
	return d.UpdateFields(ctx, id, map[string]any{
		"status":    status,
		"update_by": updateBy,
	})
}

func (d *WfGraphDao) Delete(ctx context.Context, id int64, updateBy string) error {
	_, err := d.UpdateWhere(ctx, &WfGraphWhere{ID: &id}, map[string]any{
		"is_delete": 1,
		"status":    models.WfGraphStatusDeleted,
		"update_by": updateBy,
	})
	return err
}

// DeleteWhere 按条件软删（小心使用，会影响命中的全部行）。
func (d *WfGraphDao) DeleteWhere(ctx context.Context, where *WfGraphWhere, updateBy string) (int64, error) {
	return d.UpdateWhere(ctx, where, map[string]any{
		"is_delete": 1,
		"status":    models.WfGraphStatusDeleted,
		"update_by": updateBy,
	})
}

// WfGraphWhere 通用条件结构。
type WfGraphWhere struct {
	ID       *int64
	GraphID  string
	Name     string
	Type     string
	Status   *models.WfGraphStatus
	RecordID *int64
	CreateBy string
}

func (w *WfGraphWhere) apply(tx *gorm.DB) (*gorm.DB, bool) {
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
	if w.Name != "" {
		tx = tx.Where("name = ?", w.Name)
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
	if w.RecordID != nil {
		tx = tx.Where("record_id = ?", *w.RecordID)
		has = true
	}
	if w.CreateBy != "" {
		tx = tx.Where("create_by = ?", w.CreateBy)
		has = true
	}
	return tx, has
}

type WfGraphQuery struct {
	GraphID  string
	Name     string
	Type     string
	Status   *models.WfGraphStatus
	RecordID *int64
	CreateBy string

	OrderBy string
	Offset  int
	Limit   int
}

func (d *WfGraphDao) List(ctx context.Context, q *WfGraphQuery) ([]*models.WfGraph, int64, error) {
	if q == nil {
		q = &WfGraphQuery{}
	}
	tx := d.db.WithContext(ctx).Model(&models.WfGraph{}).Where("is_delete = 0")
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

	var list []*models.WfGraph
	if err := tx.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
