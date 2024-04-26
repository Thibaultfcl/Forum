package main

import (
	"database/sql"
	"fmt"
	"forum/functions"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// port of the server
const port = ":8080"

// home page
func Home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/home.html")
}

// function that creat a table User
func createTableUser(db *sql.DB) {
	//creating the user table if not already created
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username VARCHAR(12) NOT NULL,
            password VARCHAR(12) NOT NULL,
			email TEXT NOT NULL,
			isAdmin BOOL NOT NULL DEFAULT FALSE,
			isBanned BOOL NOT NULL DEFAULT FALSE,
			pp BLOB
        )
    `)
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	//open the database with sqlite3
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		panic(err.Error())
	}
	//creat the 2 tables
	createTableUser(db)
	defer db.Close()

	//handle the different pages
	http.HandleFunc("/home", Home)
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) { functions.Signin(w, r, db) })
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) { functions.Signup(w, r, db) })

	//load the CSS and the images
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	//start the local host
	fmt.Println("\n(http://localhost:8080/home) - Server started on port", port)
	http.ListenAndServe(port, nil)
}
