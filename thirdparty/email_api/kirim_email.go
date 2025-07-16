package email_api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
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
func (s KirimEmail) SendEmail(from, domain, apiKey, subject, to, message string, attachment []string) error {
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

	// fmt.Println("domain", domain)
	// fmt.Println("api", apiKey)
	// fmt.Println("from", from)
	// fmt.Println("to", to)

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
