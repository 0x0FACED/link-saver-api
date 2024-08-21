package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/storage"
)

func (p *Postgres) SaveUser(ctx context.Context, tx *sql.Tx, u *models.User) (int, error) {
	var err error
	var id int
	q := `INSERT INTO users (username) VALUES ($1) RETURNING id`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, u.UserName).Scan(&id)
	} else {
		err = p.db.QueryRowContext(ctx, q, u.UserName).Scan(&id)
	}
	if err != nil {
		log.Println("[DB] error SaveUser():", err)
		return -1, err
	}

	return id, nil
}

func (p *Postgres) GetUsernameByID(ctx context.Context, tx *sql.Tx, id int) (string, error) {
	var username string
	var err error
	q := `SELECT username FROM users WHERE id = $1`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, id).Scan(&username)
	} else {
		err = p.db.QueryRowContext(ctx, q, id).Scan(&username)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return "", storage.ErrUserNotFound
		}
		log.Println("[DB] error GetUserByUsername():", err)
		return "", err
	}
	return username, nil
}

func (p *Postgres) GetUserByUsername(ctx context.Context, tx *sql.Tx, username string) (*models.User, error) {
	var u models.User
	var err error
	q := `SELECT id, username FROM users WHERE username = $1`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, username).Scan(&u.ID, &u.UserName)
	} else {
		err = p.db.QueryRowContext(ctx, q, username).Scan(&u.ID, &u.UserName)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrUserNotFound
		}
		log.Println("[DB] error GetUserByUsername():", err)
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) GetUserIDByUsername(ctx context.Context, tx *sql.Tx, username string) (int, error) {
	var id int
	var err error
	q := `SELECT id FROM users WHERE username = $1`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, username).Scan(&id)
	} else {
		err = p.db.QueryRowContext(ctx, q, username).Scan(&id)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return -1, storage.ErrUserNotFound
		}
		log.Println("[DB] error GetUserByUsername():", err)
		return -1, err
	}
	return id, nil
}
