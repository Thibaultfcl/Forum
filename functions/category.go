package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type CategoryPageData struct {
	Id                 int
	IsLoggedIn         bool
	UserLiked          bool
	ProfilePicture     string
	IsAdmin            bool
	IsBanned           bool
	Category           CategoryData
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
	AllCategories      []CategoryData
	Posts              []PostData
}

type CategoryData struct {
	Id    int
	Name  string
	NbofP int
}

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

	posts := getPostsFromCategory(w, db, categoryID, token)
	category := getCategoryById(w, db, categoryID)
	categories := getCategoriesByNumberOfPost(w, db)
	allCategories := getAllCategories(w, db)
	categoriesFollowed := getCategoriesFollowed(w, db, token)

	//get the user data from the database
	row = db.QueryRow("SELECT id, isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var id int
	var isAdmin, isBanned bool
	var pp []byte
	//scan and get the data
	err = row.Scan(&id, &isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			serveCategoryPage(w, 0, false, false, "", false, false, categories, nil, nil, posts, category)
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

	row = db.QueryRow(("SELECT user_id, category_id FROM user_liked_categories WHERE user_id = ? AND category_id = ?"), id, categoryID)
	var userID int
	var categoryID2 int
	err = row.Scan(&userID, &categoryID2)
	if err != nil {
		if err == sql.ErrNoRows {
			serveCategoryPage(w, id, true, false, profilePicture, isAdmin, isBanned, categories, categoriesFollowed, allCategories, posts, category)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	serveCategoryPage(w, id, true, true, profilePicture, isAdmin, isBanned, categories, categoriesFollowed, allCategories, posts, category)
}

func serveCategoryPage(w http.ResponseWriter, userID int, isLoggedIn bool, userLiked bool, pp string, isAdmin bool, isBanned bool, categories []CategoryData, categoriesFollowed []CategoryData, allCategories []CategoryData, posts []PostData, category []CategoryData) {
	userData := CategoryPageData{Id: userID, IsLoggedIn: isLoggedIn, UserLiked: userLiked, ProfilePicture: pp, IsAdmin: isAdmin, IsBanned: isBanned, Categories: categories, CategoriesFollowed: categoriesFollowed, AllCategories: allCategories, Posts: posts, Category: category[0]}
	tmpl, err := template.ParseFiles("tmpl/category.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
