package scraper

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/minerva/backend/models"
)

// TorrentClient manages torrent downloads as a fallback
type TorrentClient struct {
	client   *torrent.Client
	booksDir string
}

func NewTorrentClient(booksDir string) (*TorrentClient, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = booksDir
	// Optimizations to avoid stalling
	cfg.NoDHT = false
	cfg.Seed = false

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent client: %w", err)
	}

	return &TorrentClient{
		client:   client,
		booksDir: booksDir,
	}, nil
}

// SearchTrackers searches trackers/indices for books.
// Typically done via a tracker API (like Jackett or a public tracker search).
func (t *TorrentClient) SearchTrackers(query string) ([]string, error) {
	// Placeholder: This would integrate with Jackett/Prowlarr or parse public trackers
	// and return a list of magnet links.
	return []string{
		// "magnet:?xt=urn:btih:EXAMPLE",
	}, nil
}

// DownloadFromMagnet starts downloading an EPUB/PDF from a magnet link.
func (t *TorrentClient) DownloadFromMagnet(magnetURI string, book *models.Book, onComplete func(filePath string)) error {
	t2, err := t.client.AddMagnet(magnetURI)
	if err != nil {
		return fmt.Errorf("failed to add magnet: %w", err)
	}

	log.Printf("Fetching torrent metadata for: %s", book.Title)
	<-t2.GotInfo() // Blocks until metadata is fetched

	var targetFile *torrent.File

	// Identify the EPUB or PDF inside the torrent
	for _, file := range t2.Files() {
		ext := strings.ToLower(filepath.Ext(file.Path()))
		if ext == ".epub" || ext == ".pdf" {
			targetFile = file
			targetFile.Download()
			log.Printf("Found targeted file %s in torrent, starting download.", targetFile.Path())
			break // Priority on the first match. For collections, we'd need more logic.
		}
	}

	if targetFile == nil {
		return fmt.Errorf("no EPUB or PDF found in torrent")
	}

	// Goroutine to monitor download progress asynchronously
	go func() {
		// Wait for this specific file or the entire torrent to finish.
		// WaitAll() waits for all pieces that have Download() called on them.
		t.client.WaitAll() 
		log.Printf("Torrent download completed for: %s", targetFile.Path())
		book.Status = "Completed"
		
		fullPath := filepath.Join(t.booksDir, targetFile.Path())
		
		if onComplete != nil {
			onComplete(fullPath)
		}
	}()

	return nil
}

func (t *TorrentClient) Close() {
	if t.client != nil {
		t.client.Close()
	}
}
