package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/proto-files/link_service/gen"
)

var (
	ErrConnectDB      = errors.New("err connect db")
	ErrUserNotFound   = errors.New("user not found")
	ErrLinksNotFound  = errors.New("links not found")
	ErrNoRowsAffected = errors.New("no rows affected")
	ErrBeginTx        = "Cant begin tx"
)

type Database interface {
	Connect() error

	LinkWorker
	UserWorker
}

type UserWorker interface {
	SaveUser(ctx context.Context, tx *sql.Tx, u *models.User) (int, error)
	GetUserByTelegramID(ctx context.Context, tx *sql.Tx, userID int64) (*models.User, error)
	GetUserIDByTelegramID(ctx context.Context, tx *sql.Tx, userID int64) (int, error)
	GetTelegramIDByID(ctx context.Context, tx *sql.Tx, id int) (int64, error)
}

type LinkWorker interface {
	SaveLink(ctx context.Context, l *models.Link) error
	GetUserLinks(ctx context.Context, userID int64) ([]*gen.Link, error)
	GetContentByTelegramIDOriginalURL(ctx context.Context, userID int64, originalURL string) ([]byte, error)
	GetLinksByTelegramIDDesc(ctx context.Context, userID int64, desc string) ([]*gen.Link, error)
	GetLinkByID(ctx context.Context, id int) (*models.Link, error)
	DeleteLink(ctx context.Context, l *models.Link) error
}
