package models

import "time"

type Link struct {
	ID          int       `json:"id" db:"id"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	UserName    string    `json:"username" db:"username" required:"true"`
	Description string    `json:"description" db:"description" required:"true"`
	Content     []byte    `db:"content"`
	DateAdded   time.Time `json:"date_added" db:"date_added"`
}
