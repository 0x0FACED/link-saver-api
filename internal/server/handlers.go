package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *server) handleRoutes() {
	s.r.POST("/api/link", s.saveLinkHandler)
	s.r.GET("/api/link/:id", s.getLinkHandler)
	s.r.GET("/api/link/:username", s.getLinksHandler)
	s.r.POST("/api/link/:id", s.deleteLinkHandler)
}

func (s *server) saveLinkHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "")
}

func (s *server) deleteLinkHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "")
}

func (s *server) getLinkHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "")
}

func (s *server) getLinksHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "")
}
