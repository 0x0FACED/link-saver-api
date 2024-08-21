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
)

func (s *LinkService) GetContentFromDatabase(ctx context.Context, username, originalURL string) ([]byte, error) {
	return s.db.GetContentByUsernameOriginalURL(ctx, username, originalURL)
}

func (s *LinkService) GetURLFromRedis(ctx context.Context, username, generatedURL string) (string, error) {
	return s.redis.GetOriginalURL(ctx, username, generatedURL)
}

func hash(username, url string) string {
	data := fmt.Sprintf("%s:%s:%d", username, url, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *LinkService) saveToDatabase(ctx context.Context, link *models.Link) error {
	err := s.db.SaveLink(ctx, link)
	if err != nil {
		s.Logger.Error("Error while saving to db: " + err.Error())
		return err
	}
	//gen := hash(link.UserName, link.OriginalURL)
	//log.Println("generated link: ", gen)
	//err = s.redis.SaveLink(ctx, link.UserName, gen)
	//if err != nil {
	// TODO: maybe delete from db if cant save to redis
	//	return err
	//}

	return nil
}

func getFullLink(username, generatedURL string) string {
	return "http://localhost:8000" + "/gen" + "/" + username + "/" + generatedURL
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
