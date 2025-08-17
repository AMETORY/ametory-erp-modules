package meta

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/meta/whatsapp_api"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type MetaService struct {
	db                 *gorm.DB
	ctx                *context.ERPContext
	WhatsappApiService *whatsapp_api.WhatsAppAPIService
	storageProvider    string
	facebookBaseURL    string
	baseURL            string
}

func NewMetaService(db *gorm.DB, ctx *context.ERPContext, baseURL, facebookBaseURL string, storageProvider string) *MetaService {
	return &MetaService{
		db:                 db,
		ctx:                ctx,
		WhatsappApiService: whatsapp_api.NewWhatsAppAPIService(db, ctx, baseURL, facebookBaseURL, storageProvider),
		facebookBaseURL:    facebookBaseURL,
		storageProvider:    storageProvider,
		baseURL:            baseURL,
	}
}

func (c *MetaService) VerifyFacebook(req *http.Request, FacebookVerifyToken string) (string, error) {
	verifyToken := req.URL.Query().Get("hub.verify_token")
	challenge := req.URL.Query().Get("hub.challenge")

	utils.LogJson(req.URL.Query())
	fmt.Println("verifyToken", verifyToken)
	fmt.Println("FacebookVerifyToken", FacebookVerifyToken)

	if verifyToken == FacebookVerifyToken {
		return challenge, nil
	} else {
		return "", errors.New("invalid verify token")
	}
}
