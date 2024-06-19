package functions

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type PostPageData struct {
	IsLoggedIn         bool
	UserID             int
	ProfilePicture     string
	IsAdmin            bool
	IsBanned           bool
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
	Post               PostData
	Comments           []CommentData
}

type PostData struct {
	Title           string
	Content         string
	Category        string
	CategoryID      int
	Author          string
	AuthorID        int
	AuthorPicture   string
	TimePosted      string
	Liked           bool
	Reported        bool
	NbofLikes       int
	NbofComments    int
	UserIsAdmin     bool
	UserIsModerator bool
	UserID          int
	PostID          int
	IsLoggedIn      bool
}

// post page
func Post(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)

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
	row := db.QueryRow("SELECT id, isAdmin, isBanned, pp FROM users WHERE UUID=?", token)
	var id int
	var isAdmin, isBanned bool
	var pp []byte
	err = row.Scan(&id, &isAdmin, &isBanned, &pp)
	if err != nil {
		if err == sql.ErrNoRows {
			servePostPage(w, false, 0, "", false, false, categories, nil, post, comments)
			return
		} else {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}
	if isBanned {
		http.Redirect(w, r, "/ban", redirect)
	}

	post.IsLoggedIn = true
	for i := range comments {
		comments[i].IsLoggedIn = true
	}

	var profilePicture string
	if pp != nil {
		profilePicture = base64.StdEncoding.EncodeToString(pp)
	}

	servePostPage(w, true, id, profilePicture, isAdmin, isBanned, categories, categoriesFollowed, post, comments)
}

