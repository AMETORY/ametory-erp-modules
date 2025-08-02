package app

import (
	"context"
	"log"
	"net/mail"

	"github.com/AMETORY/ametory-erp-modules/app/flow_engine"
	"github.com/AMETORY/ametory-erp-modules/thirdparty"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/email_api"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/google"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/kafka"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/redis"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/websocket"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/whatsmeow_client"
)

// WithEmailSender initializes the EmailSender with the given SMTP server,
// port, username, password, and from address.
//
// This option is used to send emails using the email package.
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

// WithEmailAPIService initializes the EmailAPIService with the given from address,
// domain, API key, and EmailAPI sender.
//
// This option is used to send emails using the EmailAPI package.
func WithEmailAPIService(from, domain, apiKey, apiSecret string, sender email_api.EmailAPI) AppContainerOption {
	return func(c *AppContainer) {
		c.EmailAPIService = email_api.NewEmailApiService(
			from,
			domain,
			apiKey,
			apiSecret,
			sender,
		)
		log.Println("EmailAPIService initialized")
	}
}

// WithWatzapClient initializes the WatzapClient with the given API key, number key, mock number,
// isMock boolean, and redis key.
//
// This option is used to send WhatsApp messages using the Watzap API.
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

// WithWhatsmeowService initializes the WhatsmeowService with the given base URL, mock number, isMock boolean,
// and redis key.
//
// This option is used to send WhatsApp messages using the Whatsmeow API.
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

// WithFirestore initializes the Firestore service with the given context, Firebase credential file, and bucket name.
//
// This option is used to interact with the Firestore database.
//
// It panics if the Firestore service cannot be initialized.
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
		c.erpContext.Firestore = fireStore
		log.Println("Firestore initialized")
	}
}

// WithFCMService initializes the FCMService with the given context and credential path.
//
// This option is used to send notifications using Firebase Cloud Messaging (FCM).
// It creates an instance of FCMServiceV2 with the provided context and optional credential path.
// If the credential path is specified, it attempts to initialize the FCM client with the credentials file.

func WithFCMService(ctx *context.Context, credentialPath *string) AppContainerOption {
	return func(c *AppContainer) {
		c.FCMService = google.NewFCMServiceV2(
			ctx,
			credentialPath,
		)
		log.Println("FCMService initialized")
	}
}

// WithRedisService initializes the RedisService with the given context, address, password, and database number.
//
// This option is used to interact with the Redis database.
//
// It creates an instance of RedisService with the provided context, address, password, and database number.
func WithRedisService(ctx context.Context, address, password string, db int) AppContainerOption {
	return func(c *AppContainer) {
		c.RedisService = redis.NewRedisService(ctx, address, password, db)
		log.Println("RedisService initialized")
	}
}

// WithWebsocketService initializes the WebsocketService with the given context.
//
// This option is used to interact with the Websocket service.
//
// It creates an instance of WebsocketService with the provided context.
func WithWebsocketService() AppContainerOption {
	return func(c *AppContainer) {
		c.WebsocketService = websocket.NewWebsocketService()
		log.Println("WebsocketService initialized")
	}
}

// WithAppService adds the given AppService to the AppContainer.
//
// It is an optional option.
func WithAppService(appService any) AppContainerOption {
	return func(c *AppContainer) {
		c.AppService = appService
		log.Println("AppService initialized")
	}
}

// WithGoogleAPIService initializes the GoogleAPIService with the given API key.
//
// This option is used to interact with the Google Places API.
//
// It creates an instance of GoogleAPIService with the provided context and API key.
func WithGoogleAPIService(apiKey string) AppContainerOption {
	return func(c *AppContainer) {
		c.GoogleAPIService = google.NewGoogleAPIService(c.erpContext, apiKey)
		log.Println("GoogleAPIService initialized")
	}
}

// WithGeminiService initializes the GeminiService with the given API key.
//
// This option is used to interact with the Gemini AI service.
//
// It creates an instance of GeminiService with the provided context and API key.
// If the API key is empty, it does nothing.
func WithGeminiService(apiKey string) AppContainerOption {
	return func(c *AppContainer) {
		if apiKey == "" {
			return
		}
		c.GeminiService = google.NewGeminiService(c.erpContext, apiKey)
		log.Println("GeminiService initialized")
	}
}

// WithFlowEngine initializes the FlowEngine with the given context.
//
// This option is used to interact with the Flow Engine service.
//
// It creates an instance of FlowEngine with the provided context.
func WithFlowEngine() AppContainerOption {
	return func(c *AppContainer) {

		c.FlowEngine = flow_engine.NewFlowEngine()
		log.Println("FlowEngine initialized")
	}
}

// WithKafkaService adds the KafkaService to the AppContainer with the provided context and server.
//
// This option initializes a new instance of KafkaService using the specified context and server,
// allowing interaction with a Kafka message broker.

func WithKafkaService(ctx context.Context, server *string) AppContainerOption {
	return func(c *AppContainer) {

		c.KafkaService = kafka.NewKafkaService(ctx, server)
		log.Println("KafkaService initialized")
	}
}
