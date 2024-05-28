package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// category page
func Category(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	categoryName := r.URL.Path[10:]
	row := db.QueryRow("SELECT id FROM categories WHERE name=?", categoryName)
	var categoryID int
	err := row.Scan(&categoryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	posts := getPostsFromCategory(w, db, categoryID)
	category := getCategoryById(w, db, categoryID)
	categories := getCategoriesByNumberOfPost(w, db)
	allCategories := getAllCategories(w, db)

	//get the user data from the database
	row = db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var isAdmin, isBanned bool
	var pp []byte
	//scan and get the data
	err = row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveCategoryPage(w, false, "", categories, nil, posts, nil)
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

	serveCategoryPage(w, true, profilePicture, categories, allCategories, posts, category)
}

func serveCategoryPage(w http.ResponseWriter, isLoggedIn bool, pp string, categories []CategoryData, allCategories []CategoryData, posts []PostData, category []CategoryData) {
	userData := UserData{IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories, AllCategories: allCategories, Posts: posts, Category: category[0]}
	tmpl, err := template.ParseFiles("tmpl/category.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
