package postgres

import (
	"context"
	"errors"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/link-saver-api/internal/storage"
	"github.com/0x0FACED/link-saver-api/internal/wrap"
	"github.com/0x0FACED/proto-files/link_service/gen"
)

func (p *Postgres) SaveLink(ctx context.Context, l *models.Link) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return wrap.E(pkg, "failed to BeginTx()", err)
	}
	defer tx.Rollback()

	id, err := p.GetUserIDByTelegramID(ctx, tx, l.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			u := &models.User{
				ID:     id,
				UserID: l.UserID,
			}
			id, err = p.SaveUser(ctx, tx, u)
			if err != nil {
				return wrap.E(pkg, "failed to SaveLink(), SaveUser()", err)
			}
		} else {
			return wrap.E(pkg, "failed to GetUserIDByTelegramID()", err)
		}
	}

	q := `INSERT INTO links (original_url, user_id, description, content) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, q, l.OriginalURL, id, l.Description, l.Content)
	if err != nil {
		return wrap.E(pkg, "failed to SaveLink(), q="+q, err)
	}

	if err = tx.Commit(); err != nil {
		return wrap.E(pkg, "failed to Commit()", err)
	}

	return nil
}

func (p *Postgres) GetUserLinks(ctx context.Context, userID int64) ([]*gen.Link, error) {
	user_ID, err := p.GetUserIDByTelegramID(ctx, nil, userID)
	if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
		return nil, wrap.E(pkg, "failed to GetUserIDByTelegramID()", err)
	}

	if errors.Is(err, storage.ErrUserNotFound) {
		u := &models.User{
			UserID: userID,
		}
		user_ID, err = p.SaveUser(ctx, nil, u)
		if err != nil {
			return nil, wrap.E(pkg, "failed to SaveUser()", err)
		}
	}

	q := `SELECT id, original_url, description FROM links WHERE user_id = $1`
	rows, err := p.db.QueryContext(ctx, q, user_ID)
	if err != nil {
		return nil, wrap.E(pkg, "failed to GetUserLinks(), q="+q, err)
	}
	defer rows.Close()

	var links []*gen.Link
	for rows.Next() {
		var l gen.Link
		if err := rows.Scan(&l.LinkId, &l.OriginalUrl, &l.Description); err != nil {
			return nil, wrap.E(pkg, "failed to Scan()", err)
		}
		links = append(links, &l)
	}

	if err := rows.Err(); err != nil {
		return nil, wrap.E(pkg, "error in rows.Err()()", err)
	}

	return links, nil
}

func (p *Postgres) GetContentByTelegramIDOriginalURL(ctx context.Context, userID int64, originalURL string) ([]byte, error) {
	user_ID, err := p.GetUserIDByTelegramID(ctx, nil, userID)
	if err != nil && err != storage.ErrUserNotFound {
		return nil, wrap.E(pkg, "failed to GetUserIDByTelegramID()", err)
	}

	if err == storage.ErrUserNotFound {
		u := &models.User{
			UserID: userID,
		}
		user_ID, err = p.SaveUser(ctx, nil, u)
		if err != nil {
			return nil, wrap.E(pkg, "failed to SaveUser()", err)
		}
	}
	var content []byte

	q := `SELECT content FROM links WHERE user_id = $1 AND original_url = $2`
	err = p.db.QueryRowContext(ctx, q, user_ID, originalURL).Scan(&content)
	if err != nil {
		return nil, wrap.E(pkg, "failed to GetContent(), q="+q, err)
	}

	return content, nil
}

func (p *Postgres) GetLinksByTelegramIDDesc(ctx context.Context, userID int64, desc string) ([]*gen.Link, error) {
	user_ID, err := p.GetUserIDByTelegramID(ctx, nil, userID)
	if err != nil && err != storage.ErrUserNotFound {
		return nil, wrap.E(pkg, "failed to GetUserIDByTelegramID()", err)
	}

	if err == storage.ErrUserNotFound {
		u := &models.User{
			UserID: userID,
		}
		user_ID, err = p.SaveUser(ctx, nil, u)
		if err != nil {
			return nil, wrap.E(pkg, "failed to SaveUser()", err)
		}
	}

	q := `SELECT id, original_url, description FROM links WHERE user_id = $1 AND description LIKE $2`
	rows, err := p.db.QueryContext(ctx, q, user_ID, "%"+desc+"%")
	if err != nil {
		return nil, wrap.E(pkg, "failed to GetLinks(), q="+q, err)
	}
	defer rows.Close()

	var links []*gen.Link
	for rows.Next() {
		var l gen.Link
		if err := rows.Scan(&l.LinkId, &l.OriginalUrl, &l.Description); err != nil {
			return nil, wrap.E(pkg, "failed to Scan()", err)
		}
		links = append(links, &l)
	}

	if err := rows.Err(); err != nil {
		return nil, wrap.E(pkg, "error in rows.Err()", err)
	}

	return links, nil
}

func (p *Postgres) GetLinkByID(ctx context.Context, id int) (*models.Link, error) {
	q := `SELECT id, original_url, user_id, description FROM links WHERE id = $1`

	l := models.Link{}
	var user_ID int

	err := p.db.QueryRowContext(ctx, q, id).Scan(&l.ID, &l.OriginalURL, &user_ID, &l.Description)
	if err != nil {
		return nil, wrap.E(pkg, "failed to GetLinkByID()", err)
	}

	var userID int64
	userID, err = p.GetTelegramIDByID(ctx, nil, user_ID)
	l.UserID = userID

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, wrap.E(pkg, "unusual behavior, link exists, user not exists", err)
		}
		return nil, wrap.E(pkg, "failed to GetTelegramIDByID()", err)
	}

	return &l, nil
}

func (p *Postgres) DeleteLink(ctx context.Context, id int) (string, int64, error) {
	var originalURL string
	var telegramUserID int64
	var userID int

	q := `DELETE FROM links WHERE id = $1 RETURNING original_url, user_id`
	err := p.db.QueryRowContext(ctx, q, id).Scan(&originalURL, &userID)
	if err != nil {
		return "", -1, wrap.E(pkg, "failed to DeleteLink()", err)
	}

	telegramUserID, err = p.GetTelegramIDByID(ctx, nil, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", -1, wrap.E(pkg, "unusual behavior, link deleted, user not exists", err)
		}
		return "", -1, wrap.E(pkg, "failed to GetTelegramIDByID()", err)
	}

	return originalURL, telegramUserID, nil
}
