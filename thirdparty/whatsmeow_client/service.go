package whatsmeow_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type WhatsmeowService struct {
	BaseURL    string
	MockNumber string
	RedisKey   string
	IsMock     bool
	chatData   *WaMessage
}

// NewWhatsmeowService creates a new instance of WhatsmeowService with the given parameters.
//
// If redisKey is empty, it will be set to "whatsmeow:notif".
func NewWhatsmeowService(baseURL, mockNumber string, isMock bool, redisKey string) *WhatsmeowService {
	if redisKey == "" {
		redisKey = "whatsmeow:notif"
	}
	return &WhatsmeowService{
		BaseURL:    baseURL,
		MockNumber: mockNumber,
		IsMock:     isMock,
		RedisKey:   redisKey,
	}
}

// SetChatData sets the chat data for the WhatsmeowService instance.
//
// It takes a WaMessage object as the parameter and sets the chat data for the
// instance to the given parameter.
func (s *WhatsmeowService) SetChatData(msg WaMessage) {
	s.chatData = &msg
}

// SendChatMessage sends a chat message using the previously set chat data.
//
// It returns the response from the Whatsmeow API if the message is sent
// successfully, or an error if any step fails.
func (s *WhatsmeowService) SendChatMessage() (any, error) {
	if s.chatData == nil {
		return nil, errors.New("chat data is empty")
	}
	return s.SendMessage(*s.chatData)
}

