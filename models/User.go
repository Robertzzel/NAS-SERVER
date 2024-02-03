package models

type User struct {
	IsAuthenticated   bool
	Name              string
	UserRootDirectory string
}

func NewUser() User {
	return User{}
}
