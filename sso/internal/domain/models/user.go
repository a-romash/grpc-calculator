package models

type User struct {
	ID       int
	Email    string
	Password []byte
}
