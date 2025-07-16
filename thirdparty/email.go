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
	"os"

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
	Tls            bool
}

// NewSMTPSender creates a new instance of SMTPSender with the specified SMTP server details and sender address.
// It initializes the SMTP sender with the provided server, port, username, password, and from address.
// The recipient list is initialized as empty and can be set using methods like SetAddress.
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

// SetTls sets the TLS setting for the SMTP sender.
// If tls is true, it will use a TLS connection when sending the email.
// If tls is false, it will use a non-TLS connection.
func (s *SMTPSender) SetTls(tls bool) {
	fmt.Printf(`
	init email With TLS
	servername: %s
	port: %d
	tls: %t
	username : %s
	password: %s
	`, s.smtpServer, s.smtpPort, tls, s.smtpUsername, s.smtpPassword)
	s.Tls = tls
}

// SetTemplate sets the template files for the SMTP sender.
// It sets the layout and body templates for the email message.
// If the file does not exist, it will use an empty template.
// The method returns the SMTPSender instance for method chaining.
func (s *SMTPSender) SetTemplate(layout string, template string) *SMTPSender {
	if _, err := os.Stat(layout); os.IsNotExist(err) {
		log.Printf("warning: layout template %s not found. using empty template", layout)
	}
	if _, err := os.Stat(template); os.IsNotExist(err) {
		log.Printf("warning: body template %s not found. using empty template", template)
	}

	s.layoutTemplate = layout
	s.bodyTemplate = template
	return s
}

// SetAddress sets the recipient address for the SMTP sender.
// It takes the recipient's name and email address and sets it as the to address.
// It returns the SMTPSender instance for method chaining.
func (s *SMTPSender) SetAddress(name string, email string) *SMTPSender {
	// s.to = append(s.to, mail.Address{Address: email, Name: name})
	s.to = []mail.Address{{Address: email, Name: name}}
	return s
}

// SendEmail sends an email using the SMTP sender.
//
// It takes a subject string, a data object to be passed to the template, and a slice of strings representing the paths to the files to attach.
//
// It returns an error if something goes wrong.
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

// SendEmailWithTemplate sends an email with the given subject, message, and attachment.
//
// The message must be a complete HTML email message.
//
// It returns an error if something goes wrong.
func (s *SMTPSender) SendEmailWithTemplate(subject, message string, attachment []string) error {
	return s.send(subject, attachment)
}

// send sends an email using the SMTP sender.
//
// It takes a subject string and a slice of strings representing the paths to the files to attach.
//
// It returns an error if something goes wrong.
//
// The method implements the logic for sending an email using the email package.
// It creates a new email message, sets the from address, the recipients, the subject, the body, and the attachments.
// It then sends the email message using the email package.
// If the email server uses TLS, it uses the sendEmailWithTLS method to send the email.
// Otherwise, it uses the email package to send the email.
func (s *SMTPSender) send(subject string, attachment []string) error {

	e := email.NewEmail()
	e.From = s.from.Address
	fmt.Println("FROM", e.From)
	log.Println("FROM", e.From)
	for _, v := range s.to {
		e.To = append(e.To, v.Address)
	}
	fmt.Println("TO", e.To)
	log.Println("TO", e.To)
	e.Subject = subject
	e.HTML = []byte(s.body)
	for _, v := range attachment {
		e.AttachFile(v)
	}
	var client *smtp.Client
	var auth smtp.Auth
	var err error
	if s.smtpPort == 587 || s.smtpPort == 2525 || s.Tls {
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

// Start implements the Start method of the smtp.Auth interface.
//
// It takes a server *smtp.ServerInfo and returns a string, a slice of bytes, and an error.
//
// It modifies the server to use TLS and then calls the Start method of the underlying Auth.
func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

// sendEmailWithTLS sends an email using a TLS connection.
//
// This function establishes a secure connection to the SMTP server using TLS,
// authenticates the client, and sends the email. It takes an *email.Email
// object as input and returns an *smtp.Client and an error if any issues
// occur during the process.
//
// The TLS configuration currently skips certificate verification, which is
// not recommended for production environments. Ensure that a valid certificate
// is used to avoid potential security risks.
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
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", s.smtpServer, s.smtpPort))
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
