package instagram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

type InstagramService struct {
	ctx *context.ERPContext
}

func NewInstagramService(ctx *context.ERPContext) *InstagramService {
	service := &InstagramService{
		ctx: ctx,
	}
	return service
}

func (t *InstagramService) SendInstagramMessage(from, to, message, accessToken string) error {
	// var instagramToken = config.GetEnv("INSTAGRAM_API_TOKEN")
	payload := objects.FacebookWebhookMessaging{
		Recipient: objects.FacebookWebhookRecipient{
			ID: to,
		},
		Message: objects.FacebookWebhookMessage{
			Text: message,
		},
	}

	log.Println("SEND MSG", fmt.Sprintf("https://graph.instagram.com/v21.0/%s/messages", from))
	log.Println("accessToken", accessToken)
	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://graph.instagram.com/v21.0/%s/messages", from), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("RESPONSE", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}
	return nil
}

func (t *InstagramService) SaveMessage(msg *models.InstagramMessage) error {
	if err := t.ctx.DB.Create(msg).Error; err != nil {
		return err
	}
	return nil
}

func (t *InstagramService) GetSessionMessageBySessionName(sessionName string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := t.ctx.DB.Preload("Contact.Tags").Model(&models.InstagramMessageSession{})

	if sessionName != "" {
		stmt = stmt.Where("session = ?", sessionName)
	}

	if request.URL.Query().Get("search") != "" || request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.
			Joins("LEFT JOIN contacts ON contacts.id = instagram_message_sessions.contact_id").
			Joins("LEFT JOIN contact_tags ON contact_tags.contact_model_id = contacts.id").
			Joins("LEFT JOIN tags ON tags.id = contact_tags.tag_model_id")
	}

	if request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.Where("tags.id in (?)", strings.Split(request.URL.Query().Get("tag_ids"), ","))
	}

	if request.Header.Get("ID-Company") != "" {
		if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
			stmt = stmt.Where("instagram_message_sessions.company_id is null")
		} else {
			stmt = stmt.Where("instagram_message_sessions.company_id = ?", request.Header.Get("ID-Company"))

		}
	}

	stmt = stmt.Order("last_online_at DESC")

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WhatsappMessageSession{})
	return page, nil
}

func (t *InstagramService) GetMessageSessionChatBySessionName(sessionName string, contact_id *string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := t.ctx.DB.Preload("Member.User").Preload("Contact").Model(&models.InstagramMessage{})

	if sessionName != "" {
		stmt = stmt.Where("instagram_message_session_id = ?", sessionName)
	}
	if contact_id != nil {
		stmt = stmt.Where("contact_id = ?", *contact_id)
	}

	stmt = stmt.Order("created_at DESC")

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.InstagramMessage{})
	return page, nil
}
