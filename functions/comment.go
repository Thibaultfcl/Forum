package functions

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CommentData struct {
	Content         string
	Author          string
	AuthorPicture   string
	TimePosted      string
	Liked           bool
	Reported        bool
	NbofLikes       int
	UserIsAdmin     bool
	UserIsModerator bool
	UserID          int
	CommentID       int
	IsLoggedIn      bool
}

func CreateComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//we check if the method is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//we get the token
	token := GetSessionToken(r)

	//we get the user id
	row := db.QueryRow("SELECT id, isBanned FROM users WHERE UUID=?", token)
	var id int
	var isBanned bool
	err := row.Scan(&id, &isBanned)
	if err != nil {
		http.Error(w, fmt.Sprintf("You need to be logged in to create a post: %v", err), http.StatusUnauthorized)
		return
	}
	if isBanned {
		http.Error(w, "You are banned", http.StatusUnauthorized)
		return
	}

	// we get the comment and the post id
	comment := r.FormValue("Comment")
	postID := r.FormValue("PostID")

	_, err = db.Exec("INSERT INTO comments (user_id, post_id, content) VALUES (?, ?, ?)", id, postID, comment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating comment: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post/"+postID, redirect)
}

func DeleteComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		CommentID int `json:"commentID"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM comments WHERE id=?", data.CommentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func getComments(w http.ResponseWriter, db *sql.DB, postID int, token string) []CommentData {
	var comments []CommentData
	rows, err := db.Query(`SELECT id, user_id, content, date FROM comments WHERE post_id=? ORDER BY date DESC`, postID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting comments: %v", err), http.StatusInternalServerError)
		return comments
	}
	for rows.Next() {
		var comment CommentData
		err = rows.Scan(&comment.CommentID, &comment.UserID, &comment.Content, &comment.TimePosted)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning comments: %v", err), http.StatusInternalServerError)
			return comments
		}
		var authorStr string
		var authorPP []byte
		var AuthorIsBanned bool
		row := db.QueryRow("SELECT username, pp, isBanned FROM users WHERE id=?", comment.UserID)
		err = row.Scan(&authorStr, &authorPP, &AuthorIsBanned)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting author: %v", err), http.StatusInternalServerError)
			return nil
		}
		if AuthorIsBanned {
			continue
		}
		comment.Author = authorStr
		comment.AuthorPicture = base64.StdEncoding.EncodeToString(authorPP)

		const layout = "2006-01-02T15:04:05Z07:00"
		t, err := time.Parse(layout, comment.TimePosted)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing time: %v", err), http.StatusInternalServerError)
			return nil
		}
		t = t.Local()
		elapsed := time.Since(t)
		if elapsed < time.Minute {
			comment.TimePosted = fmt.Sprintf("%d seconds ago", int(elapsed.Seconds()))
		} else if elapsed < time.Hour {
			comment.TimePosted = fmt.Sprintf("%d minutes ago", int(elapsed.Minutes()))
		} else if elapsed < time.Hour*24 {
			comment.TimePosted = fmt.Sprintf("%d hours ago", int(elapsed.Hours()))
		} else {
			comment.TimePosted = fmt.Sprintf("%d days ago", int(elapsed.Hours()/24))
		}

		var user_id int
		if token == "" {
			comment.Liked = false
			comment.Reported = false
		} else {
			row := db.QueryRow("SELECT id, isAdmin, isModerator FROM users WHERE UUID=?", token)
			err = row.Scan(&user_id, &comment.UserIsAdmin, &comment.UserIsModerator)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error getting user id: %v", err), http.StatusInternalServerError)
				return nil
			}
			row = db.QueryRow("SELECT user_id, comment_id FROM user_liked_comments WHERE user_id=? AND comment_id=?", user_id, comment.CommentID)
			var userID, commentID int
			err = row.Scan(&userID, &commentID)
			if err != nil {
				if err == sql.ErrNoRows {
					comment.Liked = false
				} else {
					http.Error(w, fmt.Sprintf("Error getting liked comments: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				comment.Liked = true
			}
			row = db.QueryRow("SELECT comment_id FROM comments_reported WHERE comment_id = ?", comment.CommentID)
			err = row.Scan(&commentID)
			if err != nil {
				if err == sql.ErrNoRows {
					comment.Reported = false
				} else {
					http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
					return nil
				}
			} else {
				comment.Reported = true
			}
		}

		row = db.QueryRow("SELECT COUNT(*) FROM user_liked_comments WHERE comment_id=?", comment.CommentID)
		err = row.Scan(&comment.NbofLikes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting number of likes: %v", err), http.StatusInternalServerError)
			return nil
		}
		rows, err := db.Query("SELECT user_id FROM user_liked_comments WHERE comment_id=?", comment.CommentID)
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
		comment.NbofLikes -= bannedUsers

		comments = append(comments, comment)
	}
	return comments
}
