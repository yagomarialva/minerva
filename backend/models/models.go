package models

import (
	"time"

	"gorm.io/gorm"
)

// Book represents an e-book in the Minerva library
type Book struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"index" json:"title"`
	Author      string         `gorm:"index" json:"author"`
	Year        int            `json:"year"`
	ISBN        string         `json:"isbn"`
	CoverURL    string         `json:"coverUrl"` // Path to local file or remote URL
	Format      string         `json:"format"`   // e.g., "EPUB", "PDF"
	FilePath    string         `json:"filePath"` // Path in /app/books
	FileSize    int64          `json:"fileSize"`
	FilesizeStr string         `json:"filesize"` // String format (e.g. "3 MB") for UI
	Extension   string         `json:"extension"`
	Language    string         `json:"language"`
	DownloadURL string         `json:"downloadUrl"`
	Status      string         `json:"status"`   // "Queued", "Downloading", "Completed", "Error"
	Description string         `json:"description"`
}

// Settings represents system and user configuration for Minerva
type Settings struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	KindleEmail    string    `json:"kindleEmail"` // The destination @kindle.com email
	SMTPServer     string    `json:"smtpServer"`
	SMTPPort       int       `json:"smtpPort"`
	SMTPUser       string    `json:"smtpUser"`
	SMTPPassword   string    `json:"smtpPassword"` // Should ideally be encrypted
	SMTPFromEmail  string    `json:"smtpFromEmail"`
	MaxConcurrency int       `json:"maxConcurrency"` // Max simultaneous downloads (default: 3)
}
