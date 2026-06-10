package dao

import (
	"context"
	"errors"
	"sync"

	"github.com/engine-go/workflow/common"
	models "github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfUserInstanceDao struct {
	db *gorm.DB
}

var (
	wfUserInstanceDao     *WfUserInstanceDao
	wfUserInstanceDaoOnce sync.Once
)

// NewWfUserInstanceDao 返回进程内单例。
// 首次调用决定底层 *gorm.DB：nil 时回退到 models.DB()。
// 需要事务隔离的场景请用 WithTx(tx)，会构造一个不共享的新实例。
func NewWfUserInstanceDao(db *gorm.DB) *WfUserInstanceDao {
	wfUserInstanceDaoOnce.Do(func() {
		if db == nil {
			db = models.DB()
		}
		wfUserInstanceDao = &WfUserInstanceDao{db: db}
	})
	return wfUserInstanceDao
}

func (d *WfUserInstanceDao) WithTx(tx *gorm.DB) *WfUserInstanceDao {
	return &WfUserInstanceDao{db: tx}
}

func (d *WfUserInstanceDao) Create(ctx context.Context, ui *models.WfUserInstance) error {
	return d.db.WithContext(ctx).Create(ui).Error
}

func (d *WfUserInstanceDao) GetByID(ctx context.Context, id int64) (*models.WfUserInstance, error) {
	var ui models.WfUserInstance
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_delete = 0", id).
		First(&ui).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ui, nil
}

func (d *WfUserInstanceDao) GetByInstanceID(ctx context.Context, instanceID string) (*models.WfUserInstance, error) {
	var ui models.WfUserInstance
	err := d.db.WithContext(ctx).
		Where("instance_id = ? AND is_delete = 0", instanceID).
		First(&ui).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ui, nil
}

// GetOne 按任意条件取一条（如 uid+instance_id）
func (d *WfUserInstanceDao) GetOne(ctx context.Context, where *WfUserInstanceWhere) (*models.WfUserInstance, error) {
	tx := d.db.WithContext(ctx).Model(&models.WfUserInstance{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return nil, errors.New("wf_user_instance: GetOne requires at least one condition")
	}
	var ui models.WfUserInstance
	if err := tx.First(&ui).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ui, nil
}

// Update 用整 struct 更新（按主键），零值字段不写入。
func (d *WfUserInstanceDao) Update(ctx context.Context, ui *models.WfUserInstance) error {
	if ui.ID == 0 {
		return errors.New("wf_user_instance: id is required for Update; use UpdateWhere for condition-based update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfUserInstance{}).
		Where("id = ? AND is_delete = 0", ui.ID).
		Updates(ui).Error
}

// UpdateWhere 按任意条件批量更新指定列；返回受影响行数。
// where 必须至少有 1 个非空条件，避免误改全表。
func (d *WfUserInstanceDao) UpdateWhere(ctx context.Context, where *WfUserInstanceWhere, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}
	tx := d.db.WithContext(ctx).Model(&models.WfUserInstance{}).Where("is_delete = 0")
	tx, has := where.apply(tx)
	if !has {
		return 0, errors.New("wf_user_instance: UpdateWhere requires at least one condition")
	}
	res := tx.Updates(fields)
	return res.RowsAffected, res.Error
}

// UpdateFields 按主键 ID 更新指定列（最常用，保留兼容）。
func (d *WfUserInstanceDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	_, err := d.UpdateWhere(ctx, &WfUserInstanceWhere{ID: &id}, fields)
	return err
}

func (d *WfUserInstanceDao) SaveUserWorkFlowData(ctx context.Context, uid int64, instanceID string, data any) error {
	_, err := d.UpdateWhere(ctx,
		&WfUserInstanceWhere{UID: &uid, InstanceID: instanceID},
		map[string]any{
			"data_info": common.StructToString(data),
		},
	)
	return err
}

// UpdateFieldsByInstanceID 按 instance_id 更新指定列。
func (d *WfUserInstanceDao) UpdateFieldsByInstanceID(ctx context.Context, instanceID string, fields map[string]any) error {
	_, err := d.UpdateWhere(ctx, &WfUserInstanceWhere{InstanceID: instanceID}, fields)
	return err
}

// UpdateStatus 按 (uid, instance_id) 推进实例状态；同时回填 status_msg。
func (d *WfUserInstanceDao) UpdateStatus(ctx context.Context, uid int64, instanceID string, status models.WfUserInstanceStatus) error {
	_, err := d.UpdateWhere(ctx,
		&WfUserInstanceWhere{UID: &uid, InstanceID: instanceID},
		map[string]any{
			"status":     status,
			"status_msg": status.String(),
		},
	)
	return err
}

// UpdateCurrStep 按 (uid, instance_id) 推进当前节点。
func (d *WfUserInstanceDao) UpdateCurrStep(ctx context.Context, uid int64, instanceID, currStepID, updateBy string) error {
	_, err := d.UpdateWhere(ctx,
		&WfUserInstanceWhere{UID: &uid, InstanceID: instanceID},
		map[string]any{
			"curr_step_id": currStepID,
			"update_by":    updateBy,
		},
	)
	return err
}

// Delete 软删（按主键 ID）。
func (d *WfUserInstanceDao) Delete(ctx context.Context, id int64, updateBy string) error {
	return d.db.WithContext(ctx).
		Model(&models.WfUserInstance{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(map[string]any{
			"is_delete": 1,
			"update_by": updateBy,
		}).Error
}

// DeleteWhere 软删（按条件）；返回受影响行数。
func (d *WfUserInstanceDao) DeleteWhere(ctx context.Context, where *WfUserInstanceWhere, updateBy string) (int64, error) {
	return d.UpdateWhere(ctx, where, map[string]any{
		"is_delete": 1,
		"update_by": updateBy,
	})
}

// WfUserInstanceWhere 通用条件结构：被 UpdateWhere/DeleteWhere/GetOne 共用。
// 任一字段为零值表示不参与过滤；至少需要一个非空才能用于写操作。
type WfUserInstanceWhere struct {
	ID         *int64
	UID        *int64
	InstanceID string
	Status     *models.WfUserInstanceStatus
	FormID     string
	CreateBy   string
}

func (w *WfUserInstanceWhere) apply(tx *gorm.DB) (*gorm.DB, bool) {
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
	if w.InstanceID != "" {
		tx = tx.Where("instance_id = ?", w.InstanceID)
		has = true
	}
	if w.Status != nil {
		tx = tx.Where("status = ?", *w.Status)
		has = true
	}
	if w.FormID != "" {
		tx = tx.Where("form_id = ?", w.FormID)
		has = true
	}
	if w.CreateBy != "" {
		tx = tx.Where("create_by = ?", w.CreateBy)
		has = true
	}
	return tx, has
}

type WfUserInstanceQuery struct {
	UID      *int64
	Status   *models.WfUserInstanceStatus
	Name     string
	FormID   string
	CreateBy string
	OrderBy  string
	Offset   int
	Limit    int
}

func (d *WfUserInstanceDao) List(ctx context.Context, q *WfUserInstanceQuery) ([]*models.WfUserInstance, int64, error) {
	if q == nil {
		q = &WfUserInstanceQuery{}
	}
	tx := d.db.WithContext(ctx).Model(&models.WfUserInstance{}).Where("is_delete = 0")
	if q.UID != nil {
		tx = tx.Where("uid = ?", *q.UID)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.FormID != "" {
		tx = tx.Where("form_id = ?", q.FormID)
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

	var list []*models.WfUserInstance
	if err := tx.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
