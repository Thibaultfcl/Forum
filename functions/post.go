package functions

import (
	"database/sql"
	"net/http"
)

func CreatePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//we check if the method is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//we get the title and the content
	title := r.FormValue("title")
	content := r.FormValue("content")

	//we get the token
	token := GetSessionToken(r)

	//we get the user id
	row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
	var id int

	//we scan to get the data
	err := row.Scan(&id)
	if err != nil {
		http.Error(w, "You need to be logged in to create a post", http.StatusUnauthorized)
		return
	}

	//we insert the post in the database
	_, err = db.Exec("INSERT INTO posts (title, content, user_id) VALUES (?, ?, ?)", title, content, id)
	if err != nil {
		http.Error(w, "Error while creating the post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", redirect)
}
