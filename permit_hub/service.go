package permit_hub

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/permit_hub/permit_hub_master"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PermitHubService struct {
	ctx                    *context.ERPContext
	MasterPermitHubService *permit_hub_master.MasterPermitHubService
}

func NewPermitHubService(ctx *context.ERPContext) *PermitHubService {
	service := PermitHubService{
		ctx:                    ctx,
		MasterPermitHubService: permit_hub_master.NewMasterPermitHubService(ctx),
	}
	if !service.ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

func (s *PermitHubService) Migrate() error {
	return s.ctx.DB.AutoMigrate(
		&models.Citizen{},
		&models.PermitDynamicRequestData{},
		&models.PermitUploadedDocument{},
		&models.PermitRequest{},
		&models.FinalPermitDocument{},
		&models.PermitApprovalLog{},
		// MASTER DATA
		&models.PermitFieldDefinition{},
		&models.PermitType{},
		&models.PermitApprovalFlow{},
		&models.PermitApprovalDecision{},
		&models.Subdistrict{},
		&models.District{},
		&models.Province{},
	)
}

func (s *PermitHubService) GetPermitTypeBySlug(slug string) (*models.PermitType, error) {
	var permitType models.PermitType
	if err := s.ctx.DB.Preload("FieldDefinitions").Where("slug = ?", slug).First(&permitType).Error; err != nil {
		return nil, err
	}
	return &permitType, nil
}

func (s *PermitHubService) CreateCitizenIfNotExists(citizen *models.Citizen) error {
	if err := s.ctx.DB.Where("nik = ?", citizen.NIK).First(&citizen).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.ctx.DB.Create(citizen).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

// CreatePermitRequest initiates a new permit request for a given citizen and permit type.
// It validates dynamic request data against the required field definitions of the permit type.
func (s *PermitHubService) CreatePermitRequest(citizenID, permitTypeSlug string, dyn *models.PermitDynamicRequestData) (*models.PermitRequest, error) {
	// Retrieve permit type by slug
	permitType, err := s.GetPermitTypeBySlug(permitTypeSlug)
	if err != nil {
		return nil, err
	}
	permitTypeID := permitType.ID

	// Initialize dynamic data if not provided
	if dyn == nil {
		dyn = &models.PermitDynamicRequestData{}
	}

	// Create a new permit request
	req := &models.PermitRequest{
		PermitTypeID: permitTypeID,
		CitizenID:    citizenID,
		Status:       "SUBMITTED",
		SubmittedAt:  time.Now(),
	}
	if err := s.ctx.DB.Create(req).Error; err != nil {
		return nil, err
	}

	// Validate dynamic request data against required fields
	for _, field := range permitType.FieldDefinitions {
		dynData := map[string]any{}
		json.Unmarshal(*dyn.Data, &dynData)
		if field.IsRequired && dynData[field.FieldKey] == nil {
			return nil, errors.New("field " + field.FieldLabel + " is required")
		}
	}

	// Save dynamic request data
	if err := s.ctx.DB.Create(dyn).Error; err != nil {
		return nil, err
	}

	// Retrieve the first approval step
	var firstStep models.PermitApprovalFlow
	s.ctx.DB.Where("permit_type_id = ?", permitTypeID).Preload("Roles").Order("order ASC").First(&firstStep)

	// Set the initial approval step and role for the permit request
	req.CurrentStepRoles = firstStep.Roles
	req.CurrentStep = 0
	s.ctx.DB.Save(req)

	return req, nil
}

func (s *PermitHubService) ApprovePermitRequestStep(requestID string, approvedBy *models.UserModel, note string, approved bool) error {
	// 1. Check approved by role
	if approvedBy.Role == nil {
		return errors.New("unauthorized: user role not found")
	}
	// 2. Get request
	var request models.PermitRequest
	if err := s.ctx.DB.Preload("CurrentStepRoles").First(&request, "id = ?", requestID).Error; err != nil {
		return errors.New("permit request not found")
	}

	// 3. Get current approval step
	var currentStep models.PermitApprovalFlow
	if err := s.ctx.DB.Where("permit_type_id = ? AND step_order = ?", request.PermitTypeID, request.CurrentStep).Preload("Roles").First(&currentStep).Error; err != nil {
		return errors.New("approval step not found")
	}

	// 4. Validate role
	authorized := false
	var approvedByRole *models.RoleModel
	for _, role := range currentStep.Roles {
		if approvedBy.Role.ID == role.ID {
			approvedByRole = &role
			authorized = true
			break
		}
	}
	if !authorized {
		return errors.New("unauthorized: user role not allowed for this step")
	}

	// 5. Log approval
	log := models.PermitApprovalLog{
		PermitRequestID: requestID,
		Step:            approvedByRole.Name,
		StepRoleID:      &approvedByRole.ID,
		StepRole:        approvedByRole,
		Status:          "REJECTED",
		ApprovedBy:      &approvedBy.ID,
		ApprovedByUser:  approvedBy,
		ApprovedAt:      time.Now(),
		Note:            note,
	}
	if approved {
		log.Status = "APPROVED"
	}
	if err := s.ctx.DB.Create(&log).Error; err != nil {
		return err
	}

	// 6. If rejected, mark request as rejected
	if !approved {
		request.Status = "REJECTED"
		return s.ctx.DB.Save(&request).Error
	}

	decision := models.PermitApprovalDecision{
		PermitRequestID: requestID,
		StepOrder:       currentStep.StepOrder,
		ApprovedAt:      time.Now(),
		ApprovedBy:      &approvedBy.ID,
		ApprovedByUser:  approvedBy,
		Note:            note,
		Decision:        request.Status,
	}

	if err := s.ctx.DB.Create(&decision).Error; err != nil {
		return err
	}

	if currentStep.ApprovalMode == "ALL" {
		var decisions []models.PermitApprovalDecision
		s.ctx.DB.Where("permit_request_id = ? AND step_order = ? AND DECISION = ?", request.ID, currentStep.StepOrder, "APPROVED").Find(&decisions)

		approvedRoles := map[string]bool{}
		for _, d := range decisions {
			approvedRoles[*d.ApprovedByUser.RoleID] = true
		}

		allApproved := true
		for _, role := range currentStep.Roles {
			if !approvedRoles[role.ID] {
				allApproved = false
				break
			}
		}

		if !allApproved {
			return nil // wait for others
		}

	}

	// 7. Check for next step
	var nextStep models.PermitApprovalFlow
	err := s.ctx.DB.Where("permit_type_id = ? AND `order` > ?", request.PermitTypeID, currentStep.StepOrder).
		Order("order ASC").First(&nextStep).Error

	if err == nil {
		// Masih ada step berikutnya
		request.CurrentStepRoles = nextStep.Roles
		request.CurrentStep = nextStep.StepOrder
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Tidak ada step lagi, final approval
		now := time.Now()
		request.Status = "APPROVED"
		request.ApprovedAt = &now
	} else {
		return err
	}

	return s.ctx.DB.Save(&request).Error
}

func (s *PermitHubService) GetAllRequests(request *http.Request) (paginate.Page, error) {

	pg := paginate.New()
	stmt := s.ctx.DB.
		Preload("PermitType", func(db *gorm.DB) *gorm.DB {
			return db.Preload("FieldDefinitions").Preload("ApprovalFlow")
		}).
		Preload("Citizen").
		Preload("RoleModel").
		Preload("PermitApprovalLog.StepRole").
		Preload("PermitUploadedDocument").
		Model(&models.PermitRequest{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("citized_ids") != "" {
		stmt = stmt.Where("citizen_id IN (?)", strings.Split(request.URL.Query().Get("citized_ids"), ","))
	}
	if request.URL.Query().Get("citizen_id") != "" {
		stmt = stmt.Where("citizen_id IN (?)", strings.Split(request.URL.Query().Get("citizen_id"), ","))
	}
	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("clock_in >= ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("clock_in <= ?", request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("created_at desc")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PermitRequest{})
	page.Page = page.Page + 1
	return page, nil

}

func (s *PermitHubService) GetRequestByID(requestID string) (*models.PermitRequest, error) {
	var request models.PermitRequest
	err := s.ctx.DB.
		Preload("PermitType", func(db *gorm.DB) *gorm.DB {
			return db.Preload("FieldDefinitions").Preload("ApprovalFlow")
		}).
		Preload("Citizen").
		Preload("RoleModel").
		Preload("PermitApprovalLog.StepRole").
		Preload("PermitUploadedDocument").
		Where("id = ?", requestID).
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (s *PermitHubService) UpdateRequest(requestID string, data *models.PermitRequest) error {
	return s.ctx.DB.Where("id = ?", requestID).Updates(data).Error
}

func (s *PermitHubService) DeleteRequest(requestID string) error {
	return s.ctx.DB.Where("id = ?", requestID).Delete(&models.PermitRequest{}).Error
}
