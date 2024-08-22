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
	q := `INSERT INTO users (telegram_user_id) VALUES ($1) RETURNING id`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, u.UserID).Scan(&id)
	} else {
		err = p.db.QueryRowContext(ctx, q, u.UserID).Scan(&id)
	}
	if err != nil {
		log.Println("[DB] error SaveUser():", err)
		return -1, err
	}

	return id, nil
}

func (p *Postgres) GetTelegramIDByID(ctx context.Context, tx *sql.Tx, id int) (int64, error) {
	var userID int64
	var err error
	q := `SELECT telegram_user_id FROM users WHERE id = $1`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, id).Scan(&userID)
	} else {
		err = p.db.QueryRowContext(ctx, q, id).Scan(&userID)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return -1, storage.ErrUserNotFound
		}
		log.Println("[DB] error GetUserByUsername():", err)
		return -1, err
	}
	return userID, nil
}

func (p *Postgres) GetUserByTelegramID(ctx context.Context, tx *sql.Tx, userID int64) (*models.User, error) {
	var u models.User
	var err error
	q := `SELECT id, telegram_user_id FROM users WHERE telegram_user_id = $1`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, userID).Scan(&u.ID, &u.UserID)
	} else {
		err = p.db.QueryRowContext(ctx, q, userID).Scan(&u.ID, &u.UserID)
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

func (p *Postgres) GetUserIDByTelegramID(ctx context.Context, tx *sql.Tx, userID int64) (int, error) {
	var id int
	var err error
	q := `SELECT id FROM users WHERE telegram_user_id = $1`
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, userID).Scan(&id)
	} else {
		err = p.db.QueryRowContext(ctx, q, userID).Scan(&id)
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
