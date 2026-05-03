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

	// Check users table structure
	fmt.Println("Users table structure:")
	rows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var type_ string
		var notnull int
		var dfltValue sql.NullString
		var pk int

		err := rows.Scan(&cid, &name, &type_, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("  %s: %s\n", name, type_)
	}

	// Check posts table structure
	fmt.Println("Posts table structure:")
	rows, err = db.Query("PRAGMA table_info(posts)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var type_ string
		var notnull int
		var dfltValue sql.NullString
		var pk int

		err := rows.Scan(&cid, &name, &type_, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("  %s: %s\n", name, type_)
	}

	// Check if there are any posts
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nTotal posts: %d\n", count)

	// Check categories table
	fmt.Println("\nCategories table structure:")
	rows, err = db.Query("PRAGMA table_info(categories)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var type_ string
		var notnull int
		var dfltValue sql.NullString
		var pk int

		err := rows.Scan(&cid, &name, &type_, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("  %s: %s\n", name, type_)
	}

	// Check if there are any categories
	var catCount int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nTotal categories: %d\n", catCount)

	// List categories
	fmt.Println("\nCategories:")
	rows, err = db.Query("SELECT id, name, slug FROM categories")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, slug string
		err := rows.Scan(&id, &name, &slug)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %d: %s (%s)\n", id, name, slug)
	}

	// Check if there are any users
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nTotal users: %d\n", userCount)

	// List users
	fmt.Println("\nUsers:")
	rows, err = db.Query("SELECT id, username, email, is_admin FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, email string
		var isAdmin bool
		err := rows.Scan(&id, &username, &email, &isAdmin)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %d: %s (%s) - Admin: %t\n", id, username, email, isAdmin)
	}
}
