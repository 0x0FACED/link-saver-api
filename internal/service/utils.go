package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/wrap"
	"github.com/gocolly/colly"
	"go.uber.org/zap"
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
		s.logger.Error("Error while saving to db", zap.Error(err))
		return wrap.E(pkg, "failed to SaveLink()", err)
	}

	return nil
}

func getFullLink(baseURL string, userID int64, generatedURL string) string {
	return fmt.Sprintf("%s/gen/%d/%s", baseURL, userID, generatedURL)
}
func (s *LinkService) saveResource(content []byte, resourcePath string) string {
	safeFileName := strings.ReplaceAll(resourcePath, "/", "_")

	resourceDir := "./resources"

	err := os.MkdirAll(resourceDir, os.ModePerm)
	if err != nil {
		s.logger.Error("Failed to create resource directory", zap.Error(err))
		return ""
	}

	filePath := filepath.Join(resourceDir, safeFileName)

	err = os.WriteFile(filePath, content, 0644)
	if err != nil {
		s.logger.Error("Failed to write resource to file", zap.Error(err))
		return ""
	}

	s.logger.Info("Resource saved", zap.String("file_path", filePath))

	return filePath
}

// Метод для загрузки ресурса по URL
func (s *LinkService) fetchResourceContent(url string, e *colly.HTMLElement) []byte {
	var content []byte
	err := s.colly.Visit(url)
	if err == nil {
		content = e.Response.Body
	} else {
		s.logger.Error("Failed to fetch resource", zap.String("url", url), zap.Error(err))
	}
	return content
}

func getRelativePath(u string) string {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return ""
	}
	// Join path to remove leading slashes if they are present
	return strings.TrimPrefix(parsedURL.Path, "/")
}
