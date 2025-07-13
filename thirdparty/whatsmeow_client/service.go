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
func (s *WhatsmeowService) SetChatData(msg WaMessage) {
	s.chatData = &msg
}
func (s *WhatsmeowService) SendChatMessage() (any, error) {
	if s.chatData == nil {
		return nil, errors.New("chat data is empty")
	}
	return s.SendMessage(*s.chatData)
}

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
