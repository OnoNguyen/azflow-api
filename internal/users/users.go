package users

import (
	"azflow-api/handlers"
	//database "azflow-api/internal/pkg/db/mysql"
	database "azflow-api/internal/pkg/db/postgresql"
	"database/sql"
	"log"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"name"`
	Password string `json:"password"`
}

func (user *User) Authenticate() bool {
	stmt, err := database.Db.Prepare("select Password from Users where Username = $1")
	if err != nil {
		log.Fatal(err)
	}
	row := stmt.QueryRow(user.Username)

	var hashedPassword string
	err = row.Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Fatal(err)
	}

	return handlers.CheckPasswordHash(user.Password, hashedPassword)
}

func (user *User) Create() {
	stmt, err := database.Db.Prepare("INSERT INTO Users(Username, Password) VALUES ($1, $2)")
	print(stmt)
	if err != nil {
		log.Fatal(err)
	}
	hashedPassword, err := handlers.HashPassord(user.Password)
	_, err = stmt.Exec(user.Username, hashedPassword)
	if err != nil {
		log.Fatal(err)
	}
}

func GetUserIdByUsername(username string) (int, error) {
	stmt, err := database.Db.Prepare("SELECT ID FROM Users WHERE Username = $1")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	row := stmt.QueryRow(username)
	var id int
	err = row.Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return 0, err
	}
	return id, nil
}

func GetAll() []User {
	stmt, err := database.Db.Prepare("select ID, Username from Users")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return users
}
