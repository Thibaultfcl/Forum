package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
)

type ProfileData struct {
	Username       string
	Email          string
	IsAdmin        bool
	IsBanned       bool
	ProfilePicture string
}

func Profile(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	//get the user data from the database
	row := db.QueryRow("SELECT username, email, isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var username, email string
	var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&username, &email, &isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			posts := getPosts(w, db)
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

	serveProfilePage(w, username, email, isAdmin, isBanned, profilePicture)
}

func serveProfilePage(w http.ResponseWriter, username string, email string, isAdmin bool, isBanned bool, pp string) {
	profileData := ProfileData{Username: username, Email: email, IsAdmin: isAdmin, IsBanned: isBanned, ProfilePicture: pp}
	tmpl, err := template.ParseFiles("tmpl/profile.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, profileData); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}
