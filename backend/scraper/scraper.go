package scraper

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/minerva/backend/models"
)

// SearchEngine represents the web scraper to find direct downloads
type SearchEngine struct {
	// Add config for target libraries
}

func NewSearchEngine() *SearchEngine {
	return &SearchEngine{}
}

// SearchDirectDownloads uses headless Chromium to search for books on open repositories.
// It bypasses basic Cloudflare protections using chromedp options.
func (s *SearchEngine) SearchDirectDownloads(ctx context.Context, query string) ([]models.Book, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	chromeCtx, cancelChrome := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancelChrome()

	// Increase timeout for possible CF challenges and slow page load
	timeoutCtx, cancelTimeout := context.WithTimeout(chromeCtx, 45*time.Second)
	defer cancelTimeout()

	// Target LibGen.li since libgen.is is timing out
	searchURL := fmt.Sprintf("https://libgen.li/index.php?req=%s", url.QueryEscape(query))
	
	type ScrapedBook struct {
		Author    string `json:"author"`
		Title     string `json:"title"`
		Language  string `json:"language"`
		Filesize  string `json:"filesize"`
		Extension string `json:"extension"`
		Link      string `json:"link"`
	}
	var scraped []ScrapedBook

	js := `
		Array.from(document.querySelectorAll('table#tablelibgen tr')).slice(1).map(tr => {
			const tds = tr.querySelectorAll('td');
			if(tds.length < 9) return null;
			
			const titleNode = tds[0].querySelector('a[href*="edition.php"]');
			const title = titleNode ? titleNode.innerText.replace(/<[^>]*>?/gm, '').trim() : '';
			const author = tds[1].innerText.trim();
			const language = tds[4].innerText.trim();
			const size = tds[6].innerText.trim();
			const ext = tds[7].innerText.trim();
			const link = tds[8].querySelector('a') ? tds[8].querySelector('a').href : '';
			
			return {
				author: author,
				title: title,
				language: language,
				filesize: size,
				extension: ext.toUpperCase(),
				link: link,
			};
		}).filter(b => b && b.title && ['PDF', 'EPUB', 'CBR', 'CBZ', 'MOBI'].includes(b.extension))
	`

	log.Printf("Navigating to: %s", searchURL)
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(searchURL),
		// Wait for the table to render
		chromedp.WaitVisible(`table#tablelibgen`, chromedp.ByQuery),
		chromedp.Evaluate(js, &scraped),
	)

	if err != nil {
		log.Printf("Scraping error: %v", err)
		return nil, err
	}

	var results []models.Book
	for _, s := range scraped {
		results = append(results, models.Book{
			Title:       s.Title,
			Author:      s.Author,
			Extension:   s.Extension,
			FilesizeStr: s.Filesize,
			Language:    s.Language,
			DownloadURL: s.Link,
			Status:      "Found",
		})
	}

	return results, nil
}
