package functions

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type PostPageData struct {
	IsLoggedIn         bool
	ProfilePicture     string
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
	Post               PostData
	Comments           []CommentData
}

type PostData struct {
	Title         string
	Content       string
	Category      string
	Author        string
	AuthorPicture string
	TimePosted    string
	Liked         bool
	NbofLikes     int
	UserID        int
	PostID        int
	IsLoggedIn    bool
}

// post page
func Post(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	postIDStr := r.URL.Path[6:]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	post := getPostById(w, db, postID, token)
	categories := getCategoriesByNumberOfPost(w, db)
	categoriesFollowed := getCategoriesFollowed(w, db, token)
	comments := getComments(w, db, postID, token)

	//get the user data from the database
	row := db.QueryRow("SELECT isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var isAdmin, isBanned bool
	var pp []byte
	//scan and get the data
	err = row.Scan(&isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			servePostPage(w, false, "", categories, nil, post, comments)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}
	post.IsLoggedIn = true
	for i := range comments {
		comments[i].IsLoggedIn = true
	}

	var profilePicture string
	if pp != nil {
		profilePicture = base64.StdEncoding.EncodeToString(pp)
	}

	servePostPage(w, true, profilePicture, categories, categoriesFollowed, post, comments)
}

func servePostPage(w http.ResponseWriter, isLoggedIn bool, pp string, categories []CategoryData, categoriesFollowed []CategoryData, post PostData, comments []CommentData) {
	userData := PostPageData{IsLoggedIn: isLoggedIn, ProfilePicture: pp, Categories: categories, CategoriesFollowed: categoriesFollowed, Post: post, Comments: comments}
	tmpl, err := template.ParseFiles("tmpl/post.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, userData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

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

func getPosts(w http.ResponseWriter, db *sql.DB, token string) []PostData {
	var posts []PostData
	rows, err := db.Query(`SELECT id, title, content, date, category, author FROM posts ORDER BY date DESC`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
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

		var liked bool
		var user_id int
		if token == "" {
			liked = false
		} else {
			row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
			err := row.Scan(&user_id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", user_id, id)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				liked = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", id)
		var nbofLikes int
		err = row.Scan(&nbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		posts = append(posts, PostData{Title: title, Content: content, Category: categoryStr, Author: authorStr, AuthorPicture: base64.StdEncoding.EncodeToString(authorPP), TimePosted: elapsedStr, Liked: liked, NbofLikes: nbofLikes, UserID: user_id, PostID: id})
	}
	return posts
}

func getPostsFromUser(w http.ResponseWriter, db *sql.DB, authorID int, token string) []PostData {
	var posts []PostData
	rows, err := db.Query("SELECT id, title, content, date, category, author FROM posts WHERE author=? ORDER BY date DESC", authorID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
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

		var liked bool
		var user_id int
		if token == "" {
			liked = false
		} else {
			row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
			err := row.Scan(&user_id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", user_id, id)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				liked = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", id)
		var nbofLikes int
		err = row.Scan(&nbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		posts = append(posts, PostData{Title: title, Content: content, Category: categoryStr, Author: authorStr, AuthorPicture: base64.StdEncoding.EncodeToString(authorPP), TimePosted: elapsedStr, Liked: liked, NbofLikes: nbofLikes, UserID: user_id, PostID: id})
	}
	return posts
}

func getPostsFromCategory(w http.ResponseWriter, db *sql.DB, categoryID int, token string) []PostData {
	var posts []PostData
	rows, err := db.Query("SELECT id, title, content, date, category, author FROM posts WHERE category=? ORDER BY date DESC", categoryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
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

		var liked bool
		var user_id int
		if token == "" {
			liked = false
		} else {
			row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
			err := row.Scan(&user_id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", user_id, id)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				liked = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", id)
		var nbofLikes int
		err = row.Scan(&nbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		posts = append(posts, PostData{Title: title, Content: content, Category: categoryStr, Author: authorStr, AuthorPicture: base64.StdEncoding.EncodeToString(authorPP), TimePosted: elapsedStr, Liked: liked, NbofLikes: nbofLikes, UserID: user_id, PostID: id})
	}
	return posts
}

func getPostById(w http.ResponseWriter, db *sql.DB, id int, token string) PostData {
	var posts PostData
	row := db.QueryRow("SELECT title, content, date, category, author FROM posts WHERE id=? ORDER BY date DESC", id)
	var category, author int
	var title, content, date string
	err := row.Scan(&title, &content, &date, &category, &author)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	var categoryStr, authorStr string
	row = db.QueryRow("SELECT name FROM categories WHERE id=?", category)
	err = row.Scan(&categoryStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	row = db.QueryRow("SELECT username FROM users WHERE id=?", author)
	err = row.Scan(&authorStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	var authorPP []byte
	row = db.QueryRow("SELECT pp FROM users WHERE id=?", author)
	err = row.Scan(&authorPP)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	const layout = "2006-01-02T15:04:05Z07:00"
	t, err := time.Parse(layout, date)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
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
	var liked bool
	var user_id int
	if token == "" {
		liked = false
	} else {
		row := db.QueryRow("SELECT id FROM users WHERE UUID=?", token)
		err := row.Scan(&user_id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return PostData{}
		}
		row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", user_id, id)
		var userID, postID int
		err = row.Scan(&userID, &postID)
		if err != nil {
			if err == sql.ErrNoRows {
				liked = false
			} else {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return PostData{}
			}
		} else {
			liked = true
		}
	}
	row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", id)
	var nbofLikes int
	err = row.Scan(&nbofLikes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	posts = PostData{Title: title, Content: content, Category: categoryStr, Author: authorStr, AuthorPicture: base64.StdEncoding.EncodeToString(authorPP), TimePosted: elapsedStr, Liked: liked, NbofLikes: nbofLikes, UserID: user_id, PostID: id}
	return posts
}

func getCategoriesByNumberOfPost(w http.ResponseWriter, db *sql.DB) []CategoryData {
	var categories []CategoryData
	rows, _ := db.Query("SELECT id, name, number_of_posts FROM categories ORDER BY number_of_posts DESC LIMIT 5")
	for rows.Next() {
		var id int
		var categoryName string
		var categoryNbofP int
		err := rows.Scan(&id, &categoryName, &categoryNbofP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		categories = append(categories, CategoryData{Id: id, Name: categoryName, NbofP: categoryNbofP})
	}
	return categories
}

func getAllCategories(w http.ResponseWriter, db *sql.DB) []CategoryData {
	var categories []CategoryData
	rows, _ := db.Query("SELECT id, name, number_of_posts FROM categories ORDER BY number_of_posts DESC")
	for rows.Next() {
		var id int
		var categoryName string
		var categoryNbofP int
		err := rows.Scan(&id, &categoryName, &categoryNbofP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		categories = append(categories, CategoryData{Id: id, Name: categoryName, NbofP: categoryNbofP})
	}
	return categories
}

func getCategoryById(w http.ResponseWriter, db *sql.DB, id int) []CategoryData {
	var category []CategoryData
	row := db.QueryRow("SELECT name, number_of_posts FROM categories WHERE id=?", id)
	var name string
	var nbofP int
	err := row.Scan(&name, &nbofP)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return nil
	}
	category = append(category, CategoryData{Id: id, Name: name, NbofP: nbofP})
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
	rows, err := db.Query("SELECT category_id FROM user_liked_categories WHERE user_id=?", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return nil
	}
	for rows.Next() {
		var categoryID int
		err := rows.Scan(&categoryID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row := db.QueryRow("SELECT name, number_of_posts FROM categories WHERE id=?", categoryID)
		var name string
		var nbofP int
		err = row.Scan(&name, &nbofP)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		categories = append(categories, CategoryData{Id: categoryID, Name: name, NbofP: nbofP})
	}
	return categories
}
