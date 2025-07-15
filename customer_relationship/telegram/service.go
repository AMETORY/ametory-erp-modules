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

// NewTelegramService initializes a new instance of TelegramService with the provided context.
// It sets the context for the service which will be used for various operations within the service.
// Returns a pointer to the newly created TelegramService.

func NewTelegramService(ctx *context.ERPContext) *TelegramService {
	service := &TelegramService{
		ctx: ctx,
	}
	return service
}

// SetInput sets the input message for the Telegram service to use in the SendTelegramMessage method.
func (t *TelegramService) SetInput(input *TelegramMsg) {
	t.input = input
}

// SetToken sets the bot name and token for the Telegram service.
// These must be set before making any API requests that require authentication.

func (t *TelegramService) SetToken(botName, token *string) {
	t.botName = botName
	t.token = token
}

// GetWebhookInfo retrieves the current status of the webhook for the Telegram bot.
// It sends a GET request to the Telegram Bot API to obtain information about the webhook set for the bot.
// Returns a map containing the webhook information if successful, or an error if the request fails or bot credentials are not set.

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

// SetWebhook sets the webhook for the Telegram bot.
// It sends a POST request to the Telegram Bot API to set the webhook for the bot.
// Returns an error if the request fails or bot credentials are not set.
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

// GetUserProfilePhotos retrieves the user profile photos for the given user ID.
//
// It sends a GET request to the Telegram Bot API to obtain the user profile photos.
// Returns a map containing the last photo's file information if successful, or an error if the request fails or bot credentials are not set.
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
	}
	return fileResp, nil
}

// GetFile retrieves file information from the Telegram Bot API using the provided file ID.
//
// It constructs a request URL with the bot token and file ID, then sends a GET request to the Telegram API.
// Returns a map containing the file information if successful, or an error if the request fails
// or the bot credentials are not set.

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

// GetMe retrieves bot information from the Telegram Bot API.
//
// It constructs a request URL with the bot token, then sends a GET request to the Telegram API.
// Returns a map containing the bot information if successful, or an error if the request fails
// or the bot credentials are not set.
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

// SendCSMessage sends a customer service message using the input provided to the Telegram service.
//
// It first checks if the input is nil and returns an error if it is. Then, it calls SendTelegramMessage
// to send the message and processes the response to extract the message ID, which is stored in the input data.
// If the input specifies that the message should be saved, it calls SaveMessage to save the message data.
//
// Returns the message data if successful, or an error if any step fails.

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

// SendTelegramMessage sends a message to Telegram.
//
// It takes a TelegramMsg object as input, which must contain a chat ID and a message.
// If the input specifies a file, it will be sent as a document, photo, audio or video
// depending on the MIME type.
//
// Returns a map containing the response from the Telegram Bot API, or an error if the request fails.
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

// CheckSession checks and updates the session information for a given Telegram message.
//
// It takes a TGResponse, a ContactModel, a connection ID, and a company ID as parameters.
// The function first attempts to find an existing TelegramMessageSession for the contact.
// If no session is found, a new session is created with the provided information.
// If a session is found, it updates the LastMessage and LastOnlineAt fields with the latest data.
// Returns the TelegramMessageSession and an error, if any occurs during the database operations.

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

// SaveMessage saves a Telegram message to the database.
//
// It takes a pointer to a models.TelegramMessage struct as an argument and returns an error.
// The function uses the provided context's database connection to create a new record in the
// telegram_messages table. If the operation fails, it returns the error.
func (t *TelegramService) SaveMessage(msg *models.TelegramMessage) error {
	if err := t.ctx.DB.Create(msg).Error; err != nil {
		return err
	}
	return nil
}

// GetSessionMessageBySessionName retrieves a paginated list of Telegram message sessions for a specific session name, search query, and/or tags.
//
// It takes a session name, an optional search query, and an HTTP request as parameters.
// The function filters message sessions by session name and optionally by search query and tags.
// It returns a paginated page of TelegramMessageSession models and an error if the operation fails.
//
// The function uses request parameters to modify the pagination and filtering behavior.
// The following query parameters are supported:
//
//   - search: a string to search in the contact's name and email.
//   - tag_ids: a comma-separated list of tag IDs to filter the results.
//   - ID-Company: a header to filter the results by company ID. If the header is
//     set to "nil" or "null", only message sessions with a null company ID are returned.
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

// GetMessageSessionChatBySessionName retrieves a paginated list of Telegram messages for a specific session and contact.
//
// It takes a session name, an optional contact ID pointer, and an HTTP request as parameters.
// The function filters messages by session ID and optionally by contact ID. It returns a paginated
// page of TelegramMessage models and an error if the operation fails.
//
// The function uses request parameters to modify the pagination and filtering behavior.
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
