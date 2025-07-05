package permit_hub_master

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

type MasterPermitHubService struct {
	ctx *context.ERPContext
}

func NewMasterPermitHubService(ctx *context.ERPContext) *MasterPermitHubService {
	service := MasterPermitHubService{
		ctx: ctx,
	}

	return &service
}

func (s *MasterPermitHubService) CreatePermitFieldDefinition(pfd *models.PermitFieldDefinition) error {
	return s.ctx.DB.Create(pfd).Error
}

func (s *MasterPermitHubService) GetPermitFieldDefinitionByID(id string) (*models.PermitFieldDefinition, error) {
	var pfd models.PermitFieldDefinition
	if err := s.ctx.DB.Where("id = ?", id).First(&pfd).Error; err != nil {
		return nil, err
	}
	return &pfd, nil
}

func (s *MasterPermitHubService) UpdatePermitFieldDefinition(id string, pfd *models.PermitFieldDefinition) error {
	return s.ctx.DB.Model(&models.PermitFieldDefinition{}).Where("id = ?", id).Save(pfd).Error
}

func (s *MasterPermitHubService) DeletePermitFieldDefinition(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.PermitFieldDefinition{}).Error
}

func (s *MasterPermitHubService) CreatePermitType(pt *models.PermitType) error {
	return s.ctx.DB.Create(pt).Error
}

func (s *MasterPermitHubService) GetPermitTypeByID(id string) (*models.PermitType, error) {
	var pt models.PermitType
	if err := s.ctx.DB.
		Preload("FieldDefinitions").
		Preload("ApprovalFlow.Roles").
		Preload("PermitRequirements").
		Where("id = ?", id).First(&pt).Error; err != nil {
		return nil, err
	}

	for i, v := range pt.PermitRequirements {
		typeReq := models.PermitTypeRequirement{}
		if err := s.ctx.DB.Where("permit_type_id = ? AND permit_requirement_id = ?", pt.ID, v.ID).First(&typeReq).Error; err != nil {
			return nil, err
		}
		v.IsMandatory = typeReq.IsMandatory
		pt.PermitRequirements[i] = v
	}
	return &pt, nil
}

func (s *MasterPermitHubService) GetPermitTypes(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Model(&models.PermitType{}).
		Preload("FieldDefinitions").
		Preload("ApprovalFlow").
		Preload("PermitRequirements")
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("updated_at DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PermitType{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *MasterPermitHubService) UpdatePermitType(id string, pt *models.PermitType) error {
	return s.ctx.DB.Model(&models.PermitType{}).Where("id = ?", id).Save(pt).Error
}

func (s *MasterPermitHubService) DeletePermitType(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.PermitType{}).Error
}

func (s *MasterPermitHubService) CreatePermitApprovalFlow(flow *models.PermitApprovalFlow) error {
	return s.ctx.DB.Create(flow).Error
}
func (s *MasterPermitHubService) GetLastFlowByPermitTypeID(permitTypeID string) *models.PermitApprovalFlow {
	var flow models.PermitApprovalFlow
	if err := s.ctx.DB.Where("permit_type_id = ?", permitTypeID).Order("step_order DESC").First(&flow).Error; err != nil {
		return nil
	}
	return &flow
}

func (s *MasterPermitHubService) GetPermitApprovalFlowByID(id string) (*models.PermitApprovalFlow, error) {
	var flow models.PermitApprovalFlow
	if err := s.ctx.DB.Preload("PermitType").Preload("Role").Where("id = ?", id).First(&flow).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

func (s *MasterPermitHubService) UpdatePermitApprovalFlow(id string, flow *models.PermitApprovalFlow) error {
	return s.ctx.DB.Model(&models.PermitApprovalFlow{}).Where("id = ?", id).Save(flow).Error
}

func (s *MasterPermitHubService) DeletePermitApprovalFlow(id string) error {
	return s.ctx.DB.Where("id = ?", id).Unscoped().Delete(&models.PermitApprovalFlow{}).Error
}

func (s *MasterPermitHubService) GetPermitApprovalFlows(request *http.Request, permitTypeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.
		Preload("SubDistrict").
		Model(&models.PermitApprovalFlow{})
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("order ASC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PermitApprovalFlow{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *MasterPermitHubService) CreatePermitRequirement(req *models.PermitRequirement) error {
	return s.ctx.DB.Create(req).Error
}

func (s *MasterPermitHubService) GetPermitRequirementByID(id string) (*models.PermitRequirement, error) {
	var req models.PermitRequirement
	if err := s.ctx.DB.Preload("PermitType").Where("id = ?", id).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *MasterPermitHubService) UpdatePermitRequirement(id string, req *models.PermitRequirement) error {
	return s.ctx.DB.Model(&models.PermitRequirement{}).Where("id = ?", id).Save(req).Error
}

func (s *MasterPermitHubService) DeletePermitRequirement(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.PermitRequirement{}).Error
}

func (s *MasterPermitHubService) GetPermitRequirements(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.
		Preload("SubDistrict").
		Model(&models.PermitRequirement{})
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("name ASC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PermitRequirement{})
	page.Page = page.Page + 1
	return page, nil
}
