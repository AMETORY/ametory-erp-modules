package app

import (
	"context"
	"log"
	"net/mail"

	erpCtx "github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/thirdparty"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/email_api"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/google"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/whatsmeow_client"
)

// WithEmailSender WithEmailSender
func WithEmailSender(smtpServer string, smtpPort int, smtpUsername, smtpPassword string, from mail.Address) AppContainerOption {
	return func(c *AppContainer) {
		c.EmailSender = thirdparty.NewSMTPSender(
			smtpServer,
			smtpPort,
			smtpUsername,
			smtpPassword,
			from,
		)
		log.Println("EmailSender initialized")
	}
}

// WithEmailAPIService WithEmailAPIService
func WithEmailAPIService(from, domain, apiKey string, sender email_api.EmailAPI) AppContainerOption {
	return func(c *AppContainer) {
		c.EmailAPIService = email_api.NewEmailApiService(
			from,
			domain,
			apiKey,
			sender,
		)
		log.Println("EmailAPIService initialized")
	}
}

// WithWatzapClient WithWatzapClient
func WithWatzapClient(apiKey, numberKey, mockNumber string, isMock bool, redisKey string) AppContainerOption {
	return func(c *AppContainer) {
		c.WatzapClient = thirdparty.NewWatzapClient(
			apiKey,
			numberKey,
			mockNumber,
			isMock,
			redisKey,
		)
		log.Println("WatzapClient initialized")
	}
}

// WithWhatsmeowService WithWhatsmeowService
func WithWhatsmeowService(baseURL, mockNumber string, isMock bool, redisKey string) AppContainerOption {
	return func(c *AppContainer) {
		c.WhatsmeowService = whatsmeow_client.NewWhatsmeowService(
			baseURL,
			mockNumber,
			isMock,
			redisKey,
		)
		log.Println("WhatsmeowService initialized")
	}
}

// WithFirestore WithFirestore
func WithFirestore(ctx context.Context, firebaseCredentialFile, bucket string) AppContainerOption {
	return func(c *AppContainer) {
		fireStore, err := thirdparty.NewFirebaseApp(
			ctx,
			firebaseCredentialFile,
			bucket,
		)

		if err != nil {
			panic("Failed to initialize Firestore: " + err.Error())
		}
		c.Firestore = fireStore
		log.Println("Firestore initialized")
	}
}

// WithFCMService WithFCMService
func WithFCMService(ctx *erpCtx.ERPContext, serverKey, credentialPath *string) AppContainerOption {
	return func(c *AppContainer) {
		c.FCMService = google.NewFCMService(
			ctx,
			serverKey,
			credentialPath,
		)
		log.Println("FCMService initialized")
	}
}

func WithAppService(appService any) AppContainerOption {
	return func(c *AppContainer) {
		c.AppService = appService
		log.Println("AppService initialized")
	}
}
