package functions

import (
	"database/sql"
	"fmt"
	"net/http"
)

func CreatePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//we check if the method is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//we get the token
	token := GetSessionToken(r)

	//we get the user id
	row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
	var id int

	//we scan to get the data
	err := row.Scan(&id)
	if err != nil {
		http.Error(w, fmt.Sprintf("You need to be logged in to create a post: %v", err), http.StatusUnauthorized)
		return
	}

	//we get the title and the content
	category := r.FormValue("Category")
	title := r.FormValue("PostName")
	content := r.FormValue("PostContent")

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while starting the transaction: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert the category if it doesn't exist
	_, err = tx.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while creating the category: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the category id
	row = tx.QueryRow("SELECT id FROM categories WHERE name=?", category)
	var idCategory int
	err = row.Scan(&idCategory)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while getting the category id: %v", err), http.StatusInternalServerError)
		return
	}

	//we insert the post in the database
	_, err = tx.Exec("INSERT INTO posts (title, content, category ,author) VALUES (?, ?, ?, ?)", title, content, idCategory, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while creating the post: %v", err), http.StatusInternalServerError)
		return
	}

	// Increment the number of posts
	_, err = tx.Exec("UPDATE categories SET number_of_posts = number_of_posts + 1 WHERE id=?", idCategory)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while updating the number of posts: %v", err), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while commiting the transaction: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", redirect)
}
