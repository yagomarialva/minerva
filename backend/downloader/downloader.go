package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/minerva/backend/models"
	"github.com/minerva/backend/ws"
	"gorm.io/gorm"
)

type Downloader struct {
	db  *gorm.DB
	hub *ws.Hub
}

func NewDownloader(db *gorm.DB, hub *ws.Hub) *Downloader {
	return &Downloader{db: db, hub: hub}
}

// Start begins the background worker that monitors the queue
func (d *Downloader) Start() {
	for {
		var book models.Book
		// Find first Queued book
		if err := d.db.Where("status = ?", "Queued").First(&book).Error; err != nil {
			// No queued books, sleep and check again
			time.Sleep(3 * time.Second)
			continue
		}

		d.processDownload(&book)
	}
}

func (d *Downloader) processDownload(book *models.Book) {
	log.Printf("Starting download for: %s", book.Title)
	
	// Update status
	book.Status = "Downloading"
	d.db.Save(book)
	d.broadcastUpdate()

	err := d.downloadFile(book)
	if err != nil {
		log.Printf("Error downloading %d: %v", book.ID, err)
		book.Status = "Error"
		d.db.Save(book)
	} else {
		log.Printf("Download completed: %s", book.Title)
		book.Status = "Completed"
		d.db.Save(book)
	}
	d.broadcastUpdate()
}

func (d *Downloader) downloadFile(book *models.Book) error {
	// 1. Fetch ads.php
	resp, err := http.Get(book.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	html := string(bodyBytes)

	// 2. Extract get.php link
	re := regexp.MustCompile(`href="(get\.php\?[^"]+)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return fmt.Errorf("could not find download link in mirror page")
	}

	directURL := "https://libgen.li/" + matches[1]

	// 3. Download the actual file
	fileResp, err := http.Get(directURL)
	if err != nil {
		return err
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != 200 {
		return fmt.Errorf("failed to download file, status code: %d", fileResp.StatusCode)
	}

	// 4. Save to disk
	safeTitle := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(book.Title, "_")
	fileName := fmt.Sprintf("%d_%s.%s", book.ID, safeTitle, book.Extension)
	filePath := filepath.Join("/app/books", fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 5. Track progress and write
	buffer := make([]byte, 32*1024)
	var downloaded int64
	totalSize := fileResp.ContentLength

	// Fallback if Content-Length is missing
	if totalSize <= 0 {
		totalSize = 5 * 1024 * 1024 // assume 5MB for progress bar fallback
	}

	for {
		n, err := fileResp.Body.Read(buffer)
		if n > 0 {
			out.Write(buffer[:n])
			downloaded += int64(n)

			percent := int((float64(downloaded) / float64(totalSize)) * 100)
			if percent > 100 {
				percent = 100
			}

			// Broadcast progress
			d.hub.Broadcast <- ws.Message{
				Type: "PROGRESS",
				Payload: map[string]interface{}{
					"id":      book.ID,
					"percent": percent,
				},
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	book.FilePath = filePath
	return nil
}

func (d *Downloader) broadcastUpdate() {
	d.hub.Broadcast <- ws.Message{
		Type: "UPDATE_QUEUE",
		Payload: map[string]string{
			"status": "refresh",
		},
	}
}
