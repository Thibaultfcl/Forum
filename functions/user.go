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
	UserID             int
	ProfilePicture     string
	IsAdmin            bool
	IsBanned           bool
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
	AllCategories      []CategoryData
	Posts              []PostData
	HisAccount         bool
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

	row = db.QueryRow("SELECT id, isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var id int
	var isAdmin, isBanned bool
	var pp []byte
	err = row.Scan(&id, &isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveUserPage(w, user, false, 0, "", false, false, categories, nil, nil, posts, false)
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

	if id == userID {
		serveUserPage(w, user, true, id, profilePicture, isAdmin, isBanned, categories, categoriesFollowed, allCategories, posts, true)
		return
	} else {
		serveUserPage(w, user, true, id, profilePicture, isAdmin, isBanned, categories, categoriesFollowed, allCategories, posts, false)
		return
	}
}

func serveUserPage(w http.ResponseWriter, userPage UserPage, isLoggedIn bool, userID int, pp string, isAdmin bool, isBanned bool, categories []CategoryData, categoriesFollowed []CategoryData, allCategories []CategoryData, posts []PostData, hisAccount bool) {
	userData := UserPageData{UserPage: userPage, IsLoggedIn: isLoggedIn, UserID: userID, ProfilePicture: pp, IsAdmin: isAdmin, IsBanned: isBanned, Categories: categories, CategoriesFollowed: categoriesFollowed, AllCategories: allCategories, Posts: posts, HisAccount: hisAccount}
	tmpl, err := template.ParseFiles("tmpl/user.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
