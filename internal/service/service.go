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
	logger *logger.ZapLogger
	colly  *colly.Collector
	cfg    config.GRPCConfig
}

func New(cfg *config.Config, redis *redis.Redis, logger *logger.ZapLogger) *LinkService {
	logger.Debug("Database config: ",
		zap.String("db_name", cfg.Database.Name),
		zap.String("db_host", cfg.Database.Host),
		zap.String("db_port", cfg.Database.Port),
		zap.String("db_username", cfg.Database.Username),
		zap.String("db_password", cfg.Database.Password),
		zap.String("db_driver", cfg.Database.Driver),
	)

	// TODO: add more drivers
	var db storage.Database
	switch cfg.Database.Driver {
	case "postgres":
		db = postgres.New(cfg.Database)
	default:
		db = postgres.New(cfg.Database)
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
		logger: logger,
		colly:  c,
		cfg:    cfg.GRPC,
	}
}
