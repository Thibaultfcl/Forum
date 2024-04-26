package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// constant for the redirect
const port = ":8080"
const redirect = 301

// variable that stock the conected user
var ConnectedUser string

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

//function that handle the sign up
func Signup(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//we check if the method is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//we get the username, the email and the password
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	//request SQL to check if the user already exist
	row := db.QueryRow("SELECT * FROM users WHERE username=?", username)
	var id int
	var storedUsername, storedPassword, storedEmail string
	var isAdmin, isBanned bool

	//we scan to get the data
	err := row.Scan(&id, &storedUsername, &storedPassword, &storedEmail, &isAdmin, &isBanned)

	//we check if the username is already used
	if err == nil {
		fmt.Fprintln(w, "This username already exist, please select another one")
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//we creat a new user in the db
	_, err = db.Exec("INSERT INTO users (username, password, email, isAdmin, isBanned) VALUES (?, ?, ?, FALSE, FALSE)", username, password, email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	//change the user connected
	ConnectedUser = username

	http.Redirect(w, r, "/home", redirect)
}

//function that handle the login
func Signin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//we check if the method is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//we get the email and the password
	email := r.FormValue("email")
	password := r.FormValue("password")

	//request SQL to get the user
	row := db.QueryRow("SELECT * FROM users WHERE email=? AND password=?", email, password)
	var id int
	var storedUsername, storedPassword, storedEmail string
	var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&id, &storedUsername, &storedPassword, &storedEmail, &isAdmin, &isBanned, &pp)

	//we compare the data
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintln(w, "User not found")
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		}
		return
	}
	//change the user connected
	ConnectedUser = storedUsername

	http.Redirect(w, r, "/home", redirect)
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

	http.HandleFunc("/home", Home)
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) { Signin(w, r, db) })
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) { Signup(w, r, db) })

	//load the CSS and the images
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	//start the local host
	fmt.Println("\n(http://localhost:8080/home) - Server started on port", port)
	http.ListenAndServe(port, nil)
}
