package whatsapp

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WhatsappService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewWhatsappService(db *gorm.DB, ctx *context.ERPContext) *WhatsappService {
	return &WhatsappService{
		db:  db,
		ctx: ctx,
	}
}

func (ws *WhatsappService) CreateWhatsappMessage(whatsappMessage *models.WhatsappMessageModel) error {
	return ws.db.Create(whatsappMessage).Error
}

func (ws *WhatsappService) GetWhatsappMessage(id string) (*models.WhatsappMessageModel, error) {
	var whatsappMessage models.WhatsappMessageModel
	err := ws.db.Where("id = ?", id).First(&whatsappMessage).Error
	if err != nil {
		return nil, err
	}
	return &whatsappMessage, nil
}

func (ws *WhatsappService) UpdateWhatsappMessage(whatsappMessage *models.WhatsappMessageModel) error {
	return ws.db.Save(whatsappMessage).Error
}

func (ws *WhatsappService) DeleteWhatsappMessage(id string) error {
	return ws.db.Delete(&models.WhatsappMessageModel{}, "id = ?", id).Error
}

func (ws *WhatsappService) GetWhatsappMessages(request http.Request, search string, session string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.db.Model(&models.WhatsappMessageModel{})

	if search != "" {
		stmt = stmt.Where("sender ILIKE ? OR receiver ILIKE ? OR message ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if session == "" {
		return paginate.Page{}, errors.New("session is required")
	}

	stmt = stmt.Where("session = ?", session)

	// if request.Header.Get("ID-Company") != "" {
	// 	stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	// }

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageModel{})
	return page, nil

}