func servePostPage(w http.ResponseWriter, isLoggedIn bool, userID int, pp string, isAdmin bool, isBanned bool, categories []CategoryData, categoriesFollowed []CategoryData, post PostData, comments []CommentData) {
	userData := PostPageData{IsLoggedIn: isLoggedIn, UserID: userID, ProfilePicture: pp, IsAdmin: isAdmin, IsBanned: isBanned, Categories: categories, CategoriesFollowed: categoriesFollowed, Post: post, Comments: comments}
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

func DeletePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		PostID     int `json:"postID"`
		CategoryID int `json:"categoryID"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM posts WHERE id=?", data.PostID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}

	_, err = db.Exec("UPDATE categories SET number_of_posts = number_of_posts - 1 WHERE id=?", data.CategoryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while updating the number of posts: %v", err), http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT number_of_posts FROM categories WHERE id=?", data.CategoryID)
	var nbofPosts int
	err = row.Scan(&nbofPosts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if nbofPosts == 0 {
		_, err = db.Exec("DELETE FROM categories WHERE id=?", data.CategoryID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func getPosts(w http.ResponseWriter, db *sql.DB, token string) []PostData {
	var posts []PostData
	rows, err := db.Query(`SELECT id, title, content, date, category, author FROM posts ORDER BY date DESC`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
	}
	for rows.Next() {
		var post PostData
		var date string
		err := rows.Scan(&post.PostID, &post.Title, &post.Content, &date, &post.CategoryID, &post.AuthorID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		var authorPP []byte
		var authorIsBanned bool
		row := db.QueryRow("SELECT name FROM categories WHERE id=?", post.CategoryID)
		err = row.Scan(&post.Category)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row = db.QueryRow("SELECT username, pp, isBanned FROM users WHERE id=?", post.AuthorID)
		err = row.Scan(&post.Author, &authorPP, &authorIsBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		if authorIsBanned {
			continue
		}
		post.AuthorPicture = base64.StdEncoding.EncodeToString(authorPP)

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
		post.TimePosted = elapsedStr

		if token == "" {
			post.Liked = false
			post.Reported = false
		} else {
			row := db.QueryRow("SELECT id, isAdmin, isModerator FROM users WHERE UUID=?", token)
			err := row.Scan(&post.UserID, &post.UserIsAdmin, &post.UserIsModerator)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", post.UserID, post.PostID)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Liked = true
			}
			row = db.QueryRow("SELECT post_id FROM posts_reported WHERE post_id = ?", post.PostID)
			err = row.Scan(&postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Reported = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Reported = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT user_id FROM user_liked_posts WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err = rows.Scan(&userID)
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
		post.NbofLikes -= bannedUsers

		row = db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofComments)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err = db.Query("SELECT user_id FROM comments WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID2 int
		var bannedUsers2 int
		for rows.Next() {
			err = rows.Scan(&userID2)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID2)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers2++
			}
		}
		post.NbofComments -= bannedUsers2

		posts = append(posts, post)
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
		var post PostData
		var date string
		err := rows.Scan(&post.PostID, &post.Title, &post.Content, &date, &post.CategoryID, &post.AuthorID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		var authorPP []byte
		var authorIsBanned bool
		row := db.QueryRow("SELECT name FROM categories WHERE id=?", post.CategoryID)
		err = row.Scan(&post.Category)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row = db.QueryRow("SELECT username, pp, isBanned FROM users WHERE id=?", post.AuthorID)
		err = row.Scan(&post.Author, &authorPP, &authorIsBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		if authorIsBanned {
			continue
		}
		post.AuthorPicture = base64.StdEncoding.EncodeToString(authorPP)

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
		post.TimePosted = elapsedStr

		if token == "" {
			post.Liked = false
			post.Reported = false
		} else {
			row := db.QueryRow("SELECT id, isAdmin, isModerator FROM users WHERE UUID=?", token)
			err := row.Scan(&post.UserID, &post.UserIsAdmin, &post.UserIsModerator)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", post.UserID, post.PostID)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Liked = true
			}
			row = db.QueryRow("SELECT post_id FROM posts_reported WHERE post_id = ?", post.PostID)
			err = row.Scan(&postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Reported = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Reported = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT user_id FROM user_liked_posts WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err = rows.Scan(&userID)
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
		post.NbofLikes -= bannedUsers

		row = db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofComments)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err = db.Query("SELECT user_id FROM comments WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID2 int
		var bannedUsers2 int
		for rows.Next() {
			err = rows.Scan(&userID2)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID2)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers2++
			}
		}
		post.NbofComments -= bannedUsers2

		posts = append(posts, post)
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
		var post PostData
		var date string
		err := rows.Scan(&post.PostID, &post.Title, &post.Content, &date, &post.CategoryID, &post.AuthorID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		var authorPP []byte
		var authorIsBanned bool
		row := db.QueryRow("SELECT name FROM categories WHERE id=?", post.CategoryID)
		err = row.Scan(&post.Category)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row = db.QueryRow("SELECT username, pp, isBanned FROM users WHERE id=?", post.AuthorID)
		err = row.Scan(&post.Author, &authorPP, &authorIsBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		if authorIsBanned {
			continue
		}
		post.AuthorPicture = base64.StdEncoding.EncodeToString(authorPP)

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
		post.TimePosted = elapsedStr

		if token == "" {
			post.Liked = false
			post.Reported = false
		} else {
			row := db.QueryRow("SELECT id, isAdmin, isModerator FROM users WHERE UUID=?", token)
			err := row.Scan(&post.UserID, &post.UserIsAdmin, &post.UserIsModerator)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", post.UserID, post.PostID)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Liked = true
			}
			row = db.QueryRow("SELECT post_id FROM posts_reported WHERE post_id = ?", post.PostID)
			err = row.Scan(&postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Reported = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Reported = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT user_id FROM user_liked_posts WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err = rows.Scan(&userID)
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
		post.NbofLikes -= bannedUsers

		row = db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofComments)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err = db.Query("SELECT user_id FROM comments WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID2 int
		var bannedUsers2 int
		for rows.Next() {
			err = rows.Scan(&userID2)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID2)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers2++
			}
		}
		post.NbofComments -= bannedUsers2

		posts = append(posts, post)
	}
	return posts
}

func getPostById(w http.ResponseWriter, db *sql.DB, id int, token string) PostData {
	var post PostData
	row := db.QueryRow("SELECT title, content, date, category, author FROM posts WHERE id=? ORDER BY date DESC", id)
	post.PostID = id
	var date string
	err := row.Scan(&post.Title, &post.Content, &date, &post.CategoryID, &post.AuthorID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}

	var authorPP []byte
	var authorIsBanned bool
	row = db.QueryRow("SELECT name FROM categories WHERE id=?", post.CategoryID)
	err = row.Scan(&post.Category)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	row = db.QueryRow("SELECT username, pp, isBanned FROM users WHERE id=?", post.AuthorID)
	err = row.Scan(&post.Author, &authorPP, &authorIsBanned)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	if authorIsBanned {
		return PostData{}
	}
	post.AuthorPicture = base64.StdEncoding.EncodeToString(authorPP)

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
	post.TimePosted = elapsedStr

	if token == "" {
		post.Liked = false
		post.Reported = false
	} else {
		row := db.QueryRow("SELECT id, isAdmin, isModerator FROM users WHERE UUID=?", token)
		err := row.Scan(&post.UserID, &post.UserIsAdmin, &post.UserIsModerator)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return PostData{}
		}
		row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", post.UserID, post.PostID)
		var userID, postID int
		err = row.Scan(&userID, &postID)
		if err != nil {
			if err == sql.ErrNoRows {
				post.Liked = false
			} else {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return PostData{}
			}
		} else {
			post.Liked = true
		}
		row = db.QueryRow("SELECT post_id FROM posts_reported WHERE post_id = ?", post.PostID)
			err = row.Scan(&postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Reported = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return PostData{}
				}
			} else {
				post.Reported = true
			}
	}

	row = db.QueryRow("SELECT COUNT(*) FROM user_liked_posts WHERE post_id=?", post.PostID)
	err = row.Scan(&post.NbofLikes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	rows, err := db.Query("SELECT user_id FROM user_liked_posts WHERE post_id=?", post.PostID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	var userID int
	var bannedUsers int
	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return PostData{}
		}
		row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID)
		var isBanned bool
		err = row.Scan(&isBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return PostData{}
		}
		if isBanned {
			bannedUsers++
		}
	}
	post.NbofLikes -= bannedUsers

	row = db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id=?", post.PostID)
	err = row.Scan(&post.NbofComments)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	rows, err = db.Query("SELECT user_id FROM comments WHERE post_id=?", post.PostID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return PostData{}
	}
	var userID2 int
	var bannedUsers2 int
	for rows.Next() {
		err = rows.Scan(&userID2)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return PostData{}
		}
		row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID2)
		var isBanned bool
		err = row.Scan(&isBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return PostData{}
		}
		if isBanned {
			bannedUsers2++
		}
	}
	post.NbofComments -= bannedUsers2

	return post
}

func getMostLikedPosts(w http.ResponseWriter, db *sql.DB, token string) []PostData {
	var posts []PostData
	rows, err := db.Query(`
	SELECT p.id, p.title, p.content, p.date, p.category, p.author, 
		COALESCE(l.like_count, 0) as like_count
	FROM posts p
	LEFT JOIN (
		SELECT post_id, COUNT(*) as like_count 
		FROM user_liked_posts 
		GROUP BY post_id
	) l ON p.id = l.post_id
	ORDER BY like_count DESC
	`)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return posts
	}
	for rows.Next() {
		var post PostData
		var date string
		err := rows.Scan(&post.PostID, &post.Title, &post.Content, &date, &post.CategoryID, &post.AuthorID, &post.NbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}

		var authorPP []byte
		var authorIsBanned bool
		row := db.QueryRow("SELECT name FROM categories WHERE id=?", post.CategoryID)
		err = row.Scan(&post.Category)
		if (err != nil) {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row = db.QueryRow("SELECT username, pp, isBanned FROM users WHERE id=?", post.AuthorID)
		err = row.Scan(&post.Author, &authorPP, &authorIsBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		if authorIsBanned {
			continue
		}
		post.AuthorPicture = base64.StdEncoding.EncodeToString(authorPP)

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
		post.TimePosted = elapsedStr

		if token == "" {
			post.Liked = false
			post.Reported = false
		} else {
			row := db.QueryRow("SELECT id, isAdmin, isModerator FROM users WHERE UUID=?", token)
			err := row.Scan(&post.UserID, &post.UserIsAdmin, &post.UserIsModerator)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, post_id FROM user_liked_posts WHERE user_id = ? AND post_id = ?", post.UserID, post.PostID)
			var userID, postID int
			err = row.Scan(&userID, &postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Liked = true
			}
			row = db.QueryRow("SELECT post_id FROM posts_reported WHERE post_id = ?", post.PostID)
			err = row.Scan(&postID)
			if err != nil {
				if err == sql.ErrNoRows {
					post.Reported = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				post.Reported = true
			}
		}

		rows, err := db.Query("SELECT user_id FROM user_liked_posts WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID int
		var bannedUsers int
		for rows.Next() {
			err = rows.Scan(&userID)
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
		post.NbofLikes -= bannedUsers

		row = db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id=?", post.PostID)
		err = row.Scan(&post.NbofComments)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err = db.Query("SELECT user_id FROM comments WHERE post_id=?", post.PostID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		var userID2 int
		var bannedUsers2 int
		for rows.Next() {
			err = rows.Scan(&userID2)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			row := db.QueryRow("SELECT isBanned FROM users WHERE id=?", userID2)
			var isBanned bool
			err = row.Scan(&isBanned)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return nil
			}
			if isBanned {
				bannedUsers2++
			}
		}
		post.NbofComments -= bannedUsers2

		posts = append(posts, post)
	}
	return posts
}
