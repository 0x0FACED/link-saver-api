package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
)

func (p *Postgres) SaveResource(ctx context.Context, res *models.Resource) error {
	q := `
		INSERT INTO resources (name, type_id, content) 
		VALUES ($1, $2, $3)
		ON CONFLICT (name) 
		DO UPDATE SET content = EXCLUDED.content
	`

	_, err := p.db.ExecContext(ctx, q, res.Name, res.Type, res.Content)
	if err != nil {
		return fmt.Errorf("failed to save resource: %w", err)
	}

	return nil
}

func (p *Postgres) GetResourceContentByNameType(ctx context.Context, name string, resType models.ResourceType) ([]byte, error) {
	q := `
		SELECT content 
		FROM resources 
		WHERE name = $1 AND type_id = $2
	`

	var content []byte
	err := p.db.QueryRowContext(ctx, q, name, resType).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get resource content: %w", err)
		}
		return nil, fmt.Errorf("failed to get resource content: %w", err)
	}

	return content, nil
}
