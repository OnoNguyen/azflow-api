package users

import (
	"database/sql"
	"log"
)

func mapToUsers(row *sql.Rows, users *[]User) {
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		log.Fatal(err)
	}
	*users = append(*users, user)
}
