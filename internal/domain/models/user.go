package models

type User struct {
	ID     int   `db:"id"`
	UserID int64 `db:"telegram_user_id"`
}
