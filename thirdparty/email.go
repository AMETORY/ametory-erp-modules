package thirdparty

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/mail"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// SMTPSender SMTPSender
type SMTPSender struct {
	smtpServer     string
	smtpPort       int
	smtpUsername   string
	smtpPassword   string
	layoutTemplate string
	bodyTemplate   string
	body           string
	from           mail.Address
	to             []mail.Address
}

// NewSMTPSender NewSMTPSender
func NewSMTPSender(smtpServer string, smtpPort int, smtpUsername, smtpPassword string, from mail.Address) *SMTPSender {
	return &SMTPSender{
		smtpServer:   smtpServer,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		from:         from,
		to:           []mail.Address{},
	}
}

func (s *SMTPSender) SetTemplate(layout string, template string) *SMTPSender {
	s.layoutTemplate = layout
	s.bodyTemplate = template
	return s
}
func (s *SMTPSender) SetAddress(name string, email string) *SMTPSender {
	// s.to = append(s.to, mail.Address{Address: email, Name: name})
	s.to = []mail.Address{{Address: email, Name: name}}
	return s
}

// SendEmail SendEmail
func (s *SMTPSender) SendEmail(subject string, data interface{}, attachment []string) error {
	if s.layoutTemplate == "" || s.bodyTemplate == "" {
		return errors.New("no template")
	}
	if len(s.to) == 0 {
		return errors.New("no recipient")
	}
	t := template.Must(template.ParseFiles(s.layoutTemplate, s.bodyTemplate))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		return err
	}

	s.body = buf.String()

	return s.send(subject, attachment)
}

// SendEmail SendEmail
func (s *SMTPSender) SendEmailWithTemplate(subject, message string, attachment []string) error {
	return s.send(subject, attachment)
}

func (s *SMTPSender) send(subject string, attachment []string) error {

	e := email.NewEmail()
	e.From = s.from.Address
	fmt.Println("FROM", e.From)
	for _, v := range s.to {
		e.To = append(e.To, v.String())
	}
	fmt.Println("TO", e.To)
	e.Subject = subject
	e.HTML = []byte(s.body)
	for _, v := range attachment {
		e.AttachFile(v)
	}
	var client *smtp.Client
	var auth smtp.Auth
	var err error
	if s.smtpPort == 587 || s.smtpPort == 2525 {
		_, err := s.sendEmailWithTLS(e)
		if err != nil {
			return err
		}
	} else {
		auth = unencryptedAuth{smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpServer)}

		client, err = smtp.Dial(fmt.Sprintf("%s:%d", s.smtpServer, s.smtpPort))
		if err != nil {
			return err
		}
		defer client.Close()

		// Authenticate with the server
		if err := client.Auth(auth); err != nil {
			log.Printf("ERROR #1 %v", err)
			return err
		}
		// Send the email message
		if err := e.Send(fmt.Sprintf("%s:%d", s.smtpServer, s.smtpPort), auth); err != nil {
			log.Printf("ERROR #2 %v", err)
			return err
		}
	}

	return nil
}

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

func (s *SMTPSender) sendEmailWithTLS(e *email.Email) (*smtp.Client, error) {
	// Konfigurasi SMTP server
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Hati-hati dengan ini, sebaiknya gunakan sertifikat yang valid
		ServerName:         s.smtpServer,
	}

	// Auth untuk SMTP
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpServer)

	// Header dan pesan email dalam format HTML
	// headers := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n", to, subject)
	// msg := headers + body

	// Koneksi TLS

	// Buat koneksi ke server SMTP
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.smtpServer, s.smtpPort))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat koneksi TLS: %v", err)
	}
	defer conn.Close()

	// Buat client SMTP
	client, err := smtp.NewClient(conn, s.smtpServer)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat client SMTP: %v", err)
	}
	defer client.Quit()

	if err = client.StartTLS(tlsConfig); err != nil {
		return nil, err
	}

	// Autentikasi
	if err := client.Auth(auth); err != nil {
		return nil, fmt.Errorf("gagal autentikasi: %v", err)
	}

	if err := e.Send(fmt.Sprintf("%s:%d", s.smtpServer, s.smtpPort), auth); err != nil {
		log.Printf("ERROR TLS #2 %v", err)
		return nil, err
	}

	fmt.Println("SEND EMAIL to", e.To)

	return client, nil

}
