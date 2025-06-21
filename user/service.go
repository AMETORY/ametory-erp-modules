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
func (service *UserService) GetUserByCode(code string) (*models.UserModel, error) {
	user := &models.UserModel{}
	if err := service.db.Where("code = ?", code).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (service *UserService) GetCompanyUsers(companyID string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.
		Joins("JOIN user_companies ON users.id = user_companies.user_model_id").
		Where("user_companies.company_model_id = ?", companyID)
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("users.full_name ILIKE ? OR users.email ILIKE ? OR users.phone_number ILIKE ?",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
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
		service.db.Where("ref_id = ? and ref_type = ?", v.ID, "user").First(&file)
		if file.ID != "" {
			v.ProfilePicture = &file
		}
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
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
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
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

func (s *UserService) GetLastClockinByUser(userID string, companyID string, thresholdDuration time.Duration) (*models.UserActivityModel, error) {
	activity := &models.UserActivityModel{}
	err := s.db.Where("user_id = ? AND activity_type = ? AND company_id = ? AND started_at >= ?", userID, models.UserActivityClockIn, companyID, time.Now().Add(-thresholdDuration)).
		Order("started_at DESC").
		First(activity).Error
	if err != nil {
		return nil, err
	}
	return activity, nil
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
	if err := s.db.Where("ref_id = ? and ref_type in (?)", ID, []string{"user_activity", "clock_in", "clock_out", "check_point", "check_point_finish"}).Find(&files).Error; err != nil {
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
func (s *UserService) FinishActivityByActivityID(userID string, activityID string, latitude *float64, longitude *float64, notes *string) (*models.UserActivityModel, error) {
	// Find the last activity with the given type
	activity := &models.UserActivityModel{}
	if err := s.db.Where("user_id = ? and id = ? and finished_at is null", userID, activityID).Order("started_at DESC").First(activity).Error; err != nil {
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
