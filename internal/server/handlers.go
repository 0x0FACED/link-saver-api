package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *server) serveLink(ctx echo.Context) error {
	u := ctx.Param("user_id")
	userID, _ := strconv.ParseInt(u, 10, 64)
	url := ctx.Param("url")
	s.logger.Debug("Received serveLink() request with params",
		zap.String("user", u),
		zap.String("gen_url", url),
	)
	original, err := s.service.GetURLFromRedis(context.TODO(), userID, url)
	if err != nil {
		s.logger.Error("Error GetURLFromRedis()",
			zap.Error(err),
		)

		return ctx.HTML(http.StatusInternalServerError, "internal server error")
	}

	s.logger.Debug("Original URL from Redis",
		zap.String("original_url", original),
	)

	content, err := s.service.GetContentFromDatabase(context.TODO(), userID, original)
	if err != nil {
		s.logger.Error("Error GetContentFromDatabase()",
			zap.Error(err),
		)

		return ctx.HTML(http.StatusNotFound, "content not found in database")
	}

	return ctx.HTML(http.StatusOK, string(content[:]))
}
