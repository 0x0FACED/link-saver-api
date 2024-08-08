package storage

import (
	"errors"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
)

var (
	ErrConnectDB = errors.New("err connect do")
)

type Database interface {
	Connect(connStr string) error

	LinkWorker
}

type LinkWorker interface {
	SaveLink() error
	GetLinkByDescription(desc string) (models.Link, error)
	GetLinksByUsername(username string) ([]models.Link, error)
	DeleteLink() error
}
