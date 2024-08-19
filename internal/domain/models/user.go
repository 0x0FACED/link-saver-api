package models

type User struct {
	ID       int    `db:"id"`
	UserName string `db:"username"`
}
