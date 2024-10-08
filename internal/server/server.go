package server

import (
	"log"
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
	logger  *logger.ZapLogger
}

func New(cfg *config.Config, logger *logger.ZapLogger) *server {
	r := redis.New(cfg.Redis)
	s := service.New(cfg, r, logger)
	logger.Debug("Redis and service entities are created")
	return &server{
		config:  cfg.Server,
		echo:    echo.New(),
		service: s,
		logger:  logger,
	}
}

func Start() error {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("Failed to load config: " + err.Error())
		return err
	}
	logger := logger.New(cfg.Logger)

	logger.Info("Config loaded")

	lis, err := net.Listen("tcp", cfg.GRPC.Host+":"+cfg.GRPC.Port)
	if err != nil {
		logger.Error("Failed to listen: " + err.Error())
		return err
	}
	logger.Info("Start listen tcp on port: ")
	s := grpc.NewServer()

	srv := New(cfg, logger)
	srv.configureRouter()

	go srv.echo.Start(srv.config.Host + ":" + srv.config.Port)

	logger.Info("HTTP server started")

	gen.RegisterLinkServiceServer(s, srv.service)
	logger.Info("Service registered and started, waiting for connections...")
	return s.Serve(lis)
}

func (s *server) configureRouter() {

	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())

	s.echo.Static("/", "/root/static")

	// handler to return html page to user
	s.echo.GET("/gen/:user_id/:url", s.serveLink)
	s.echo.GET("/", s.mainHandler)
}
