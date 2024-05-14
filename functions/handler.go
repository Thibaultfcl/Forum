package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
)

type UserData struct {
	IsLoggedIn     bool
	ProfilePicture string
	Categories     []string
	Posts		  []PostData
}

type PostData struct {
	Id       int
	Title    string
	Content  string
	Category string
	Author   string
}

// home page
func Home(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	posts := getPosts(w, db)
	fmt.Println("posts: ", posts)

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

func getPosts(w http.ResponseWriter, db *sql.DB) []PostData {
	var posts []PostData
	rows, err := db.Query(`SELECT id, title, content, date, category, author FROM posts ORDER BY date DESC`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return nil
	}
	for rows.Next() {
		var id, category, author int
		var title, content, date string
		err := rows.Scan(&id, &title, &content, &date, &category, &author)
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
		posts = append(posts, PostData{Id: id, Title: title, Content: content, Category: categoryStr, Author: authorStr})
	}
	return posts
}

func serveHomePage(w http.ResponseWriter, isLoggedIn bool, pp string, categories []string, posts []PostData) {
	userData := UserData{IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories, Posts: posts}
	tmpl, err := template.ParseFiles("tmpl/home.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}
