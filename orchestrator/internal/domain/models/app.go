package models

type App struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Secret string `db:"secret"`
}
