package activity

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type ActivityService struct {
	ctx *context.ERPContext
}

func NewActivityService(ctx *context.ERPContext) *ActivityService {
	return &ActivityService{ctx: ctx}
}

func (service *ActivityService) CreateActivity(activity *models.BudgetActivityModel) error {
	return service.ctx.DB.Create(activity).Error
}

func (service *ActivityService) GetActivity(id string) (*models.BudgetActivityModel, error) {
	var activity models.BudgetActivityModel
	err := service.ctx.DB.Where("id = ?", id).First(&activity).Error
	return &activity, err
}

func (service *ActivityService) UpdateActivity(activity *models.BudgetActivityModel) error {
	return service.ctx.DB.Save(activity).Error
}

func (service *ActivityService) DeleteActivity(id string) error {
	return service.ctx.DB.Delete(models.BudgetActivityModel{}, "id = ?", id).Error
}
func (service *ActivityService) CreateActivityDetail(activityID string, activityDetail *models.BudgetActivityDetailModel) error {
	activityDetail.BudgetActivityID = &activityID
	return service.ctx.DB.Create(activityDetail).Error
}

func (service *ActivityService) GetActivityDetail(activityID, id string) (*models.BudgetActivityDetailModel, error) {
	var activityDetail models.BudgetActivityDetailModel
	err := service.ctx.DB.Where("budget_activity_id = ? AND id = ?", activityID, id).First(&activityDetail).Error
	return &activityDetail, err
}

func (service *ActivityService) UpdateActivityDetail(activityID string, activityDetail *models.BudgetActivityDetailModel) error {
	activityDetail.BudgetActivityID = &activityID
	return service.ctx.DB.Save(activityDetail).Error
}

func (service *ActivityService) DeleteActivityDetail(activityID, id string) error {
	return service.ctx.DB.Where("budget_activity_id = ? AND id = ?", activityID, id).Delete(models.BudgetActivityDetailModel{}).Error
}
