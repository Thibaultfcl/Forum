package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
)

type UserData struct {
	IsLoggedIn bool
	ProfilePicture string
	Categories []string
}

// home page
func Home(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)

	fmt.Println("token: ", token)

	//get the user data from the database
    row := db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
    var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveHomePage(w, false, "", nil)
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

	serveHomePage(w, true, profilePicture, categories)
}

func serveHomePage(w http.ResponseWriter, isLoggedIn bool, pp string, categories []string) {
    userData := UserData{IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories}
    tmpl, err := template.ParseFiles("tmpl/home.html")
    if err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
        return
    }
    if err := tmpl.Execute(w, userData); err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
    }
}