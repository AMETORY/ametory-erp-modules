package meta

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/meta/whatsapp_api"
	"gorm.io/gorm"
)

type MetaService struct {
	db                 *gorm.DB
	ctx                *context.ERPContext
	WhatsappApiService *whatsapp_api.WhatsAppAPIService
	storageProvider    string
	facebookBaseURL    string
}

func NewMetaService(db *gorm.DB, ctx *context.ERPContext, facebookBaseURL string, storageProvider string) *MetaService {
	return &MetaService{
		db:                 db,
		ctx:                ctx,
		WhatsappApiService: whatsapp_api.NewWhatsAppAPIService(db, ctx, facebookBaseURL, storageProvider),
	}
}

func (c *MetaService) VerifyFacebook(req *http.Request, FacebookVerifyToken string) (string, error) {
	verifyToken := req.URL.Query().Get("hub.verify_token")
	challenge := req.URL.Query().Get("hub.challenge")

	if verifyToken == FacebookVerifyToken {
		return challenge, nil
	} else {
		return "", errors.New("invalid verify token")
	}
}
