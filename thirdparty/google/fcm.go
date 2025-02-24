package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"

	"firebase.google.com/go/v4/messaging"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	fcm "github.com/appleboy/go-fcm"
)

type FCMService struct {
	serverKey *string
	ctx       *context.ERPContext
	client    *fcm.Client
}

type FCMMessage struct {
	To   string            `json:"to"`
	Data map[string]string `json:"data"`
}

func NewFCMService(ctx *context.ERPContext, serverKey, credentialPath *string) *FCMService {
	var client *fcm.Client
	// fmt.Println("LOAD CONFIG", *serverKey, *credentialPath)
	if credentialPath != nil {
		cl, err := fcm.NewClient(
			*ctx.Ctx,
			fcm.WithCredentialsFile(*credentialPath),
		)
		if err == nil {
			client = cl
		}
	}
	return &FCMService{serverKey: serverKey, ctx: ctx, client: client}
}

func (s *FCMService) SendFCMV2MessageByUserID(userID, title, body string, data map[string]string) error {
	var user = models.UserModel{}
	err := s.ctx.DB.Model(&user).Where("id = ? or email = ?", userID, userID).First(&user).Error
	if err != nil {
		log.Println("Error getting user", err)
		return err
	}

	var pushTokens []models.PushTokenModel
	err = s.ctx.DB.Model(&pushTokens).Where("user_id = ? AND type = ?", user.ID, "fcm").Find(&pushTokens).Error
	if err != nil {
		log.Println("Error getting push token", err)
		return err
	}

	for _, pushToken := range pushTokens {
		err = s.SendFCMV2Message(pushToken.Token, "HALLO", "INI CUMAN TEST", map[string]string{})
		if err != nil {
			log.Println("Error sending fcm message", err)
			return err
		}
	}

	log.Println("Success sending fcm message to", user.FullName)
	// var user = models.UserModel{}

	return nil
}
func (s *FCMService) SendFCMV2Message(token, title, body string, data map[string]string) error {
	if s.client == nil {
		return fmt.Errorf("client is not set")
	}
	data["title"] = title
	data["body"] = body
	utils.LogJson(data)
	resp, err := s.client.Send(
		*s.ctx.Ctx,
		&messaging.Message{
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Token: token,
			Data:  data,
		},
	)

	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Println("success count:", resp.SuccessCount)
	fmt.Println("failure count:", resp.FailureCount)
	fmt.Println("message id:", resp.Responses[0].MessageID)
	fmt.Println("error msg:", resp.Responses[0].Error)
	return nil
}
func (s *FCMService) SendFCMMessage(token, title, body string) error {
	if s.serverKey == nil {
		return fmt.Errorf("server key is not set")
	}
	fcmMessage := FCMMessage{
		To: token,
		Data: map[string]string{
			"title": title,
			"body":  body,
		},
	}

	jsonBytes, err := json.Marshal(fcmMessage)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+*s.serverKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(bodyBytes))

	return nil
}
