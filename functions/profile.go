package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type ProfileData struct {
	Username           string
	Email              string
	IsAdmin            bool
	IsBanned           bool
	ProfilePicture     string
	Posts              []PostData
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
}

func Profile(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	//get the user data from the database
	row := db.QueryRow("SELECT id, username, email, isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var id int
	var username, email string
	var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&id, &username, &email, &isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/home", http.StatusMovedPermanently)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	posts := getPostsFromUser(w, db, id)
	categories := getCategoriesByNumberOfPost(w, db)
	categoriesFollowed := getCategoriesFollowed(w, db, token)

	var profilePicture string
	if pp != nil {
		profilePicture = base64.StdEncoding.EncodeToString(pp)
	}

	serveProfilePage(w, username, email, isAdmin, isBanned, profilePicture, posts, categories, categoriesFollowed)
}

func serveProfilePage(w http.ResponseWriter, username string, email string, isAdmin bool, isBanned bool, pp string, posts []PostData, categories []CategoryData, categoriesFollowed []CategoryData) {
	profileData := ProfileData{Username: username, Email: email, IsAdmin: isAdmin, IsBanned: isBanned, ProfilePicture: pp, Posts: posts, Categories: categories, CategoriesFollowed: categoriesFollowed}
	tmpl, err := template.ParseFiles("tmpl/profile.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, profileData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
