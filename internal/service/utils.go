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

func (s *LinkService) GetContentFromDatabase(ctx context.Context, userID int64, originalURL string) ([]byte, error) {
	return s.db.GetContentByTelegramIDOriginalURL(ctx, userID, originalURL)
}

func (s *LinkService) GetURLFromRedis(ctx context.Context, userID int64, generatedURL string) (string, error) {
	return s.redis.GetOriginalURL(ctx, userID, generatedURL)
}

func hash(userID int64, url string) string {
	data := fmt.Sprintf("%d:%s:%d", userID, url, time.Now().UnixNano())
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

func getFullLink(userID int64, generatedURL string) string {
	return fmt.Sprintf("http://localhost:8000/gen/%d/%s", userID, generatedURL)
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
