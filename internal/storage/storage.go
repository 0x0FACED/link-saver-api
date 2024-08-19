package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
)

var (
	ErrConnectDB      = errors.New("err connect db")
	ErrUserNotFound   = errors.New("user not found")
	ErrLinksNotFound  = errors.New("links not found")
	ErrNoRowsAffected = errors.New("no rows affected")
)

type Database interface {
	Connect() error

	LinkWorker
	UserWorker
}

type UserWorker interface {
	SaveUser(ctx context.Context, tx *sql.Tx, u *models.User) (int, error)
	GetUserByUsername(ctx context.Context, tx *sql.Tx, username string) (*models.User, error)
	GetUserIDByUsername(ctx context.Context, tx *sql.Tx, username string) (int, error)
}

type LinkWorker interface {
	SaveLink(ctx context.Context, l *models.Link) error
	GetUserLinks(ctx context.Context, username string) ([]models.Link, error)
	GetLinksByUsernameDesc(ctx context.Context, username string, desc string) ([]models.Link, error)
	DeleteLink(ctx context.Context, l *models.Link) error
}
