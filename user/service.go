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

// NewUserService creates a new instance of UserService.
//
// It takes an ERPContext as an argument.
//
// It returns a pointer to a UserService.
func NewUserService(erpContext *context.ERPContext) *UserService {
	return &UserService{erpContext: erpContext, db: erpContext.DB}
}

// GetUserByID returns a user by their ID
//
// It takes the user ID as an argument.
//
// It returns the UserModel if found, otherwise an error.
func (service *UserService) GetUserByID(userID string) (*models.UserModel, error) {
	user := &models.UserModel{}
	if err := service.db.Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByPhone returns a user by their phone number
//
// It takes the phone number as an argument.
//
// It returns the UserModel if found, otherwise an error.
func (service *UserService) GetUserByPhone(phoneNumber string) (*models.UserModel, error) {
	user := &models.UserModel{}
	if err := service.db.Where("phone_number = ?", phoneNumber).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByCode returns a user by their code
//
// It takes the code as an argument.
//
// It returns the UserModel if found, otherwise an error.
func (service *UserService) GetUserByCode(code string) (*models.UserModel, error) {
	user := &models.UserModel{}
	if err := service.db.Where("code = ?", code).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetCompanyUsers retrieves a paginated list of users associated with a specific company.
//
// It takes a company ID and an HTTP request as input. The method uses GORM to query the database for users linked to the specified company ID. It also applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of UserModel and an error if the operation fails.
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

// GetUsers retrieves a paginated list of users from the database.
//
// It takes an HTTP request and a search query string as input. The method uses
// GORM to query the database for users, applying the search query to the full
// name, email, and phone number fields. The function utilizes pagination to
// manage the result set and applies any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of UserModel and an error if the
// operation fails.
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
		s.db.Order("created_at DESC").Where("ref_id = ? and ref_type = ?", v.ID, "user").First(&file)
		if file.ID != "" {
			v.ProfilePicture = &file
		}
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
}

// GetUserActivitiesByUserID retrieves a paginated list of user activities associated with a user ID.
//
// It takes a user ID and an HTTP request as input. The method uses GORM to query the database for user activities linked to the specified user ID. If the request contains an activity type, the method also filters the result by the activity type. If the request contains a company ID header, the method also filters the result by the company ID. The function utilizes pagination to manage the result set and applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of UserActivityModel and an error if the operation fails.
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

// GetLastClockinByUser retrieves the last user activity of type clockin associated with a user ID.
//
// It takes a user ID, a company ID, and a threshold duration as input. The method uses GORM to query the database for the last user activity of type clockin linked to the specified user ID and company ID, with a started_at timestamp greater than or equal to the current time minus the threshold duration.
//
// The function returns the UserActivityModel if found, otherwise an error.
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

// CreateActivity creates a user activity.
//
// It takes a user ID and an activity as input. The method uses GORM to create the user activity in the database.
//
// The function returns an error if the operation fails.
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

// GetFilesByID retrieves a list of files associated with a given ID and reference type.
//
// It takes a reference ID and a reference type as input. The method uses GORM to query the database for files linked to the specified reference ID and reference type.
//
// The function returns a list of FileModel and an error if the operation fails.
func (s *UserService) GetFilesByID(ID string) ([]models.FileModel, error) {
	files := []models.FileModel{}
	if err := s.db.Where("ref_id = ? and ref_type in (?)", ID, []string{"user_activity", "clock_in", "clock_out", "check_point", "check_point_finish"}).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// FinishActivityByUser finishes the last user activity of the given user ID with the given type.
//
// It takes a user ID and an activity type as input. The method uses GORM to query the database for the last user activity with the given type and user ID. It then sets the finished_at field of the activity to the current time and calculates the duration of the activity. Finally, it updates the activity in the database.
//
// The function returns the updated activity and an error if the operation fails.
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

// FinishActivityByActivityID finishes the user activity with the given activity ID.
//
// It takes a user ID and an activity ID as input. The method uses GORM to query the database for the user activity with the given ID and user ID. It then sets the finished_at field of the activity to the current time and calculates the duration of the activity. Finally, it updates the activity in the database.
//
// The function returns the updated activity and an error if the operation fails.
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

// FinishActivityByID finishes the user activity with the given ID.
//
// It takes an activity ID as input. The method uses GORM to query the database for the user activity with the given ID. It then sets the finished_at field of the activity to the current time and calculates the duration of the activity. Finally, it updates the activity in the database.
//
// The function returns an error if the operation fails.
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
