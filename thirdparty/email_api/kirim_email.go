package email_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/utils"
)

type KirimEmail struct {
}

// SendEmail sends an email using the kirim.email API.
//
// It takes the following parameters:
//   - from: the sender's email address
//   - domain: the domain to use for the email
//   - apiKey: the API key for the kirim.email domain
//   - subject: the subject of the email
//   - to: the recipient's email address
//   - message: the email message
//   - attachment: an array of strings representing the paths to the files to attach
//
// It returns an error if something goes wrong.
func (s KirimEmail) SendEmail(from, domain, apiKey, apiSecret, subject, to, message string, attachment []string) error {
	if apiSecret != "" {
		// USE VERSION 4
		return sendEmailV4(from, domain, apiKey, apiSecret, subject, to, message, attachment)
	}
	// fmt.Println(from, domain, apiKey, subject, to, message)
	url := "https://aplikasi.kirim.email/api/v3/transactional/messages"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("from", from)
	_ = writer.WriteField("to", to)
	_ = writer.WriteField("subject", subject)
	_ = writer.WriteField("html", message)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return err
	}
	req.Header.Set("domain", domain)
	req.SetBasicAuth("api", apiKey)

	fmt.Println("domain", domain)
	fmt.Println("api", apiKey)
	fmt.Println("from", from)
	fmt.Println("to", to)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(body))
	return nil
}

func sendEmailV4(from, domain, apiKey, apiSecret, subject, to, message string, attachment []string) error {
	apiURL := "https://smtp-app.kirim.email/api/v4/transactional/message"

	data := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
		ReplyTo string   `json:"reply_to,omitempty"`
	}{
		From:    from,
		To:      strings.Split(to, ","),
		Subject: subject,
		Text:    message,
		ReplyTo: from,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	utils.LogJson(data)
	fmt.Println("apiURL", apiURL)
	fmt.Println("domain", domain)
	fmt.Println("api", apiKey)
	fmt.Println("from", from)
	fmt.Println("to", to)

	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	req.SetBasicAuth(apiKey, apiSecret)
	req.Header.Set("domain", domain)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(body))
	return nil
}
