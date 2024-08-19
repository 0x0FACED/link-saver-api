package service

import (
	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/cached/redis"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	"github.com/0x0FACED/link-saver-api/internal/storage/postgres"
	"github.com/0x0FACED/proto-files/link_service/gen"
)

type LinkService struct {
	gen.UnimplementedLinkServiceServer

	db    storage.Database
	redis *redis.Redis
}

func New(cfg config.DatabaseConfig, redis *redis.Redis) *LinkService {
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

	return &LinkService{
		db:    db,
		redis: redis,
	}
}
