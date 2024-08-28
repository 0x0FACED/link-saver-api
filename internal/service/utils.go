package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/wrap"
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

func getFullLink(userID int64, generatedURL string) string {
	return fmt.Sprintf("http://localhost:8000/gen/%d/%s", userID, generatedURL)
}
