package server

import (
	"net"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/cached/redis"
	"github.com/0x0FACED/link-saver-api/internal/logger"
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

func New(cfg *config.Config, logger *logger.ZapLogger) *server {
	r := redis.New(cfg.Redis)
	s := service.New(cfg.Database, r, logger)
	logger.Debug("Redis and service entities are created")
	return &server{
		config:  cfg.Server,
		echo:    echo.New(),
		service: s,
	}
}

func Start() error {
	logger := logger.New()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("Failed to listen: " + err.Error())
		return err
	}
	logger.Info("Start listen tcp on 50051")
	s := grpc.NewServer()
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config: " + err.Error())
		return err
	}

	logger.Info("Config loaded")
	srv := New(cfg, logger)
	srv.configureRouter()
	go srv.echo.Start(srv.config.Host + ":" + srv.config.Port)
	logger.Info("Server created and router configured")
	gen.RegisterLinkServiceServer(s, srv.service)
	logger.Info("Service registered and started, waiting for connections...")
	if err := s.Serve(lis); err != nil {
		logger.Error("Failed to serve: " + err.Error())
		return err
	}
	logger.Info("Finished")
	return nil
}

func (s *server) configureRouter() {

	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())

	// handler to return html page to user
	s.echo.GET("/gen/:user_id/:url", s.serveLink)
}
