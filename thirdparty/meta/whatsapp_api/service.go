package whatsapp_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type WhatsAppAPIService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	storageProvider string
	accessToken     *string
	facebookBaseURL string
}

func NewWhatsAppAPIService(db *gorm.DB, ctx *context.ERPContext, facebookBaseURL string, storageProvider string) *WhatsAppAPIService {
	return &WhatsAppAPIService{
		db:              db,
		ctx:             ctx,
		facebookBaseURL: facebookBaseURL,
		storageProvider: storageProvider,
	}
}

func (w *WhatsAppAPIService) SetAccessToken(accessToken *string) {
	w.accessToken = accessToken
}

func (w *WhatsAppAPIService) WhatsappApiWebhook(
	req *http.Request,
	data objects.WhatsappApiWebhookRequest,
	waSession string,
	getContact func(phoneNumber, dispayName string, companyID *string) (*models.ContactModel, error),
	getSession func(phoneNumberID string, companyID *string) (*objects.WhatsappApiSession, error),
	getMessageData func(phoneNumberID string, msg *models.WhatsappMessageModel) error,
) error {
	// if w.accessToken == nil {
	// 	return errors.New("access token not set")
	// }
	now := time.Now()
	var companyID *string
	if req.Header.Get("ID-Company") != "" {
		compID := req.Header.Get("company_id")
		companyID = &compID
	}
	for _, entry := range data.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" && change.Value.MessagingProduct == "whatsapp" {

				if len(change.Value.Contacts) > 0 {
					phoneNumberID := change.Value.Metadata.PhoneNumberID
					fmt.Println("PHONE NUMBER ID", phoneNumberID)
					sessionData, err := getSession(phoneNumberID, companyID)
					if err == nil {
						waSession = sessionData.Session
						if sessionData.AccessToken != "" {
							w.SetAccessToken(&sessionData.AccessToken)
						}
					}

					contact, err := getContact(change.Value.Contacts[0].WAID, change.Value.Contacts[0].Profile.Name, companyID)
					if err != nil {
						fmt.Println("ERROR GET CONTACT BY PHONE NUMBER ID", err)
						continue
					}

					// GET MESSAGE
					for _, msg := range change.Value.Messages {
						message := ""
						// QUOTE MESSAGE

						waMsgID := utils.Uuid()

						if msg.Type == "text" {
							message = msg.Text.Body
						}
						var mediaUrl, mimeType string
						if msg.Type == "image" && msg.Image != nil {
							message = msg.Image.Caption
							path, err := w.GetMedia(msg.Image.ID, phoneNumberID)
							if err != nil {
								fmt.Println("ERROR", err)
							}
							if path != nil {
								mediaUrl = path.URL
								mimeType = path.MimeType
								path.RefID = waMsgID
								path.RefType = "whatsapp_message"
								if err := w.db.Save(path).Error; err != nil {
									fmt.Println("ERROR CREATE WHATSAPP MESSAGE #1", err)
									continue
								}
							}
						}

						session := fmt.Sprintf("%s@%s", *contact.Phone, waSession)
						waMsg := models.WhatsappMessageModel{
							Message:   message,
							MessageID: &msg.ID,
							Sender:    msg.From,
							JID:       phoneNumberID,
							Contact:   contact,
							SentAt:    &now,
							Session:   session,
							CompanyID: companyID,
							MediaURL:  mediaUrl,
							MimeType:  mimeType,
						}

						if msg.Context != nil {
							waMsg.QuotedMessageID = &msg.Context.ID
						}

						// utils.LogJson(waMsg)
						waMsg.ID = waMsgID
						// if err := w.db.Create(&waMsg).Error; err != nil {
						// 	fmt.Println("ERROR CREATE WHATSAPP MESSAGE #2", err)
						// 	continue
						// }
						err = getMessageData(phoneNumberID, &waMsg)
						if err != nil {
							fmt.Println("ERROR GET MESSAGE DATA", err)
							continue
						}

					}

				}
			}

		}

	}
	return nil
}

