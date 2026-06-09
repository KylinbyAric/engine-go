package dao

import (
	"context"
	"errors"

	models "github.com/engine-go/workflow/repository/models"
	"gorm.io/gorm"
)

type WfUserInstanceDao struct {
	db *gorm.DB
}

func NewWfUserInstanceDao(db *gorm.DB) *WfUserInstanceDao {
	if db == nil {
		db = models.DB()
	}
	return &WfUserInstanceDao{db: db}
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

func (d *WfUserInstanceDao) Update(ctx context.Context, ui *models.WfUserInstance) error {
	if ui.ID == 0 {
		return errors.New("wf_user_instance: id is required for update")
	}
	return d.db.WithContext(ctx).
		Model(&models.WfUserInstance{}).
		Where("id = ? AND is_delete = 0", ui.ID).
		Updates(ui).Error
}

func (d *WfUserInstanceDao) UpdateFields(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).
		Model(&models.WfUserInstance{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(fields).Error
}

func (d *WfUserInstanceDao) UpdateStatus(ctx context.Context, id int64, status models.WfUserInstanceStatus, msg, updateBy string) error {
	return d.UpdateFields(ctx, id, map[string]any{
		"status":     status,
		"status_msg": msg,
		"update_by":  updateBy,
	})
}

func (d *WfUserInstanceDao) UpdateCurrStep(ctx context.Context, id int64, currStepID, updateBy string) error {
	return d.UpdateFields(ctx, id, map[string]any{
		"curr_step_id": currStepID,
		"update_by":    updateBy,
	})
}

func (d *WfUserInstanceDao) Delete(ctx context.Context, id int64, updateBy string) error {
	return d.db.WithContext(ctx).
		Model(&models.WfUserInstance{}).
		Where("id = ? AND is_delete = 0", id).
		Updates(map[string]any{
			"is_delete": 1,
			"update_by": updateBy,
		}).Error
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
