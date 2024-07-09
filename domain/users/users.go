package users

import (
	"azflow-api/handlers"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	//database "azflow-api/domain/pkg/db/mysql"
	database "azflow-api/domain/pkg/db/postgresql"
)

func (user *User) Authenticate() (bool, error) {
	var hashedPassword string
	err := pgxscan.Get(context.Background(), database.Db, &hashedPassword, "SELECT Password FROM Users WHERE Username = $1", user.Username)
	return handlers.CheckPasswordHash(user.Password, hashedPassword), err
}

func (user *User) Create() (*int, error) {
	hashedPassword := handlers.HashPassord(user.Password)
	var id *int
	err := pgxscan.Get(context.Background(), database.Db, &id, "INSERT INTO Users(Username, Password) VALUES ($1, $2) returning ID", user.Username, hashedPassword)
	return id, err
}

func GetUserIdByUsername(username string) (int, error) {
	var id *int
	err := pgxscan.Get(context.Background(), database.Db, &id, "SELECT ID FROM Users WHERE Username = $1", username)
	return *id, err
}

func GetAll() ([]*User, error) {
	var users []*User
	err := pgxscan.Select(context.Background(), database.Db, &users, "SELECT ID, Username FROM Users")
	return users, err
}
