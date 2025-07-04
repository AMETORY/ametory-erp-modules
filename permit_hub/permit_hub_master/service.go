package permit_hub_master

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
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

func (s *MasterPermitHubService) UpdatePermitFieldDefinition(pfd *models.PermitFieldDefinition) error {
	return s.ctx.DB.Save(pfd).Error
}

func (s *MasterPermitHubService) DeletePermitFieldDefinition(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.PermitFieldDefinition{}).Error
}

func (s *MasterPermitHubService) CreatePermitType(pt *models.PermitType) error {
	return s.ctx.DB.Create(pt).Error
}

func (s *MasterPermitHubService) GetPermitTypeByID(id string) (*models.PermitType, error) {
	var pt models.PermitType
	if err := s.ctx.DB.Preload("FieldDefinitions").Preload("PermitApprovalFlow").Where("id = ?", id).First(&pt).Error; err != nil {
		return nil, err
	}
	return &pt, nil
}

func (s *MasterPermitHubService) GetPermitTypes(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Preload("FieldDefinitions").Preload("PermitApprovalFlow")
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

func (s *MasterPermitHubService) UpdatePermitType(pt *models.PermitType) error {
	return s.ctx.DB.Save(pt).Error
}

func (s *MasterPermitHubService) DeletePermitType(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.PermitType{}).Error
}

func (s *MasterPermitHubService) CreatePermitApprovalFlow(flow *models.PermitApprovalFlow) error {
	return s.ctx.DB.Create(flow).Error
}

func (s *MasterPermitHubService) GetPermitApprovalFlowByID(id string) (*models.PermitApprovalFlow, error) {
	var flow models.PermitApprovalFlow
	if err := s.ctx.DB.Preload("PermitType").Preload("Role").Where("id = ?", id).First(&flow).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

func (s *MasterPermitHubService) UpdatePermitApprovalFlow(flow *models.PermitApprovalFlow) error {
	return s.ctx.DB.Save(flow).Error
}

func (s *MasterPermitHubService) DeletePermitApprovalFlow(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.PermitApprovalFlow{}).Error
}

func (s *MasterPermitHubService) GetPermitApprovalFlows(request *http.Request, permitTypeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.
		Preload("PermitType", func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", permitTypeID)
		}).
		Preload("Role").
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
