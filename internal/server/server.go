package server

import (
	"errors"
	"net"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/cached/redis"
	"github.com/0x0FACED/link-saver-api/internal/service"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

type server struct {
	config  config.ServerConfig
	service *service.LinkService
	echo    *echo.Echo
}

func New(cfg *config.Config) *server {
	r := redis.New(cfg.Redis)
	s := service.New(cfg.Database, r)
	return &server{
		config:  cfg.Server,
		echo:    echo.New(),
		service: s,
	}
}

func Start() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return errors.New("Failed to listen: " + err.Error())
	}
	s := grpc.NewServer()
	cfg, err := config.Load()
	if err != nil {
		return errors.New("Failed to load config: " + err.Error())
	}
	srv := New(cfg)
	srv.configureRouter()
	gen.RegisterLinkServiceServer(s, srv.service)
	if err := s.Serve(lis); err != nil {
		return errors.New("Failed to serve: " + err.Error())
	}

	return nil
}

func (s *server) configureRouter() {

	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())

	// handler to return html page to user
	s.echo.GET("/gen/:username/:url", s.serveLink)
}
