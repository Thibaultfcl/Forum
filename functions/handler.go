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

	fmt.Println("token: ", token)

	//get the user data from the database
    row := db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
    var isAdmin, isBanned bool
	var pp []byte

	//scan and get the data
	err := row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveHomePage(w, false, "")
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	profilePicture := ""
	if pp != nil {
		profilePicture = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pp)
	}

	serveHomePage(w, true, profilePicture)
}

func serveHomePage(w http.ResponseWriter, isLoggedIn bool, pp string) {
    userData := UserData{IsLoggedIn: isLoggedIn, ProfilePicture: pp}
    tmpl, err := template.ParseFiles("tmpl/home.html")
    if err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
        return
    }
    if err := tmpl.Execute(w, userData); err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
    }
}