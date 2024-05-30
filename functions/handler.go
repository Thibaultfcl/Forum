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
	Categories     []CategoryData
	AllCategories  []CategoryData
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
	categories := getCategoriesByNumberOfPost(w, db)
	allCategories := getAllCategories(w, db)

	//get the user data from the database
	row := db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveHomePage(w, false, "", categories, nil, posts)
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

	serveHomePage(w, true, profilePicture, categories, allCategories, posts)
}

func serveHomePage(w http.ResponseWriter, isLoggedIn bool, pp string, categories []CategoryData, allCategories []CategoryData, posts []PostData) {
	userData := UserData{IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories, AllCategories: allCategories, Posts: posts}
	tmpl, err := template.ParseFiles("tmpl/home.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
