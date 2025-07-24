package activity

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type ActivityService struct {
	ctx *context.ERPContext
}

// NewActivityService creates a new instance of ActivityService.
//
// It takes an ERPContext as parameter and returns a pointer to an ActivityService.
//
// It uses the ERPContext to initialize the ActivityService.
func NewActivityService(ctx *context.ERPContext) *ActivityService {
	return &ActivityService{ctx: ctx}
}

// CreateActivity creates a new budget activity in the database.
//
// The function takes a BudgetActivityModel pointer as input and returns an error if
// the operation fails. The function uses GORM to create the budget activity data in
// the budget_activities table.
//
// If the creation is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *ActivityService) CreateActivity(activity *models.BudgetActivityModel) error {
	return service.ctx.DB.Create(activity).Error
}

// GetActivity retrieves a budget activity by its ID from the database.
//
// The function takes an activity ID as input and returns a pointer to a BudgetActivityModel
// and an error if the operation fails. It uses GORM to query the database for the budget
// activity with the given ID. If the activity is found, it returns the model; otherwise, it
// returns an error indicating what went wrong.
func (service *ActivityService) GetActivity(id string) (*models.BudgetActivityModel, error) {
	var activity models.BudgetActivityModel
	err := service.ctx.DB.Where("id = ?", id).First(&activity).Error
	return &activity, err
}

// UpdateActivity updates an existing budget activity in the database.
//
// The function takes a BudgetActivityModel pointer as input and returns an error if
// the operation fails. The function uses GORM to update the budget activity data in
// the budget_activities table.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *ActivityService) UpdateActivity(activity *models.BudgetActivityModel) error {
	return service.ctx.DB.Save(activity).Error
}

// DeleteActivity deletes a budget activity from the database by ID.
//
// The function takes an activity ID as input and returns an error if the deletion
// operation fails. The function uses GORM to delete the budget activity data from
// the budget_activities table.
//
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *ActivityService) DeleteActivity(id string) error {
	return service.ctx.DB.Delete(models.BudgetActivityModel{}, "id = ?", id).Error
}
func (service *ActivityService) CreateActivityDetail(activityID string, activityDetail *models.BudgetActivityDetailModel) error {
	activityDetail.BudgetActivityID = &activityID
	return service.ctx.DB.Create(activityDetail).Error
}

// GetActivityDetail retrieves a budget activity detail by its activity ID and detail ID.
//
// The function takes an activity ID and a detail ID as input and returns a pointer
// to a BudgetActivityDetailModel and an error if the operation fails. It uses GORM
// to query the database for the budget activity detail with the specified IDs. If
// the detail is found, it returns the model; otherwise, it returns an error indicating
// what went wrong.
func (service *ActivityService) GetActivityDetail(activityID, id string) (*models.BudgetActivityDetailModel, error) {
	var activityDetail models.BudgetActivityDetailModel
	err := service.ctx.DB.Where("budget_activity_id = ? AND id = ?", activityID, id).First(&activityDetail).Error
	return &activityDetail, err
}

// UpdateActivityDetail updates an existing budget activity detail in the database.
//
// The function takes an activity ID and a BudgetActivityDetailModel pointer as input.
// It associates the activity detail with the given activity ID and uses GORM to save
// the updated details in the database. It returns an error if the update operation fails.
// If the update is successful, the error is nil.
func (service *ActivityService) UpdateActivityDetail(activityID string, activityDetail *models.BudgetActivityDetailModel) error {
	activityDetail.BudgetActivityID = &activityID
	return service.ctx.DB.Save(activityDetail).Error
}

// DeleteActivityDetail deletes a budget activity detail from the database by its
// activity ID and detail ID.
//
// The function takes an activity ID and a detail ID as input and returns an error if
// the deletion operation fails. The function uses GORM to delete the budget activity
// detail data from the budget_activity_details table.
//
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *ActivityService) DeleteActivityDetail(activityID, id string) error {
	return service.ctx.DB.Where("budget_activity_id = ? AND id = ?", activityID, id).Delete(models.BudgetActivityDetailModel{}).Error
}
