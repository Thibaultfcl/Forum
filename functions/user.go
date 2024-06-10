package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type UserPageData struct {
	UserPage           UserPage
	IsLoggedIn         bool
	ProfilePicture     string
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
	AllCategories      []CategoryData
	Posts              []PostData
}

type UserPage struct {
	Username       string
	IsBanned       bool
	pp             []byte
	ProfilePicture string
}

// user page
func User(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	userIDStr := r.URL.Path[6:]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT username, isBanned, pp FROM users WHERE id=?", userID)
	var user UserPage
	err = row.Scan(&user.Username, &user.IsBanned, &user.pp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error user not found: %v", err), http.StatusInternalServerError)
		return
	}

	if user.pp != nil {
		user.ProfilePicture = base64.StdEncoding.EncodeToString(user.pp)
	}

	posts := getPostsFromUser(w, db, userID, token)
	categories := getCategoriesByNumberOfPost(w, db)
	allCategories := getAllCategories(w, db)
	categoriesFollowed := getCategoriesFollowed(w, db, token)

	row = db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var isAdmin, isBanned bool
	var pp []byte
	err = row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveUserPage(w, user, false, "", categories, nil, nil, posts)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	for i := range posts {
		posts[i].IsLoggedIn = true
	}

	var profilePicture string
	if pp != nil {
		profilePicture = base64.StdEncoding.EncodeToString(pp)
	}

	serveUserPage(w, user, true, profilePicture, categories, categoriesFollowed, allCategories, posts)
}

func serveUserPage(w http.ResponseWriter, userPage UserPage, isLoggedIn bool, pp string, categories []CategoryData, categoriesFollowed []CategoryData, allCategories []CategoryData, posts []PostData) {
	userData := UserPageData{UserPage: userPage, IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories, CategoriesFollowed: categoriesFollowed, AllCategories: allCategories, Posts: posts}
	tmpl, err := template.ParseFiles("tmpl/user.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
