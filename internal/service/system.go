package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"github.com/gocolly/colly"
)

func (s *LinkService) SaveLink(ctx context.Context, req *gen.SaveLinkRequest) (*gen.SaveLinkResponse, error) {
	log.Printf("Received link: %s with description: %s by user: %s", req.OriginalUrl, req.Description, req.Username)
	// create collector
	c := colly.NewCollector()

	// onHTML -> html page, save to page.html
	c.OnHTML("html", func(e *colly.HTMLElement) {
		link := &models.Link{
			OriginalURL: req.OriginalUrl,
			UserName:    req.Username,
			Description: req.Description,
			Content:     []byte(e.Response.Body),
		}
		// save page as bytea to database and to redis with exp 24h
		err := s.save(context.TODO(), link)
		if err != nil {
			log.Println("OnHTML err: ", err)
			return
		}
		log.Printf("HTML page saved with name: %v", req.Description)
	})

	// onHTML -> image, save to imgURL.png/jpg file
	/*c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		imgURL := resolveURL(e.Attr("src"), e.Request.URL)
		err := downloadFile(outputDir, imgURL)
		if err != nil {
			log.Printf("Error saving image: %v", err)
		}
	})*/

	// onHTML -> stylesheet (CSS), save to css file
	/*c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		cssURL := resolveURL(e.Attr("href"), e.Request.URL)
		err := downloadFile(outputDir, cssURL)
		if err != nil {
			log.Printf("Error saving CSS: %v", err)
		}
	})*/

	// onHTML -> script (js), save to js file
	/*c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		jsURL := resolveURL(e.Attr("src"), e.Request.URL)
		err := downloadFile(outputDir, jsURL)
		if err != nil {
			log.Printf("Error saving JavaScript: %v", err)
		}
	})*/

	// start parse html page
	err := c.Visit(req.OriginalUrl)
	if err != nil {
		return &gen.SaveLinkResponse{Success: false}, err
	}

	log.Println("Finished!")
	return &gen.SaveLinkResponse{Success: true}, nil
}

func hash(username, url string) string {
	data := fmt.Sprintf("%s:%s:%d", username, url, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *LinkService) DeleteLink(ctx context.Context, req *gen.DeleteLinkRequest) (*gen.DeleteLinkResponse, error) {
	// TODO: impl
	return &gen.DeleteLinkResponse{Success: true}, nil
}

func (s *LinkService) GetLinks(ctx context.Context, req *gen.GetLinksRequest) (*gen.GetLinksResponse, error) {
	return &gen.GetLinksResponse{Links: []*gen.Link{}}, nil
}

func (s *LinkService) save(ctx context.Context, link *models.Link) error {
	err := s.db.SaveLink(ctx, link)
	if err != nil {
		return err
	}
	gen := hash(link.UserName, link.OriginalURL)
	log.Println("generated link: ", gen)
	err = s.redis.SaveLink(ctx, link.UserName, gen)
	if err != nil {
		// TODO: maybe delete from db if cant save to redis
		return err
	}

	log.Println("Successfully saved")
	return nil
}

func getFileNameFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "unknown"
	}

	fileName := strings.ReplaceAll(parsedURL.Path, "/", "_")
	if fileName == "" {
		fileName = "unknown.html"
	}

	return fileName
}

func resolveURL(relURL string, baseURL *url.URL) string {
	parsedURL, err := url.Parse(relURL)
	if err != nil {
		return relURL
	}
	return baseURL.ResolveReference(parsedURL).String()
}

func downloadFile(outputDir, urlStr string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fileName := getFileNameFromURL(urlStr)
	filePath := filepath.Join(outputDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Saved file: %s", filePath)
	return nil
}
