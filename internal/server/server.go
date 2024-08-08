package server

import (
	"github.com/labstack/echo/v4"
)

type server struct {
	r *echo.Echo
}

func New() *server {

	return &server{
		r: echo.New(),
	}
}

func (s *server) Get() {

}