func (w *WhatsAppAPIService) SendMessage(phoneNumberID string, message string, file []*models.FileModel, contact *models.ContactModel, quoteMsgID *string) (*objects.WaResponse, error) {
	imgID := ""
	if len(file) > 0 {
		for _, f := range file {
			if f == nil {
				continue
			}

			imageId, err := w.SendWhatsappApiImage(phoneNumberID, contact, f.Path, f.MimeType)
			if err != nil {
				return nil, err
			}

			fmt.Println("IMAGE ID", *imageId)
			imgID = *imageId
		}

	}
	var payload map[string]any
	if imgID != "" {
		payload = map[string]any{
			"messaging_product": "whatsapp",
			"recipient_type":    "individual",
			"to":                contact.Phone,
			"type":              "image",
			"image": map[string]any{
				"id":      imgID,
				"caption": message,
			},
		}
	} else {
		payload = map[string]any{
			"messaging_product": "whatsapp",
			"recipient_type":    "individual",
			"to":                contact.Phone,
			"type":              "text",
			"text": map[string]any{
				"body": message,
			},
		}
	}

	// https://graph.facebook.com/{{Version}}/{{Phone-Number-ID}}/messages

	url := fmt.Sprintf("%s/%s/messages", w.facebookBaseURL, phoneNumberID)
	fmt.Println("URL", url, "\nPAYLOAD", payload)

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *w.accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		fmt.Println("ERROR SEND WA MESSAGE", string(body))

		return nil, fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	var waResponse objects.WaResponse
	if err := json.NewDecoder(resp.Body).Decode(&waResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &waResponse, nil
}

func (w *WhatsAppAPIService) GetMedia(mediaID, phoneNumberID string) (*models.FileModel, error) {
	url := fmt.Sprintf("%s/%s?phone_number_id=%s", w.facebookBaseURL, mediaID, phoneNumberID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *w.accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %d", resp.StatusCode)
	}

	var media objects.FacebookMedia
	if err := json.NewDecoder(resp.Body).Decode(&media); err != nil {
		return nil, err
	}

	return w.downloadAndSaveMedia(media.URL, media.ID, media.MimeType)
}

func (w *WhatsAppAPIService) downloadAndSaveMedia(url, fileName, mime string) (*models.FileModel, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %d", resp.StatusCode)
	}

	filePath := path.Join(os.TempDir(), fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return nil, err
	}

	fileModel := &models.FileModel{
		FileName: fileName,
		MimeType: mime,
		Path:     filePath,
	}

	if err := w.db.Create(fileModel).Error; err != nil {
		return nil, err
	}

	return fileModel, nil
}

func (w *WhatsAppAPIService) MarkAsRead(phoneNumberID, incomingMsgID string) error {
	url := fmt.Sprintf("%s/%s/messages", w.facebookBaseURL, phoneNumberID)

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"status":            "read",
		"message_id":        incomingMsgID,
	}

	// fmt.Println("URL", url, "\nPAYLOAD")
	// utils.LogJson(payload)
	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *w.accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}

		return fmt.Errorf("failed to mark as read, status code: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (w *WhatsAppAPIService) SendWhatsappApiImage(phoneNumberID string, contact *models.ContactModel, filePath, mimeType string) (*string, error) {
	url := fmt.Sprintf("%s/%s/media", w.facebookBaseURL, phoneNumberID)

	// Buat buffer untuk menampung body permintaan
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Tambahkan field 'messaging_product'
	_ = writer.WriteField("messaging_product", "whatsapp")

	fmt.Println("UPLOAD", filePath)
	// Buka file yang akan diunggah
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Gagal membuka file:", err)
		return nil, err
	}
	defer file.Close()

	// Buat form-data untuk file
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath)))
	header.Set("Content-Type", mimeType)

	// Buat bagian form-data menggunakan header yang sudah didefinisikan
	filePart, err := writer.CreatePart(header)
	if err != nil {
		return nil, err
	}

	// Salin isi file ke bagian form-data
	_, err = io.Copy(filePart, file)
	if err != nil {
		fmt.Println("Gagal menyalin file ke form-data:", err)
		return nil, err
	}

	// Selesaikan penulisan form-data
	writer.Close()

	// Buat URL endpoint

	// Buat permintaan HTTP POST
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println("Gagal membuat permintaan:", err)
		return nil, err
	}

	// Tambahkan header Authorization dan Content-Type
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *w.accessToken))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Kirim permintaan
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Gagal mengirim permintaan:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// command, _ := http2curl.GetCurlCommand(req)
	// fmt.Println("CURL", command)
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		fmt.Println("ERROR SEND WA MESSAGE", string(body))

		return nil, fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}
	// Baca dan cetak respons
	fmt.Println("Status respons:", resp.Status)

	var waResponse struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&waResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &waResponse.ID, nil
}
