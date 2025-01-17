package thirdparty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const WatzapURL = "https://api.watzap.id"

type WatzapClient struct {
	ApiKey     string
	NumberKey  string
	MockNumber string
	RedisKey   string
	IsMock     bool
}

type SendMessageRequest struct {
	PhoneNo string `json:"phone_no"`
	Message string `json:"message"`
}

type SendFileURLRequest struct {
	PhoneNo string `json:"phone_no"`
	URL     string `json:"url"`
}

type SendImageURLRequest struct {
	PhoneNo         string `json:"phone_no"`
	URL             string `json:"url"`
	Message         string `json:"message"`
	SeparateCaption int    `json:"separate_caption"`
}

func NewWatzapClient(apiKey, numberKey, mockNumber string, isMock bool, redisKey string) *WatzapClient {
	if redisKey == "" {
		redisKey = "watzap:notif"
	}
	return &WatzapClient{
		ApiKey:     apiKey,
		NumberKey:  numberKey,
		MockNumber: mockNumber,
		IsMock:     isMock,
		RedisKey:   redisKey,
	}
}

func (c *WatzapClient) SendMessage(phoneNo, message string) error {
	req := SendMessageRequest{
		PhoneNo: phoneNo,
		Message: message,
	}

	phoneNumber := req.PhoneNo
	if c.IsMock {
		phoneNumber = c.MockNumber
	}

	jsonReq, err := json.Marshal(map[string]interface{}{
		"api_key":    c.ApiKey,
		"number_key": c.NumberKey,
		"phone_no":   phoneNumber,
		"message":    req.Message,
	})
	if err != nil {
		return err
	}

	return c.sendRequest(WatzapURL+"/v1/send_message", jsonReq)
}

func (c *WatzapClient) SendFileURL(phoneNo, url string) error {
	req := SendFileURLRequest{
		PhoneNo: phoneNo,
		URL:     url,
	}

	phoneNumber := req.PhoneNo
	if c.IsMock {
		phoneNumber = c.MockNumber
	}

	jsonReq, err := json.Marshal(map[string]interface{}{
		"api_key":    c.ApiKey,
		"number_key": c.NumberKey,
		"phone_no":   phoneNumber,
		"url":        req.URL,
	})
	if err != nil {
		return err
	}

	return c.sendRequest(WatzapURL+"/v1/send_file_url", jsonReq)
}

func (c *WatzapClient) SendImageURL(phoneNo, url, message string, separateCaption int) error {
	req := SendImageURLRequest{
		PhoneNo:         phoneNo,
		URL:             url,
		Message:         message,
		SeparateCaption: separateCaption,
	}

	phoneNumber := req.PhoneNo
	if c.IsMock {
		phoneNumber = c.MockNumber
	}

	jsonReq, err := json.Marshal(map[string]interface{}{
		"api_key":          c.ApiKey,
		"number_key":       c.NumberKey,
		"phone_no":         phoneNumber,
		"url":              req.URL,
		"message":          req.Message,
		"separate_caption": req.SeparateCaption,
	})
	if err != nil {
		return err
	}

	return c.sendRequest(WatzapURL+"/v1/send_image_url", jsonReq)
}

func (c *WatzapClient) sendRequest(url string, jsonReq []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gagal mengirim request, status code: %d", resp.StatusCode)
	}

	return nil
}
