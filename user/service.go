package user

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type UserService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

func NewUserService(erpContext *context.ERPContext) *UserService {
	return &UserService{erpContext: erpContext, db: erpContext.DB}
}

func (service *UserService) GetUserByID(userID string) (*models.UserModel, error) {
	user := &models.UserModel{}
	if err := service.db.Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUsers(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("users.full_name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	stmt = stmt.Model(&models.UserModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.UserModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.UserModel)
	newItems := make([]models.UserModel, 0)

	for _, v := range *items {
		file := models.FileModel{}
		s.db.Where("ref_id = ? and ref_type = ?", v.ID, "user").First(&file)
		if file.ID != "" {
			v.ProfilePicture = &file
		}
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
}

func (s *UserService) GetUserActivitiesByUserID(userID string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.UserActivityModel{}).Where("user_id = ?", userID)
	if request.URL.Query().Get("activity_type") != "" {
		stmt = stmt.Where("activity_type = ?", request.URL.Query().Get("activity_type"))
	}
	if request.URL.Query().Get("sort") != "" {
		stmt = stmt.Order(request.URL.Query().Get("sort"))
	} else {
		stmt = stmt.Order("created_at DESC")
	}
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.UserActivityModel{})

	items := page.Items.(*[]models.UserActivityModel)
	newItems := make([]models.UserActivityModel, 0)
	for _, v := range *items {
		files, _ := s.GetFilesByID(v.ID)
		v.Files = files
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
}

func (s *UserService) CreateActivity(userID string, activity *models.UserActivityModel) error {
	now := time.Now()
	if activity.StartedAt == nil {
		activity.StartedAt = &now
	}
	if err := s.db.Create(activity).Error; err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetFilesByID(ID string) ([]models.FileModel, error) {
	files := []models.FileModel{}
	if err := s.db.Where("ref_id = ? and ref_type = ?", ID, "user_activity").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (s *UserService) FinishActivityByUser(userID string, activityType models.UserActivityType, latitude *float64, longitude *float64, notes *string) (*models.UserActivityModel, error) {
	// Find the last activity with the given type
	activity := &models.UserActivityModel{}
	if err := s.db.Where("user_id = ? and activity_type = ? and finished_at is null", userID, activityType).Order("started_at DESC").First(activity).Error; err != nil {
		return nil, err
	}

	// Set the finished_at field to the current time
	now := time.Now()
	activity.FinishedAt = &now

	// Calculate and set the duration
	if activity.StartedAt != nil {
		duration := now.Sub(*activity.StartedAt)
		activity.Duration = &duration
		activity.FinishedLatitude = latitude
		activity.FinishedLongitude = longitude
		activity.FinishedNotes = notes
	}

	// Update the activity in the database
	if err := s.db.Save(activity).Error; err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *UserService) FinishActivityByID(activityID string) error {
	// Find the activity by ID
	activity := &models.UserActivityModel{}
	if err := s.db.Where("id = ?", activityID).First(activity).Error; err != nil {
		return err
	}

	// Set the finished_at field to the current time
	now := time.Now()
	activity.FinishedAt = &now

	// Calculate and set the duration
	if activity.StartedAt != nil {
		duration := now.Sub(*activity.StartedAt)
		activity.Duration = &duration
	}

	// Update the activity in the database
	if err := s.db.Save(activity).Error; err != nil {
		return err
	}

	return nil
}
