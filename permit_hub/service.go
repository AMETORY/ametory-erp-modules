package permit_hub

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/permit_hub/citizen"
	"github.com/AMETORY/ametory-erp-modules/permit_hub/permit_hub_master"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PermitHubService struct {
	ctx                    *context.ERPContext
	MasterPermitHubService *permit_hub_master.MasterPermitHubService
	CitizenService         *citizen.CitizenService
}

func NewPermitHubService(ctx *context.ERPContext) *PermitHubService {
	service := PermitHubService{
		ctx:                    ctx,
		MasterPermitHubService: permit_hub_master.NewMasterPermitHubService(ctx),
		CitizenService:         citizen.NewCitizenService(ctx),
	}
	if !service.ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

func (s *PermitHubService) Migrate() error {
	return s.ctx.DB.AutoMigrate(
		&models.Citizen{},
		&models.PermitType{},
		&models.PermitRequirement{},
		&models.PermitApprovalFlow{},
		&models.PermitFieldDefinition{},
		&models.PermitRequest{},
		&models.PermitDynamicRequestData{},
		&models.PermitUploadedDocument{},
		&models.FinalPermitDocument{},
		&models.PermitApprovalLog{},
		&models.PermitApprovalDecision{},
		&models.PermitTypeRequirement{},
		&models.SubDistrict{},
		&models.District{},
		&models.City{},
		&models.Province{},
	)
}

func (s *PermitHubService) GetPermitTypeBySlug(slug string) (*models.PermitType, error) {
	var permitType models.PermitType
	if err := s.ctx.DB.Preload("FieldDefinitions").
		Preload("ApprovalFlow").
		Preload("PermitRequirements").
		Where("slug = ?", slug).First(&permitType).Error; err != nil {
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
func (s *PermitHubService) CreatePermitRequest(citizenID, subDistrictID, permitTypeSlug string, dyn *models.PermitDynamicRequestData, uploadedDocuments []models.PermitUploadedDocument) (*models.PermitRequest, error) {
	var permitType models.PermitType
	if err := s.ctx.DB.Preload("FieldDefinitions").
		Preload("ApprovalFlow").
		Preload("PermitRequirements").
		Where("slug = ?", permitTypeSlug).First(&permitType).Error; err != nil {
		return nil, err
	}
	// Retrieve permit type by slug

	permitTypeID := permitType.ID
	req := &models.PermitRequest{
		Code:          utils.RandString(8, true),
		PermitTypeID:  permitTypeID,
		CitizenID:     citizenID,
		Status:        "SUBMITTED",
		SubmittedAt:   time.Now(),
		SubDistrictID: &subDistrictID,
		Documents:     uploadedDocuments,
	}
	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {

		reqID := utils.Uuid()
		for _, v := range permitType.PermitRequirements {
			var typeRequirement models.PermitTypeRequirement
			if err := tx.Where("permit_type_id = ? AND permit_requirement_id = ?", permitType.ID, v.ID).First(&typeRequirement).Error; err != nil {
				return err
			}
			if typeRequirement.IsMandatory {
				var found bool
				for _, d := range uploadedDocuments {
					if d.PermitRequirementCode == nil {
						continue
					}
					if *d.PermitRequirementCode == v.Code {
						found = true
						break
					}
				}

				if !found {
					return errors.New("mandatory document : " + v.Name + " not uploaded")
				}

			}
		}

		// Initialize dynamic data if not provided
		if dyn == nil {
			dyn = &models.PermitDynamicRequestData{}
		}

		for i, v := range uploadedDocuments {
			v.PermitRequestID = &reqID
			uploadedDocuments[i] = v
		}
		// Create a new permit request

		req.ID = reqID
		if err := tx.Create(req).Error; err != nil {
			return err
		}

		// Validate dynamic request data against required fields
		for _, field := range permitType.FieldDefinitions {
			dynData := map[string]any{}
			json.Unmarshal(*dyn.Data, &dynData)
			if field.IsRequired && dynData[field.FieldKey] == nil {
				return errors.New("field " + field.FieldLabel + " is required")
			}

			if field.FieldType == models.CHECKBOX || field.FieldType == models.SELECT {
				var options []string
				err := json.Unmarshal(*field.Options, &options)
				if err != nil {
					return err
				}
				dataValue, ok := dynData[field.FieldKey].(string)
				if !ok {
					return errors.New("field " + field.FieldLabel + " no value")
				}

				if !slices.Contains(options, dataValue) {
					return errors.New("field " + field.FieldLabel + " invalid value")
				}
			}
		}

		dyn.PermitRequestID = req.ID

		// Save dynamic request data
		if err := tx.Create(dyn).Error; err != nil {
			return err
		}

		// Retrieve the first approval step
		var firstStep models.PermitApprovalFlow
		tx.Where("permit_type_id = ?", permitTypeID).Preload("Roles").Order(`"step_order" ASC`).First(&firstStep)
		if len(firstStep.Roles) == 0 {
			return errors.New("no approval roles found")
		}

		// Set the initial approval step and role for the permit request
		req.CurrentStepRoles = firstStep.Roles
		req.CurrentStep = 0
		tx.Save(req)
		return nil
	})
	if err != nil {
		return nil, err
	}

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
		s.ctx.DB.Model(&request).Association("CurrentStepRoles").Clear()
		request.CurrentStepRoles = nextStep.Roles
		request.CurrentStep = nextStep.StepOrder
		request.Status = "PROCESSING"
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
		Preload("CurrentStepRoles").
		// Preload("ApprovalLogs").
		// Preload("Documents").
		// Preload("SubDistrict.District.City.Province").
		// Preload("FinalPermitDocuments").
		Model(&models.PermitRequest{})

	if request.Header.Get("ID-SubDistrict") != "" {
		stmt = stmt.Where("sub_district_id = ?", request.Header.Get("ID-SubDistrict"))
	}
	if request.URL.Query().Get("citized_ids") != "" {
		stmt = stmt.Where("citizen_id IN (?)", strings.Split(request.URL.Query().Get("citized_ids"), ","))
	}
	if request.URL.Query().Get("citizen_id") != "" {
		stmt = stmt.Where("citizen_id IN (?)", strings.Split(request.URL.Query().Get("citizen_id"), ","))
	}
	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("submitted_at >= ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("submitted_at <= ?", request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("submitted_at desc")
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
			return db.Preload("FieldDefinitions").Preload("PermitRequirements").Preload("ApprovalFlow.Roles")
		}).
		Preload("Citizen").
		Preload("CurrentStepRoles").
		Preload("ApprovalLogs").
		Preload("Documents").
		Preload("SubDistrict.District.City.Province").
		Preload("FinalPermitDocuments").
		Preload("DynamicRequestData").
		Where("id = ?", requestID).
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (s *PermitHubService) GetPermitRequestByCode(code string) (*models.PermitRequest, error) {
	var request models.PermitRequest
	err := s.ctx.DB.
		Preload("PermitType", func(db *gorm.DB) *gorm.DB {
			return db.Preload("FieldDefinitions").Preload("PermitRequirements").Preload("ApprovalFlow.Roles")
		}).
		Preload("Citizen").
		Preload("CurrentStepRoles").
		Preload("ApprovalLogs").
		Preload("Documents").
		Preload("SubDistrict.District.City.Province").
		Preload("FinalPermitDocuments").
		Preload("DynamicRequestData").
		Where("code = ?", code).
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
