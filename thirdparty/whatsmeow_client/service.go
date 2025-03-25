package whatsmeow_client

import (
	"bytes"
	"encoding/json"
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

func (s *WhatsmeowService) SendMessage(msg WaMessage) (map[string]interface{}, error) {
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
