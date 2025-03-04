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

func (ws *WhatsappService) GetWhatsappMessages(request http.Request, search string, JID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.db.Model(&models.WhatsappMessageModel{})

	if search != "" {
		stmt = stmt.Where("sender ILIKE ? OR receiver ILIKE ? OR message ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if JID == "" {
		return paginate.Page{}, errors.New("jid is required")
	}

	stmt = stmt.Where("j_id = ?", JID)
	stmt = stmt.Order("created_at DESC")

	if request.URL.Query().Get("session") != "" {
		stmt = stmt.Where("session = ?", request.URL.Query().Get("session"))
	}

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageModel{})
	return page, nil

}
func (ws *WhatsappService) GetWhatsappLastMessages(JID, session string) (models.WhatsappMessageModel, error) {
	var msg models.WhatsappMessageModel
	stmt := ws.db
	stmt = stmt.Order("created_at DESC").Where("j_id = ? and session = ?", JID, session)
	err := stmt.First(&msg).Error
	return msg, err

}

func (ws *WhatsappService) GetMessageSession(JID string) ([]models.WhatsappMessageModel, error) {

	var waGroup []struct {
		Session string `db:"session"`
	}
	err := ws.db.Model(&models.WhatsappMessageModel{}).Select("session").Where("j_id = ?", JID).Group("session").Find(&waGroup).Error
	if err != nil {
		return []models.WhatsappMessageModel{}, err
	}
	var waMsgs []models.WhatsappMessageModel = []models.WhatsappMessageModel{}
	for _, v := range waGroup {
		var waMsg models.WhatsappMessageModel
		ws.db.Where("j_id = ? AND session = ?", JID, v.Session).Order("created_at DESC").First(&waMsg)
		waMsgs = append(waMsgs, waMsg)
	}
	return waMsgs, nil
}
