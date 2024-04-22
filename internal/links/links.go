package links

import (
	//database "azflow-api/internal/pkg/db/mysql"
	database "azflow-api/internal/pkg/db/postgresql"
	"azflow-api/internal/users"
	"log"
)

type Link struct {
	ID      string
	Title   string
	Address string
	User    *users.User
}

func GetAll() []Link {
	stmt, err := database.Db.Prepare("select L.id, L.title, L.address, L.UserID, U.Username from Links L inner join Users U on L.UserID = U.ID")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var links []Link
	var username string
	var id string
	for rows.Next() {
		var link Link
		err := rows.Scan(&link.ID, &link.Title, &link.Address, &id, &username)
		if err != nil {
			log.Fatal(err)
		}
		link.User = &users.User{ID: id, Username: username}
		links = append(links, link)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return links
}

func (link Link) Save() int64 {
	stmt, err := database.Db.Prepare("INSERT INTO Links(Title, Address, UserID) VALUES ($1, $2, $3) RETURNING ID")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	var linkID int64

	err = stmt.QueryRow(link.Title, link.Address, link.User.ID).Scan(&linkID)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
	log.Print("Row inserted")
	return linkID
}
