package functions

import (
	"database/sql"
	"fmt"
	"net/http"
)

const redirect = 301

// function that handle the sign up
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

	//hash the password
	password, err = HashPassword(password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//we creat a new user in the db
	_, err = db.Exec("INSERT INTO users (username, password, email, isAdmin, isBanned) VALUES (?, ?, ?, FALSE, FALSE)", username, password, email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", redirect)
}

// function that handle the login
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
	row := db.QueryRow("SELECT * FROM users WHERE email=?", email)
	var id int
	var storedUsername, storedPassword, storedEmail string
	var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&id, &storedUsername, &storedPassword, &storedEmail, &isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintln(w, "User not found")
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		}
		return
	}

	//check the password
	if !CheckPasswordHash(password, storedPassword) {
		fmt.Fprintln(w, "Wrong password")
		return
	}

	http.Redirect(w, r, "/home", redirect)
}