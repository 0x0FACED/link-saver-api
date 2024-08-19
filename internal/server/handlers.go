package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *server) serveLink(ctx echo.Context) error {
	//u := ctx.Param("username")
	//url := ctx.Param("url")

	// Check if link exists in Redis
	// if true -> return c.HTML with page from redis
	// else -> link expired or didnt exist, try to gen new one
	return ctx.HTML(http.StatusOK, "")
}
