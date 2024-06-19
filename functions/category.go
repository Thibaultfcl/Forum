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

	categoryName := r.URL.Path[10:]
	row := db.QueryRow("SELECT id FROM categories WHERE name=?", categoryName)
	var categoryID int
	err := row.Scan(&categoryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	posts := getPostsFromCategory(w, db, categoryID, token)
	category := getCategoryById(w, r, db, categoryID)
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
	if isBanned {
		http.Redirect(w, r, "/ban", redirect)
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

func serveCategoryPage(w http.ResponseWriter, userID int, isLoggedIn bool, userLiked bool, pp string, isAdmin bool, isBanned bool, categories []CategoryData, categoriesFollowed []CategoryData, allCategories []CategoryData, posts []PostData, category CategoryData) {
	userData := CategoryPageData{Id: userID, IsLoggedIn: isLoggedIn, UserLiked: userLiked, ProfilePicture: pp, IsAdmin: isAdmin, IsBanned: isBanned, Categories: categories, CategoriesFollowed: categoriesFollowed, AllCategories: allCategories, Posts: posts, Category: category}
	tmpl, err := template.ParseFiles("tmpl/category.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func getCategoriesByNumberOfPost(w http.ResponseWriter, db *sql.DB) []CategoryData {
	var categories []CategoryData
	rows, err := db.Query(`
        SELECT c.id, c.name, COUNT(p.id) as post_count
        FROM categories c
        LEFT JOIN posts p ON p.category = c.id
        GROUP BY c.id, c.name
        ORDER BY post_count DESC
        LIMIT 5
    `)
	if err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
        return nil
    }
    defer rows.Close()

	for rows.Next() {
		var category CategoryData
		err := rows.Scan(&category.Id, &category.Name, &category.NbofP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT author FROM posts WHERE category=?", category.Id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err := rows.Scan(&userID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers++
			}
		}
		category.NbofP -= bannedUsers
		if category.NbofP <= 0 {
			continue
		}
		
		categories = append(categories, category)
	}
	return categories
}

func getAllCategories(w http.ResponseWriter, db *sql.DB) []CategoryData {
	var categories []CategoryData
	rows, err := db.Query(`
        SELECT c.id, c.name, COUNT(p.id) as post_count
        FROM categories c
        LEFT JOIN posts p ON p.category = c.id
        GROUP BY c.id, c.name
        ORDER BY post_count DESC
    `)
	if err != nil {
        http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
        return nil
    }
    defer rows.Close()

	for rows.Next() {
		var category CategoryData
		err := rows.Scan(&category.Id, &category.Name, &category.NbofP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT author FROM posts WHERE category=?", category.Id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err := rows.Scan(&userID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers++
			}
		}
		category.NbofP -= bannedUsers
		if category.NbofP <= 0 {
			continue
		}
		
		categories = append(categories, category)
	}
	return categories
}

func getCategoryById(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) CategoryData {
	var category CategoryData
	row := db.QueryRow(`
        SELECT c.id, c.name, COUNT(p.id) as post_count
        FROM categories c
        LEFT JOIN posts p ON p.category = c.id
        WHERE c.id = ?
        GROUP BY c.id, c.name
    `, id)
	err := row.Scan(&category.Id, &category.Name, &category.NbofP)
	if err != nil {
		if err == sql.ErrNoRows {
            http.Error(w, "Category not found", http.StatusNotFound)
            return CategoryData{}
        }
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return CategoryData{}
	}
	rows, err := db.Query("SELECT author FROM posts WHERE category=?", category.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return CategoryData{}
	}
	var userID int
	var bannedUsers int
	for rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return CategoryData{}
		}
		row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID)
		var isBanned bool
		err = row.Scan(&isBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return CategoryData{}
		}
		if isBanned {
			bannedUsers++
		}
	}
	category.NbofP -= bannedUsers
	if category.NbofP <= 0 {
		http.Redirect(w, r, "/home", redirect)
	}
	return category
}

func getCategoriesFollowed(w http.ResponseWriter, db *sql.DB, token string) []CategoryData {
	var categories []CategoryData
	if token == "" {
		return categories
	}

	row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
	var id int
	err := row.Scan(&id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return nil
	}

	rows, err := db.Query(`
        SELECT c.id, c.name, COUNT(p.id) as post_count
        FROM categories c
        JOIN user_liked_categories ulc ON ulc.category_id = c.id
        LEFT JOIN posts p ON p.category = c.id
        WHERE ulc.user_id = ?
        GROUP BY c.id, c.name
    `, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return nil
	}
	for rows.Next() {
		var category CategoryData
		err := rows.Scan(&category.Id, &category.Name, &category.NbofP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT author FROM posts WHERE category=?", category.Id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err := rows.Scan(&userID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers++
			}
		}
		category.NbofP -= bannedUsers
		if category.NbofP <= 0 {
			continue
		}
		categories = append(categories, category)
	}
	return categories
}