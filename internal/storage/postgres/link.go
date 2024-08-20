package postgres

import (
	"context"
	"log"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	"github.com/0x0FACED/link-saver-api/internal/wrap"
	"github.com/0x0FACED/proto-files/link_service/gen"
)

func (p *Postgres) SaveLink(ctx context.Context, l *models.Link) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return wrap.E("Cant begin tx with err", err)
	}
	defer tx.Rollback()

	u, err := p.GetUserByUsername(ctx, tx, l.UserName)
	var id int
	if err == storage.ErrUserNotFound {
		u = &models.User{
			UserName: l.UserName,
		}
		id, err = p.SaveUser(ctx, tx, u)
		if err != nil {
			return err
		}
	}

	q := `INSERT INTO links (original_url, user_id, description, content) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, q, l.OriginalURL, id, l.Description, l.Content)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetUserLinks(ctx context.Context, username string) ([]*gen.Link, error) {
	userID, err := p.GetUserIDByUsername(ctx, nil, username)
	if err != nil {
		return nil, storage.ErrUserNotFound
	}
	q := `SELECT id, original_url, description FROM links WHERE user_id = $1`
	rows, err := p.db.QueryContext(ctx, q, userID)
	if err != nil {
		log.Println("[DB] error GetUserLinks():", err)
		return nil, err
	}
	defer rows.Close()

	var links []*gen.Link
	for rows.Next() {
		var l gen.Link
		if err := rows.Scan(&l.LinkId, &l.OriginalUrl, &l.Description); err != nil {
			log.Println("[DB] error scanning row in GetUserLinks():", err)
			return nil, err
		}
		links = append(links, &l)
	}

	if err := rows.Err(); err != nil {
		log.Println("[DB] error in rows.Err() GetUserLinks():", err)
		return nil, err
	}

	return links, nil
}

func (p *Postgres) GetLinksByUsernameDesc(ctx context.Context, username string, desc string) ([]*gen.Link, error) {
	userID, err := p.GetUserIDByUsername(ctx, nil, username)
	if err != nil {
		return nil, storage.ErrUserNotFound
	}

	q := `SELECT id, original_url, description FROM links WHERE user_id = $1 AND description LIKE $2`
	rows, err := p.db.QueryContext(ctx, q, userID, "%"+desc+"%")
	if err != nil {
		log.Println("[DB] error GetLinksByUsernameDesc():", err)
		return nil, err
	}
	defer rows.Close()

	var links []*gen.Link
	for rows.Next() {
		var l gen.Link
		if err := rows.Scan(&l.LinkId, &l.OriginalUrl, &l.Description); err != nil {
			log.Println("[DB] error scanning row in GetLinksByUsernameDesc():", err)
			return nil, err
		}
		links = append(links, &l)
	}

	if err := rows.Err(); err != nil {
		log.Println("[DB] error in rows.Err() GetLinksByUsernameDesc():", err)
		return nil, err
	}

	return links, nil
}

func (p *Postgres) GetLinkByID(ctx context.Context, id int) (*models.Link, error) {
	q := `SELECT original_url, username, description, content FROM links WHERE id = $1`
	l := models.Link{}
	err := p.db.QueryRowContext(ctx, q, id).Scan(&l.OriginalURL, &l.UserName, &l.Description, &l.Content)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (p *Postgres) DeleteLink(ctx context.Context, l *models.Link) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("[DB] error starting tx in DeleteLink():", err)
		return err
	}
	defer tx.Rollback()

	userID, err := p.GetUserIDByUsername(ctx, tx, l.UserName)
	if err != nil {
		return storage.ErrUserNotFound
	}

	q := `DELETE FROM links WHERE id = $1 AND user_id = $2 AND original_url = $3`
	result, err := tx.ExecContext(ctx, q, l.ID, userID, l.OriginalURL)
	if err != nil {
		log.Println("[DB] error executing delete in DeleteLink():", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("[DB] error checking rows affected in DeleteLink():", err)
		return err
	}

	if rowsAffected == 0 {
		return storage.ErrNoRowsAffected
	}

	if err = tx.Commit(); err != nil {
		log.Println("[DB] error committing tx in DeleteLink():", err)
		return err
	}

	return nil
}
