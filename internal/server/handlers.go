package server

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *server) serveLink(ctx echo.Context) error {
	u := ctx.Param("username")
	url := ctx.Param("url")
	s.service.Logger.Debug("Received serveLink() request with params",
		zap.String("user", u),
		zap.String("gen_url", url),
	)
	original, err := s.service.GetURLFromRedis(context.TODO(), u, url)
	if err != nil {
		s.service.Logger.Error("Error GetURLFromRedis()",
			zap.Error(err),
		)
	}

	s.service.Logger.Debug("Original URL from Redis",
		zap.String("original_url", original),
	)

	content, err := s.service.GetContentFromDatabase(context.TODO(), u, original)
	if err != nil {
		s.service.Logger.Error("Error GetContentFromDatabase()",
			zap.Error(err),
		)
	}
	// Check if link exists in Redis
	// if true -> return c.HTML with page from redis
	// else -> link expired or didnt exist, try to gen new one
	return ctx.HTML(http.StatusOK, string(content[:]))
}
