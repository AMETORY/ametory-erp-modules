package attendance_policy

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AttendancePolicyService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewAttendancePolicyService creates a new instance of AttendancePolicyService.
//
// The service is created by providing a pointer to an ERPContext. The ERPContext
// is used for authentication and authorization purposes.
func NewAttendancePolicyService(ctx *context.ERPContext) *AttendancePolicyService {
	return &AttendancePolicyService{db: ctx.DB, ctx: ctx}
}

// Migrate creates the attendance policy table in the database.
//
// The function takes a pointer to a GORM DB as parameter and uses it to create
// the attendance policy table in the database. If the table already exists, the
// function does nothing and returns nil. If an error occurs during the
// operation, the function returns the error object.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AttendancePolicy{},
	)
}

// Create creates a new attendance policy in the database.
//
// The function takes a pointer to an AttendancePolicy as parameter and uses it
// to create a new attendance policy in the database. If the operation is
// successful, the function returns nil; otherwise, it returns an error
// object indicating what went wrong.
func (s *AttendancePolicyService) Create(input *models.AttendancePolicy) error {
	return s.db.Create(input).Error
}

// FindOne retrieves an attendance policy by its ID from the database.
//
// The function takes an ID as a parameter and attempts to find the corresponding
// attendance policy. It preloads related entities such as WorkShift, Organization,
// Branch, and WorkLocation to ensure they are included in the result.
//
// Parameters:
// 	id (string): The ID of the attendance policy to retrieve.
//
// Returns:
// 	*models.AttendancePolicy: The attendance policy model instance if found, or nil if not found.
// 	error: An error object if the operation fails, or nil if successful.

func (s *AttendancePolicyService) FindOne(id string) (*models.AttendancePolicy, error) {
	var input models.AttendancePolicy
	err := s.db.Where("id = ?", id).Preload("WorkShift").Preload("Organization").Preload("Branch").Preload("WorkLocation").First(&input).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &input, nil
}

// FindAll retrieves a list of attendance policies from the database, paginated and filtered.
//
// The function takes an HTTP request as a parameter and uses it to determine the
// filter criteria and pagination parameters. It preloads related entities such as
// WorkShift, Organization, Branch, and WorkLocation to ensure they are included
// in the result.
//
// Parameters:
//
//	request (*http.Request): The HTTP request object containing filter criteria and pagination parameters.
//
// Returns:
//
//	paginate.Page: A Paginate.Page object containing the list of attendance policies and pagination metadata.
//	error: An error object if the operation fails, or nil if successful.
func (a *AttendancePolicyService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.AttendancePolicy{}).Preload("WorkShift").Preload("Organization").Preload("Branch").Preload("WorkLocation")
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AttendancePolicy{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateAttedancePolicyByWorkLocation updates the attendance policies associated with the given work location with the
// work location's latitude and longitude.
//
// The function takes a pointer to a WorkLocationModel as input and uses it to update the attendance policies in the
// database. If the update operation fails, an error is returned.
func (a *AttendancePolicyService) UpdateAttedancePolicyByWorkLocation(workLocation *models.WorkLocationModel) error {
	if workLocation == nil {
		return errors.New("work location is null")
	}
	return a.db.Model(&models.AttendancePolicy{}).Where("work_location_id = ?", workLocation.ID).Updates(map[string]any{
		"lat": workLocation.Latitude,
		"lng": workLocation.Longitude,
	}).Error

}

// Update modifies an existing attendance policy in the database.
//
// The function takes a pointer to an AttendancePolicy as a parameter and saves the changes
// to the database. If the update operation is successful, it returns nil; otherwise, it returns an error.
//
// Parameters:
// 	input (*models.AttendancePolicy): The attendance policy model instance to be updated.
//
// Returns:
// 	error: An error object if the update operation fails, or nil if successful.

func (s *AttendancePolicyService) Update(input *models.AttendancePolicy) error {
	return s.db.Save(input).Error
}

// Delete deletes an attendance policy from the database.
//
// The function takes the ID of the attendance policy as a parameter and
// attempts to delete it from the database. If the deletion is successful,
// the function returns nil; otherwise, it returns an error object indicating
// what went wrong.
//
// Parameters:
//
//	id (string): The ID of the attendance policy to be deleted.
//
// Returns:
//
//	error: An error object if the deletion fails, or nil if successful.
func (s *AttendancePolicyService) Delete(id string) error {
	return s.db.Delete(&models.AttendancePolicy{}, "id = ?", id).Error
}

// FindBestPolicy retrieves the most applicable attendance policy for the given company, branch, organization, and work shift.
//
// The function takes the IDs of the company, branch, organization, and work shift as parameters.
// It queries the database to find the most applicable attendance policy based on the given IDs.
// If a matching record is found, the function returns the attendance policy model instance; otherwise, it returns an error.
//
// The query is ordered by the following rules:
//  1. Work shift ID (most specific)
//  2. Organization ID
//  3. Branch ID
//  4. Company ID (least specific)
//  5. Default policy (if no other matching policy is found)
//
// Parameters:
//
//	companyID (string): The ID of the company.
//	branchID (*string): The ID of the branch. If empty, the query will ignore the branch ID.
//	orgID (*string): The ID of the organization. If empty, the query will ignore the organization ID.
//	shiftID (*string): The ID of the work shift. If empty, the query will ignore the work shift ID.
//
// Returns:
//
//	*models.AttendancePolicy: The attendance policy model instance if found, or nil if not found.
//	error: An error object if the query fails, or nil if successful.
func (s *AttendancePolicyService) FindBestPolicy(
	companyID string,
	branchID, orgID, shiftID *string,
) (*models.AttendancePolicy, error) {

	var policy models.AttendancePolicy

	query := s.db.Model(&models.AttendancePolicy{})

	query = query.Where(`
		company_id = ? AND
		(branch_id IS NULL OR branch_id = ?) AND
		(organization_id IS NULL OR organization_id = ?) AND
		(work_shift_id IS NULL OR work_shift_id = ?)`,
		companyID,
		utils.StringOrEmpty(branchID),
		utils.StringOrEmpty(orgID),
		utils.StringOrEmpty(shiftID),
	)

	query = query.Order(`
		CASE 
			WHEN work_shift_id IS NOT NULL THEN 1
			WHEN organization_id IS NOT NULL THEN 2
			WHEN branch_id IS NOT NULL THEN 3
			WHEN company_id IS NOT NULL THEN 4
			ELSE 5
		END ASC
	`)

	err := query.First(&policy).Error
	if err != nil {
		return nil, err
	}

	return &policy, nil
}
