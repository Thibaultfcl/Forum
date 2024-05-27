package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
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

func getPosts(w http.ResponseWriter, db *sql.DB) []PostData {
	var posts []PostData
	rows, err := db.Query(`SELECT title, content, date, category, author FROM posts ORDER BY date DESC`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
	}
	for rows.Next() {
		var category, author int
		var title, content, date string
		err := rows.Scan(&title, &content, &date, &category, &author)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var categoryStr, authorStr string
		row := db.QueryRow("SELECT name FROM categories WHERE id=?", category)
		err = row.Scan(&categoryStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row = db.QueryRow("SELECT username FROM users WHERE id=?", author)
		err = row.Scan(&authorStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var authorPP []byte
		row = db.QueryRow("SELECT pp FROM users WHERE id=?", author)
		err = row.Scan(&authorPP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		const layout = "2006-01-02T15:04:05Z07:00"
		t, err := time.Parse(layout, date)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		t = t.Local()
		elapsed := time.Since(t)
		var elapsedStr string
		if elapsed < time.Minute {
			elapsedStr = fmt.Sprintf("%d seconds ago", int(elapsed.Seconds()))
		} else if elapsed < time.Hour {
			elapsedStr = fmt.Sprintf("%d minutes ago", int(elapsed.Minutes()))
		} else if elapsed < time.Hour*24 {
			elapsedStr = fmt.Sprintf("%d hours ago", int(elapsed.Hours()))
		} else {
			elapsedStr = fmt.Sprintf("%d days ago", int(elapsed.Hours()/24))
		}
		posts = append(posts, PostData{Title: title, Content: content, Category: categoryStr, Author: authorStr, AuthorPicture: base64.StdEncoding.EncodeToString(authorPP), TimePosted: elapsedStr})
	}
	return posts
}

func getPostsFromUser(w http.ResponseWriter, db *sql.DB, authorID int) []PostData {
	var posts []PostData
	rows, err := db.Query("SELECT title, content, date, category, author FROM posts WHERE author=? ORDER BY date DESC", authorID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
	}
	for rows.Next() {
		var category, author int
		var title, content, date string
		err := rows.Scan(&title, &content, &date, &category, &author)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var categoryStr, authorStr string
		row := db.QueryRow("SELECT name FROM categories WHERE id=?", category)
		err = row.Scan(&categoryStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row = db.QueryRow("SELECT username FROM users WHERE id=?", author)
		err = row.Scan(&authorStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var authorPP []byte
		row = db.QueryRow("SELECT pp FROM users WHERE id=?", author)
		err = row.Scan(&authorPP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		const layout = "2006-01-02T15:04:05Z07:00"
		t, err := time.Parse(layout, date)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		t = t.Local()
		elapsed := time.Since(t)
		var elapsedStr string
		if elapsed < time.Minute {
			elapsedStr = fmt.Sprintf("%d seconds ago", int(elapsed.Seconds()))
		} else if elapsed < time.Hour {
			elapsedStr = fmt.Sprintf("%d minutes ago", int(elapsed.Minutes()))
		} else if elapsed < time.Hour*24 {
			elapsedStr = fmt.Sprintf("%d hours ago", int(elapsed.Hours()))
		} else {
			elapsedStr = fmt.Sprintf("%d days ago", int(elapsed.Hours()/24))
		}
		posts = append(posts, PostData{Title: title, Content: content, Category: categoryStr, Author: authorStr, AuthorPicture: base64.StdEncoding.EncodeToString(authorPP), TimePosted: elapsedStr})
	}
	return posts
}