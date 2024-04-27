package functions

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

	//open the default profile picture
	file, err := os.Open("img/profileDefault.jpg")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening image: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	//read the image
	ppDefault, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading image: %v", err), http.StatusInternalServerError)
		return
	}

	//generate a session token
	token, err := GenerateSessionToken()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//we creat a new user in the db
	_, err = db.Exec("INSERT INTO users (username, password, email, isAdmin, isBanned, pp, UUID) VALUES (?, ?, ?, FALSE, FALSE, ?, ?)", username, password, email, ppDefault, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//set the session token
	SetSessionToken(token, w)

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
	var token string

	//scan and get the data
	err := row.Scan(&id, &storedUsername, &storedPassword, &storedEmail, &isAdmin, &isBanned, &pp, &token)
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

	//generate a session token
	token, err = GenerateSessionToken()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//we update the user with the new token
	_, err = db.Exec("UPDATE users SET UUID = ? WHERE email = ?", token, email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//set the session token
	SetSessionToken(token, w)

	http.Redirect(w, r, "/home", redirect)
}

// function that handle the logout
func Logout(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)

	_, err := db.Exec("UPDATE users SET UUID = '' WHERE UUID = ?", token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	ClearSessionToken(w)

	http.Redirect(w, r, "/home", redirect)
}
