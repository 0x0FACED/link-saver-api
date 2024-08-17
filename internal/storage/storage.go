package storage

import (
	"errors"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
)

var (
	ErrConnectDB = errors.New("err connect db")
)

type Database interface {
	Connect() error

	LinkWorker
}

type LinkWorker interface {
	SaveLink() error
	GetLinksByUsernameDesc(username string, desc string) ([]models.Link, error)
	DeleteLink() error
}
