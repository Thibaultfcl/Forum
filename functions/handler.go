package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type UserData struct {
	IsLoggedIn     bool
	ProfilePicture string
	Categories     []string
	Posts          []PostData
}

type PostData struct {
	Title         string
	Content       string
	Category      string
	Author        string
	AuthorPicture string
	TimePosted    string
}

// home page
func Home(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	posts := getPosts(w, db)

	//get the user data from the database
	row := db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveHomePage(w, false, "", nil, posts)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	var profilePicture string
	if pp != nil {
		profilePicture = base64.StdEncoding.EncodeToString(pp)
	}

	var categories []string
	rows, _ := db.Query("SELECT name FROM categories")
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		categories = append(categories, category)
	}

	serveHomePage(w, true, profilePicture, categories, posts)
}

func serveHomePage(w http.ResponseWriter, isLoggedIn bool, pp string, categories []string, posts []PostData) {
	userData := UserData{IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories, Posts: posts}
	tmpl, err := template.ParseFiles("tmpl/home.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
