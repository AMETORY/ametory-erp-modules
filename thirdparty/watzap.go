package thirdparty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Package thirdparty provides a client for interacting with the Watzap API.

const WatzapURL = "https://api.watzap.id"

// WatzapClient is a client for interacting with the Watzap API.
type WatzapClient struct {
	ApiKey     string
	NumberKey  string
	MockNumber string
	RedisKey   string
	IsMock     bool
}

// SendMessageRequest represents the request payload for sending a message.
type SendMessageRequest struct {
	PhoneNo string `json:"phone_no"`
	Message string `json:"message"`
}

// SendFileURLRequest represents the request payload for sending a file URL.
type SendFileURLRequest struct {
	PhoneNo string `json:"phone_no"`
	URL     string `json:"url"`
}

// SendImageURLRequest represents the request payload for sending an image URL.
type SendImageURLRequest struct {
	PhoneNo         string `json:"phone_no"`
	URL             string `json:"url"`
	Message         string `json:"message"`
	SeparateCaption int    `json:"separate_caption"`
}

// NewWatzapClient creates a new instance of WatzapClient with the given parameters.
//
// If redisKey is empty, it will be set to "watzap:notif".
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

// SendMessage sends a message to the specified phone number using the Watzap API.
//
// It returns an error if the message could not be sent.
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

// SendFileURL sends a file URL to the specified phone number using the Watzap API.
//
// It returns an error if the file URL could not be sent.
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

// SendImageURL sends an image URL with an optional message to the specified phone number using the Watzap API.
//
// It returns an error if the image URL could not be sent.
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

// sendRequest sends a HTTP POST request to the specified URL with the given JSON payload.
//
// It returns an error if the request fails or the response status is not OK.
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
		return fmt.Errorf("failed to send request, status code: %d", resp.StatusCode)
	}

	return nil
}
