package postgres

import (
	"database/sql"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func (p *Postgres) Connect(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return storage.ErrConnectDB
	}

	if db.Ping() != nil {
		return storage.ErrConnectDB
	}

	p.db = db

	return nil
}

func (p *Postgres) SaveLink() error {
	panic("not implemented") // TODO: Implement
}

func (p *Postgres) GetLinkByDescription(desc string) (models.Link, error) {
	panic("not implemented") // TODO: Implement
}

func (p *Postgres) GetLinksByUsername(username string) ([]models.Link, error) {
	panic("not implemented") // TODO: Implement
}

func (p *Postgres) DeleteLink() error {
	panic("not implemented") // TODO: Implement
}
