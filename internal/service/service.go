package service

import (
	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/cached/redis"
	"github.com/0x0FACED/link-saver-api/internal/logger"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	"github.com/0x0FACED/link-saver-api/internal/storage/postgres"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"github.com/gocolly/colly"
	"go.uber.org/zap"
)

var pkg = "service"

type LinkService struct {
	gen.UnimplementedLinkServiceServer

	db     storage.Database
	redis  *redis.Redis
	Logger *logger.ZapLogger
	colly  *colly.Collector
}

func New(cfg config.DatabaseConfig, redis *redis.Redis, logger *logger.ZapLogger) *LinkService {
	logger.Debug("Database config: ",
		zap.String("db_name", cfg.Name),
		zap.String("db_host", cfg.Host),
		zap.String("db_port", cfg.Port),
		zap.String("db_username", cfg.Username),
		zap.String("db_password", cfg.Password),
		zap.String("db_driver", cfg.Driver),
	)

	// TODO: add more drivers
	var db storage.Database
	switch cfg.Driver {
	case "postgres":
		db = postgres.New(cfg)
	default:
		db = postgres.New(cfg)
	}

	if db.Connect() == storage.ErrConnectDB {
		panic("cant connect db, panic!")
	}

	logger.Info("Successfully connected to database")

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowURLRevisit(),
	)

	logger.Info("Created colly instance")

	return &LinkService{
		db:     db,
		redis:  redis,
		Logger: logger,
		colly:  c,
	}
}
