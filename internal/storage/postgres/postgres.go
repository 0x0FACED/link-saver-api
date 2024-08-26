package postgres

import (
	"database/sql"
	"fmt"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/wrap"
	"github.com/0x0FACED/link-saver-api/migrations"
	_ "github.com/lib/pq"
)

var pkg = "storage/postgres"

type Postgres struct {
	db     *sql.DB
	config config.DatabaseConfig
}

func New(cfg config.DatabaseConfig) *Postgres {
	return &Postgres{
		config: cfg,
	}
}

func (p *Postgres) Connect() error {
	connStr := p.getConnStr()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return wrap.E(pkg, "failed to Open()", err)
	}

	if db.Ping() != nil {
		return wrap.E(pkg, "failed to Ping()", err)
	}

	p.db = db

	err = migrations.Up(connStr)
	if err != nil {
		return wrap.E(pkg, "failed to Up()", err)
	}

	return nil
}

func (p Postgres) getConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.config.Username, p.config.Password, p.config.Host, p.config.Port, p.config.Name)
}
