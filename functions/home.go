package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type UserData struct {
	IsLoggedIn         bool
	UserID             int
	ProfilePicture     string
	IsAdmin            bool
	IsBanned           bool
	Categories         []CategoryData
	AllCategories      []CategoryData
	CategoriesFollowed []CategoryData
	Posts              []PostData
	MostLikedPosts     []PostData
}

// home page
func Home(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)

	posts := getPosts(w, db, token)
	categories := getCategoriesByNumberOfPost(w, db)
	allCategories := getAllCategories(w, db)
	categoriesFollowed := getCategoriesFollowed(w, db, token)
	mostLikedPost := getMostLikedPosts(w, db, token)

	//get the user data from the database
	row := db.QueryRow("SELECT id, isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var user UserData
	var pp []byte
	err := row.Scan(&user.UserID, &user.IsAdmin, &user.IsBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveHomePage(w, false, 0, "", false, false, categories, nil, nil, posts, mostLikedPost)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}
	if user.IsBanned {
		http.Redirect(w, r, "/ban", redirect)
	}

	for i := range posts {
		posts[i].IsLoggedIn = true
	}
	for i := range mostLikedPost {
		mostLikedPost[i].IsLoggedIn = true
	}

	if pp != nil {
		user.ProfilePicture = base64.StdEncoding.EncodeToString(pp)
	}

	serveHomePage(w, true, user.UserID, user.ProfilePicture, user.IsAdmin, user.IsBanned, categories, allCategories, categoriesFollowed, posts, mostLikedPost)
}

func serveHomePage(w http.ResponseWriter, isLoggedIn bool, userID int, pp string, isAdmin bool, isBanned bool, categories []CategoryData, allCategories []CategoryData, categoriesFollowed []CategoryData, posts []PostData, mostLikedPost []PostData) {
	userData := UserData{IsLoggedIn: isLoggedIn, UserID: userID, ProfilePicture: pp, IsAdmin: isAdmin, IsBanned: isBanned, Categories: categories, AllCategories: allCategories, CategoriesFollowed: categoriesFollowed, Posts: posts, MostLikedPosts: mostLikedPost}
	tmpl, err := template.ParseFiles("tmpl/home.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