// SendTyping sends a typing message to the recipient.
//
// It takes a WaMessage object as the parameter and sends a typing message to the
// recipient using the Whatsmeow API.
//
// It returns the response from the Whatsmeow API if the message is sent
// successfully, or an error if any step fails.
func (s *WhatsmeowService) SendTyping(msg WaMessage) (any, error) {
	if s.IsMock && s.MockNumber != "" {
		msg.To = s.MockNumber
		msg.IsGroup = false
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req, err := http.NewRequest("POST", s.BaseURL+"/v1/send-typing", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return response, nil
}

// SendMessage sends a message using the Whatsmeow API.
//
// It takes a WaMessage object as the parameter and sends the message to the
// recipient using the Whatsmeow API.
//
// It returns the response from the Whatsmeow API if the message is sent
// successfully, or an error if any step fails.
func (s *WhatsmeowService) SendMessage(msg WaMessage) (any, error) {
	if s.IsMock && s.MockNumber != "" {
		msg.To = s.MockNumber
		msg.IsGroup = false
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req, err := http.NewRequest("POST", s.BaseURL+"/v1/send-message", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return response, nil
}

// GetContact retrieves a list of contacts from the Whatsmeow API based on the provided parameters.
//
// It takes the following parameters:
//   - JID: the unique identifier for the session.
//   - search: a string to filter the contacts by name or other attributes.
//   - page: a string representing the page number for pagination.
//   - limit: a string representing the maximum number of contacts to retrieve per page.
//
// The function returns the raw response body as a byte slice and an error if any step in the process fails.
func (s *WhatsmeowService) GetContact(JID, search, page, limit string) ([]byte, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/contacts?jid="+JID+"&search="+search+"&limit="+limit+"&page="+page, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// CheckNumber checks whether a given number is valid and available for use with the Whatsmeow API.
//
// It takes two parameters: JID, the unique identifier for the session and number, the phone number to be checked.
//
// The function returns the raw response body as a byte slice and an error if any step in the process fails.
func (s *WhatsmeowService) CheckNumber(JID string, number string) ([]byte, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/check-number/"+JID+"/"+number, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// CheckConnected checks the connection status of a given JID with the Whatsmeow API.
//
// It sends a GET request to the Whatsmeow API and returns the raw response body as a byte slice.
// The function takes a JID as a string parameter and returns an error if any step in the process fails.
func (s *WhatsmeowService) CheckConnected(JID string) ([]byte, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/connected/"+JID, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *WhatsmeowService) DisconnectDeviceByJID(JID string) error {
	req, err := http.NewRequest("DELETE", s.BaseURL+"/v1/device-delete/"+JID, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	return nil
}

// CreateQR creates a new QR code on the Whatsmeow API.
//
// It takes a sessionID, webhook URL, and headerKey as string parameters and returns a byte slice and an error.
// The byte slice contains the QR code data in JSON format.
// The function sends a POST request to the Whatsmeow API with the sessionID, webhook URL, and headerKey as JSON payload.
// If there is an error, the method returns an error.
func (s *WhatsmeowService) CreateQR(sessionID, webhook, headerKey string) ([]byte, error) {
	req, err := http.NewRequest("POST", s.BaseURL+"/v1/create-qr", bytes.NewBufferString(`{"session":"`+sessionID+`","webhook":"`+webhook+`", "header_key":"`+headerKey+`"}`))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetDevices retrieves the list of devices on the Whatsmeow API.
//
// It takes no parameters and returns the raw response body as a byte slice and an error if any step in the process fails.
func (s *WhatsmeowService) GetDevices() ([]byte, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/devices", nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetQRImage retrieves the QR image for a given session ID from the Whatsmeow API.
//
// It sends a GET request to the API and returns the raw QR image data as a byte slice.
// If any step in the process fails, it returns an error.
func (s *WhatsmeowService) GetQRImage(sessionID string) ([]byte, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/get-qr-image/"+sessionID, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetQR retrieves the QR code string for a given session ID from the Whatsmeow API.
//
// It sends a GET request to the API endpoint with the specified session ID and
// returns the QR code string if successful. If any step in the process fails,
// it returns an error.
func (s *WhatsmeowService) GetQR(sessionID string) (string, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/get-qr/"+sessionID, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response struct {
		Message  string `json:"message"`
		Response string `json:"response"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return response.Response, nil
}

// UpdateWebhook updates the webhook for a given session ID.
//
// It sends a PUT request to the Whatsmeow API with the specified session ID,
// webhook URL, and header key as JSON payload. If the request is successful,
// it returns nil. If any step in the process fails, it returns an error.
func (s *WhatsmeowService) UpdateWebhook(sessionID string, webhook, headerKey string) error {
	// fmt.Println(s.BaseURL + "/v1/update-webhook/" + sessionID)
	// fmt.Println(`{"webhook":"` + webhook + `", "header_key":"` + headerKey + `"}`)
	req, err := http.NewRequest("PUT", s.BaseURL+"/v1/update-webhook/"+sessionID, bytes.NewBufferString(`{"webhook":"`+webhook+`", "header_key":"`+headerKey+`"}`))
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	return nil
}

// GetJIDBySessionName retrieves a JSON object containing the JID for a given session name from the Whatsmeow API.
//
// It sends a GET request to the Whatsmeow API with the session name as a URL parameter.
// If the request is successful, it returns a map[string]interface{} containing the JID
// and any other response data. If any step in the process fails, it returns an error.
func (s *WhatsmeowService) GetJIDBySessionName(sessionName string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/jid/"+sessionName, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return response, nil
}

// MarkAsRead marks messages as read in the Whatsmeow API.
//
// It sends a PUT request to the Whatsmeow API with a JSON payload containing
// the message IDs and the chat ID (sender phone number).
// If the request is successful, the messages are marked as read in the Whatsmeow API.
// If any step in the process fails, it returns an error.
func (s *WhatsmeowService) MarkAsRead(sessionID string, msgIDs []string, senderPhoneNumber string) error {
	var data = map[string]interface{}{
		"msg_ids": msgIDs,
		"chat_id": senderPhoneNumber,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// fmt.Println(s.BaseURL + "/v1/update-webhook/" + sessionID)
	// fmt.Println(`{"webhook":"` + webhook + `", "header_key":"` + headerKey + `"}`)
	req, err := http.NewRequest("PUT", s.BaseURL+"/v1/message/"+sessionID+"/mark-read", bytes.NewBufferString(string(jsonData)))
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	return nil
}

// DeviceDelete deletes a device from the Whatsmeow API.
//
// It sends a DELETE request to the Whatsmeow API with the JID of the device to be deleted.
// If the request is successful, the device is deleted from the Whatsmeow API.
// If any step in the process fails, it returns an error.
func (s *WhatsmeowService) DeviceDelete(jid string) error {
	req, err := http.NewRequest("DELETE", s.BaseURL+"/v1/device-delete/"+jid, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	return nil
}

// GetGroupInfo retrieves information about a specific group from the Whatsmeow API.
//
// It sends a GET request to the Whatsmeow API with the JID and group ID as URL parameters.
// If the request is successful, it returns a map containing the group information and
// an error if any step in the process fails.
func (s *WhatsmeowService) GetGroupInfo(JID string, groupID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", s.BaseURL+"/v1/get-group-info/"+JID+"/"+groupID, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return response, nil
}
