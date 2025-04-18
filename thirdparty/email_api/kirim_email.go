package email_api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type KirimEmail struct {
}

func (s KirimEmail) SendEmail(from, domain, apiKey, subject, to, message string, attachment []string) error {
	fmt.Println(from, domain, apiKey, subject, to, message)
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
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("domain", domain)
	req.SetBasicAuth("api", apiKey)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
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
