package server

import (
	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	"github.com/0x0FACED/link-saver-api/internal/storage/postgres"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type server struct {
	config config.ServerConfig
	db     storage.Database
	gen.UnimplementedLinkServiceServer

	echo *echo.Echo
}

func New(cfg *config.Config) *server {

	// TODO: add more drivers
	var db storage.Database
	switch cfg.Database.Driver {
	case "postgres":
		db = &postgres.Postgres{Config: cfg.Database}
	default:
		db = &postgres.Postgres{Config: cfg.Database}
	}

	if db.Connect() == storage.ErrConnectDB {
		panic("cant connect db, panic!")
	}

	return &server{
		config: cfg.Server,
		echo:   configureRouter(),
		db:     db,
	}
}

func configureRouter() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return e
}
