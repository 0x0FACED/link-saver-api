package server

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
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

		return ctx.Redirect(302, "/")
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

func (s *server) serveResourceHandler(ctx echo.Context) error {
	_typeParam := ctx.Param("type")
	name := ctx.Param("name")

	var resType models.ResourceType
	switch _typeParam {
	case "script":
		resType = models.ScriptType
	case "css":
		resType = models.CSSType
	case "image":
		resType = models.ImageType
	default:
		s.logger.Debug("Unknown type", zap.String("req", ctx.Request().RequestURI))
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown resource type")
	}

	content, err := s.service.GetResourceContentByNameType(context.TODO(), name, resType)
	if err != nil {
		s.logger.Error("Cant get resource content from DB", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown resource type")
	}

	var contentType string
	switch resType {
	case models.ScriptType:
		contentType = "application/javascript"
	case models.CSSType:
		contentType = "text/css"
	case models.ImageType:
		ext := getImageExtension(name)
		switch ext {
		case "png":
			contentType = "image/png"
		case "jpeg":
			contentType = "image/jpeg"
		case "gif":
			contentType = "image/gif"
		default:
			contentType = "application/octet-stream"
		}
	}

	return ctx.Blob(http.StatusOK, contentType, content)
}

func getImageExtension(name string) string {
	ext := filepath.Ext(name)
	switch ext {
	case ".png":
		return "png"
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".gif":
		return "gif"
	default:
		return "unknown"
	}
}

func (s *server) mainHandler(ctx echo.Context) error {
	return ctx.File("/root/static/index.html")
}
