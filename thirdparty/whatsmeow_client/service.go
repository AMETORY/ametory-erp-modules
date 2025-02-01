package whatsmeow_client

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
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

func (s *WhatsmeowService) CreateQR(sessionID string, webhook string) error {
	req, err := http.NewRequest("POST", s.BaseURL+"/v1/create-qr", bytes.NewBufferString(`{"session_id":"`+sessionID+`","webhook":"`+webhook+`"}`))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
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

func (s *WhatsmeowService) UpdateWebhook(sessionID string, webhook string) error {
	req, err := http.NewRequest("PUT", s.BaseURL+"/v1/update-webhook/"+sessionID, bytes.NewBufferString(`{"webhook":"`+webhook+`"}`))
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
