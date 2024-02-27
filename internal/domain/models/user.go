package models

type User struct {
	ID       int
	Name     string
	PassHash []byte
}
