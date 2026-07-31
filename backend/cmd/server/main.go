package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/minerva/backend/models"
	"github.com/minerva/backend/scraper"
	"github.com/minerva/backend/downloader"
	"github.com/minerva/backend/ws"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Minerva Backend...")

	// Database Setup
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "minerva.db"
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.Book{}, &models.Settings{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Init WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Init Downloader
	dl := downloader.NewDownloader(db, hub)
	go dl.Start()

	// Setup HTTP Routes
	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r)
	})

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		// Initialize scraper
		engine := scraper.NewSearchEngine()
		
		// For now, only search direct downloads (Anna's Archive etc)
		results, err := engine.SearchDirectDownloads(r.Context(), query)
		if err != nil {
			log.Printf("Search error: %v", err)
			http.Error(w, "Error executing search", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			log.Printf("JSON encode error: %v", err)
		}
	})

	http.HandleFunc("/api/library", func(w http.ResponseWriter, r *http.Request) {
		var books []models.Book
		db.Where("status = ?", "Completed").Find(&books)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	})

	http.HandleFunc("/api/queue", func(w http.ResponseWriter, r *http.Request) {
		var books []models.Book
		db.Where("status IN ?", []string{"Queued", "Downloading", "Error"}).Find(&books)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	})

	http.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var book models.Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		book.Status = "Queued"
		if err := db.Create(&book).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Broadcast update
		hub.Broadcast <- ws.Message{
			Type: "UPDATE_QUEUE",
			Payload: map[string]string{"status": "refresh"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	})

	http.HandleFunc("/api/download-file", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		var book models.Book
		if err := db.First(&book, id).Error; err != nil {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		if book.FilePath == "" {
			http.Error(w, "File not available", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(book.FilePath))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, book.FilePath)
	})

	http.HandleFunc("/api/convert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		var book models.Book
		if err := db.First(&book, id).Error; err != nil {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		if book.Extension == "EPUB" || book.FilePath == "" {
			http.Error(w, "Invalid file for conversion", http.StatusBadRequest)
			return
		}

		// Run in background
		go func(b models.Book) {
			oldPath := b.FilePath
			newPath := strings.TrimSuffix(oldPath, filepath.Ext(oldPath)) + ".epub"
			
			log.Printf("Converting %s to EPUB...", oldPath)
			cmd := exec.Command("ebook-convert", oldPath, newPath)
			if err := cmd.Run(); err != nil {
				log.Printf("Conversion failed for %d: %v", b.ID, err)
				return
			}
			
			// Success
			log.Printf("Conversion successful: %s", newPath)
			b.FilePath = newPath
			b.Extension = "EPUB"
			db.Save(&b)
			os.Remove(oldPath)

			// Tell frontend to refresh
			hub.Broadcast <- ws.Message{
				Type: "UPDATE_QUEUE",
				Payload: map[string]string{"status": "refresh"},
			}
		}(book)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "converting"})
	})

	http.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		var settings models.Settings
		
		if r.Method == http.MethodGet {
			db.FirstOrCreate(&settings, models.Settings{ID: 1})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(settings)
		} else if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			settings.ID = 1 // Ensure single row
			db.Save(&settings)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(settings)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
