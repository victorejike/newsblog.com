package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./blog.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE users SET is_admin = 1 WHERE id = 1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("User 1 is now an admin")
}
