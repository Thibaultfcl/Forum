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
}

// home page
func Home(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)

	//check if the token is empty
	if token == "" {
		http.ServeFile(w, r, "tmpl/home.html")
		return
	}

	//check if the user is connected
	row := db.QueryRow("SELECT * FROM users WHERE UUID=?", token)
	var id int
	var username, password, email string
	var isAdmin, isBanned bool
	var pp []byte
	var storedToken string

	//scan and get the data
	err := row.Scan(&id, &username, &password, &email, &isAdmin, &isBanned, &pp, &storedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			http.ServeFile(w, r, "tmpl/home.html")
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	userData := UserData{IsLoggedIn: true}

	var profilePicture string
	if pp != nil {
		profilePicture = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pp)
	}

	userData.ProfilePicture = profilePicture

	//parse the template
	tmpl, err := template.ParseFiles("tmpl/home.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	//execute the template
	if err := tmpl.Execute(w, userData); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
}
