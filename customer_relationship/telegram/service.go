package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TelegramService struct {
	ctx     *context.ERPContext
	botName *string
	token   *string
	input   *TelegramMsg
}

func NewTelegramService(ctx *context.ERPContext) *TelegramService {
	service := &TelegramService{
		ctx: ctx,
	}
	return service
}

func (t *TelegramService) SetInput(input *TelegramMsg) {
	t.input = input
}

func (t *TelegramService) SetToken(botName, token *string) {
	t.botName = botName
	t.token = token
}

func (t *TelegramService) GetWebhookInfo() (map[string]any, error) {
	if t.botName == nil || t.token == nil {
		return nil, errors.New("botName and token must be set")
	}
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", *t.token))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var webhookInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&webhookInfo); err != nil {
		return nil, err
	}
	return webhookInfo, nil
}
func (t *TelegramService) SetWebhook(webhookURL string) error {
	if t.botName == nil || t.token == nil {
		return errors.New("botName and token must be set")
	}
	// log.Println("SET WEBHOOK", fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook?url=%s", *t.token, webhookURL))
	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook?url=%s", *t.token, webhookURL), url.Values{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("set webhook failed")
	}
	return nil
}

func (t *TelegramService) GetUserProfilePhotos(userId int64) (map[string]any, error) {
	if t.botName == nil || t.token == nil {
		return nil, errors.New("botName and token must be set")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUserProfilePhotos?user_id=%d", *t.token, userId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get user profile photos")
	}

	var result struct {
		Ok     bool                 `json:"ok"`
		Result models.TelegramPhoto `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	fileResp := map[string]any{}
	for _, v := range result.Result.Photos {
		lastV := v[len(v)-1]
		fileResp, err = t.GetFile(lastV.FileID)
		if err != nil {
			return nil, err
		}
		break
	}
	return fileResp, nil
}

func (t *TelegramService) GetFile(fileId string) (map[string]interface{}, error) {
	if t.botName == nil || t.token == nil {
		return nil, errors.New("botName and token must be set")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", *t.token, fileId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get file")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
func (t *TelegramService) GetMe() (map[string]interface{}, error) {
	if t.botName == nil || t.token == nil {
		return nil, errors.New("botName and token must be set")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", *t.token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get bot information")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (ws *TelegramService) SendCSMessage() (any, error) {
	if ws.input == nil {
		return nil, errors.New("input is nil")
	}

	response, err := ws.SendTelegramMessage(ws.input)
	if err != nil {
		return nil, err
	}

	// utils.LogJson(response)
	if response != nil {
		mID, ok := response["result"].(map[string]any)["message_id"].(float64)
		if !ok {
			return nil, errors.New("failed to get message ID")
		}

		msgID := fmt.Sprintf("%.0f", mID)
		ws.input.Data.MessageID = &msgID
	}

	if ws.input.Save && ws.input.Data != nil {
		if err := ws.SaveMessage(ws.input.Data); err != nil {
			return nil, err
		}
	}

	return ws.input.Data, nil
}

func (t *TelegramService) SendTelegramMessage(input *TelegramMsg) (map[string]any, error) {
	if t.botName == nil || t.token == nil {
		return nil, errors.New("botName and token must be set")
	}
	// Create payload for the Telegram Bot API
	payload := map[string]any{
		"chat_id": input.ChatID,
		"text":    input.Message,
	}
	if input.FileURL != "" {
		if strings.Contains(input.MimeType, "audio") {
			payload["audio"] = input.FileURL
			payload["caption"] = input.Message
		} else if strings.Contains(input.MimeType, "image") {
			payload["photo"] = input.FileURL
			payload["caption"] = input.Message
		} else if strings.Contains(input.MimeType, "video") {
			payload["video"] = input.FileURL
			payload["caption"] = input.Message
		} else {
			payload["document"] = input.FileURL
			payload["caption"] = input.Message
		}
		if input.FileCaption != "" {
			payload["caption"] = input.FileCaption
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {

		return nil, err
	}

	// Send POST request to Telegram Bot API
	url := "https://api.telegram.org/bot" + *t.token + "/sendMessage"
	if input.FileURL != "" {
		if strings.Contains(input.MimeType, "image") {
			url = "https://api.telegram.org/bot" + *t.token + "/sendPhoto"
		} else if strings.Contains(input.MimeType, "audio") {
			url = "https://api.telegram.org/bot" + *t.token + "/sendAudio"
		} else if strings.Contains(input.MimeType, "video") {
			url = "https://api.telegram.org/bot" + *t.token + "/sendVideo"
		} else {
			url = "https://api.telegram.org/bot" + *t.token + "/sendDocument"
		}
	}

	fmt.Println("URL:", url)
	utils.LogJson(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	defer resp.Body.Close()

	// Log the sent message for analytics
	// err = logTelegramMessage(input.ChatID, input.Message)
	// if err != nil {

	// 	return err
	// }

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return response, nil
}

// func logTelegramMessage(chatID int64, message string) error {
// 	action := models.MessageLog{
// 		Platform:    "telegram",
// 		RecipientID: string(chatID),
// 		Message:     message,
// 		Status:      "received",
// 	}

// 	if err := database.DB.Create(&action).Error; err != nil {
// 		return err
// 	}
// 	return nil
// }

type TelegramMsg struct {
	ChatID      int64  `json:"chat_id"`
	Message     string `json:"message"`
	FileURL     string `json:"file_url"`
	FileCaption string `json:"file_caption"`
	MimeType    string `json:"mime_type"`
	Save        bool   `json:"save"`
	Data        *models.TelegramMessage
}

func (t *TelegramService) CheckSession(resp *models.TGResponse, input *models.ContactModel, connectionID, companyID string) (*models.TelegramMessageSession, error) {

	now := time.Now()
	// Create payload for the Telegram Bot API
	var sessions models.TelegramMessageSession
	err := t.ctx.DB.Where("contact_id = ?", input.ID).First(&sessions).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		sessions = models.TelegramMessageSession{
			ContactID:    &input.ID,
			SessionName:  input.Name,
			Session:      connectionID,
			LastMessage:  resp.Message.Text,
			LastOnlineAt: &now,
			CompanyID:    &companyID,
		}
		err = t.ctx.DB.Create(&sessions).Error
		if err != nil {
			return nil, err
		}
	}

	if err == nil {
		sessions.LastMessage = resp.Message.Text
		sessions.LastOnlineAt = &now
		return &sessions, t.ctx.DB.Save(&sessions).Error
	}

	return &sessions, nil
}

func (t *TelegramService) SaveMessage(msg *models.TelegramMessage) error {
	if err := t.ctx.DB.Create(msg).Error; err != nil {
		return err
	}
	return nil
}

func (ws *TelegramService) GetSessionMessageBySessionName(sessionName string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.ctx.DB.Preload("Contact.Tags").Model(&models.TelegramMessageSession{})

	if sessionName != "" {
		stmt = stmt.Where("session = ?", sessionName)
	}

	if request.URL.Query().Get("search") != "" || request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.
			Joins("LEFT JOIN contacts ON contacts.id = telegram_message_sessions.contact_id").
			Joins("LEFT JOIN contact_tags ON contact_tags.contact_model_id = contacts.id").
			Joins("LEFT JOIN tags ON tags.id = contact_tags.tag_model_id")
	}

	// if request.URL.Query().Get("is_unread") != "" || request.URL.Query().Get("is_unreplied") != "" {
	// 	stmt = stmt.
	// 		Joins("LEFT JOIN whatsapp_messages ON whatsapp_messages.session = telegram_message_sessions.session")
	// 	if request.URL.Query().Get("is_unread") != "" && request.URL.Query().Get("is_unreplied") != "" {
	// 		fmt.Println("is_unread && is_unreplied")
	// 		stmt = stmt.Where("(whatsapp_messages.is_read = ? AND whatsapp_messages.is_from_me = ?)  OR (whatsapp_messages.is_replied = ? AND whatsapp_messages.is_from_me = ?)", false, false, false, false)

	// 	} else if request.URL.Query().Get("is_unread") != "" {
	// 		fmt.Println("is_unread")
	// 		stmt = stmt.Where("(whatsapp_messages.is_read = ? AND whatsapp_messages.is_from_me = ?) ", false, false)
	// 	} else if request.URL.Query().Get("is_unreplied") != "" {
	// 		fmt.Println("is_unreplied")
	// 		stmt = stmt.Where("whatsapp_messages.is_replied = ? AND whatsapp_messages.is_from_me = ?", false, false)
	// 	}
	// }

	if request.URL.Query().Get("tag_ids") != "" {
		stmt = stmt.Where("tags.id in (?)", strings.Split(request.URL.Query().Get("tag_ids"), ","))
	}

	if request.Header.Get("ID-Company") != "" {
		if request.Header.Get("ID-Company") == "nil" || request.Header.Get("ID-Company") == "null" {
			stmt = stmt.Where("telegram_message_sessions.company_id is null")
		} else {
			stmt = stmt.Where("telegram_message_sessions.company_id = ?", request.Header.Get("ID-Company"))

		}
	}

	if request.URL.Query().Get("connection_session") != "" {
		stmt = stmt.Where("telegram_message_sessions.session = ?", request.URL.Query().Get("connection_session"))

	}

	stmt = stmt.Order("last_online_at DESC")

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TelegramMessageSession{})
	page.Page = page.Page + 1

	items := page.Items.(*[]models.TelegramMessageSession)
	newItems := make([]models.TelegramMessageSession, 0)
	for _, item := range *items {
		// fmt.Println("CONTACT", item.Contact)
		profile, err := item.Contact.GetProfilePicture(ws.ctx.DB)
		if err == nil {
			item.Contact.ProfilePicture = profile
		}

		// utils.LogJson(profile)

		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

func (ws *TelegramService) GetMessageSessionChatBySessionName(sessionName string, contact_id *string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ws.ctx.DB.Preload("Member.User").Preload("Contact").Model(&models.TelegramMessage{})

	if sessionName != "" {
		stmt = stmt.Where("telegram_message_session_id = ?", sessionName)
	}
	if contact_id != nil {
		stmt = stmt.Where("contact_id = ?", *contact_id)
	}

	stmt = stmt.Order("created_at DESC")

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TelegramMessage{})
	return page, nil
}
