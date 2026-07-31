package kindle

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/minerva/backend/models"
	"github.com/minerva/backend/ws"
	"gopkg.in/gomail.v2"
)

// Sender handles emailing books to Kindle
type Sender struct {
	hub *ws.Hub
}

func NewSender(hub *ws.Hub) *Sender {
	return &Sender{
		hub: hub,
	}
}

// SendBookToKindle sends the downloaded book file to the configured Kindle email.
func (s *Sender) SendBookToKindle(book *models.Book, settings *models.Settings) error {
	if settings.KindleEmail == "" {
		return fmt.Errorf("kindle email not configured")
	}
	if settings.SMTPServer == "" || settings.SMTPUser == "" {
		return fmt.Errorf("SMTP settings not configured")
	}

	log.Printf("Preparing to send '%s' to %s", book.Title, settings.KindleEmail)
	
	s.hub.Broadcast <- ws.Message{
		Type: "KINDLE_SEND_START",
		Payload: map[string]interface{}{
			"bookId": book.ID,
			"title":  book.Title,
			"status": "Sending to Kindle...",
		},
	}

	m := gomail.NewMessage()
	m.SetHeader("From", settings.SMTPFromEmail)
	m.SetHeader("To", settings.KindleEmail)
	m.SetHeader("Subject", fmt.Sprintf("Send to Kindle: %s", book.Title))
	m.SetBody("text/plain", "Your book, sent via MINERVA.")
	
	// Attach the book file
	m.Attach(book.FilePath)

	d := gomail.NewDialer(settings.SMTPServer, settings.SMTPPort, settings.SMTPUser, settings.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // Ensure compatibility with self-signed or older SMTP configs

	err := d.DialAndSend(m)
	if err != nil {
		s.hub.Broadcast <- ws.Message{
			Type: "KINDLE_SEND_ERROR",
			Payload: map[string]interface{}{
				"bookId": book.ID,
				"error":  err.Error(),
			},
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Successfully sent '%s' to Kindle", book.Title)
	
	s.hub.Broadcast <- ws.Message{
		Type: "KINDLE_SEND_SUCCESS",
		Payload: map[string]interface{}{
			"bookId": book.ID,
			"title":  book.Title,
			"status": "Sent to Kindle",
		},
	}

	return nil
}
