package postgres

import (
	"database/sql"
	"fmt"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db     *sql.DB
	Config config.DatabaseConfig
}

func (p *Postgres) Connect() error {
	db, err := sql.Open("postgres", p.getConnStr())
	if err != nil {
		return storage.ErrConnectDB
	}

	if db.Ping() != nil {
		return storage.ErrConnectDB
	}

	p.db = db

	return nil
}

func (p *Postgres) getConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.Config.Username, p.Config.Password, p.Config.Host, p.Config.Port, p.Config.Name)
}

func (p *Postgres) SaveLink() error {
	panic("not implemented") // TODO: Implement
}

func (p *Postgres) GetLinksByUsernameDesc(username string, desc string) ([]models.Link, error) {
	panic("not implemented") // TODO: Implement
}

func (p *Postgres) DeleteLink() error {
	panic("not implemented") // TODO: Implement
}
