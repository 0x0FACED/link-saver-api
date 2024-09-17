package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (s *LinkService) GetResourceContentByNameType(ctx context.Context, name string, resType models.ResourceType) ([]byte, error) {
	return s.db.GetResourceContentByNameType(ctx, name, resType)
}

func (s *LinkService) GetURLFromRedis(ctx context.Context, userID int64, generatedURL string) (string, error) {
	return s.redis.GetOriginalURL(ctx, userID, generatedURL)
}

func hash(userID int64, url string) string {
	data := fmt.Sprintf("%d:%s:%d", userID, url, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func hashResName(name string, resType models.ResourceType) string {
	data := fmt.Sprintf("%s:%d", name, resType)
	hash := sha256.Sum256([]byte(data))
	hashString := hex.EncodeToString(hash[:])

	var extension string
	switch resType {
	case models.ScriptType:
		extension = "js"
	case models.CSSType:
		extension = "css"
	case models.ImageType:
		extension = getImageExtension(name)
	default:
		extension = "png"
	}

	return fmt.Sprintf("%s.%s", hashString, extension)
}

func getImageExtension(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return "unknown"
	}
	return ext[1:]
}

func (s *LinkService) saveLink(ctx context.Context, link *models.Link) error {
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

func getResourceURL(baseURL string, resType string, name string) string {
	// http://localhost:8000/assets/css/some_style.css
	return fmt.Sprintf("%s/s/assets/%s/%s", baseURL, resType, name)
}
func (s *LinkService) saveResource(res *models.Resource) error {
	err := s.db.SaveResource(context.TODO(), res)
	if err != nil {
		s.logger.Error("Failed to save resource", zap.Error(err))
		return err
	}

	return nil
}

func (s *LinkService) fetchResourceContent(url string, e *colly.HTMLElement) []byte {
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Error("Failed to fetch resource", zap.String("url", url), zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read resource content", zap.String("url", url), zap.Error(err))
		return nil
	}

	return content
}

func getRelativePath(u string) string {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(parsedURL.Path, "/")
}
