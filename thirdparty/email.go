package thirdparty

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
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
func (s *SMTPSender) SendEmailWithTemplate(subject string, attachment []string) error {
	return s.send(subject, attachment)
}

func (s *SMTPSender) send(subject string, attachment []string) error {

	e := email.NewEmail()
	e.From = s.from.String()
	// for _, v := range s.to {
	// 	e.To = append(e.To, v.String())
	// }
	e.To = []string{s.to[0].Address}
	e.Subject = subject
	e.HTML = []byte(s.body)
	for _, v := range attachment {
		e.AttachFile(v)
	}

	auth := unencryptedAuth{smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpServer)}

	// Connect to the SMTP server
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", s.smtpServer, s.smtpPort))
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
