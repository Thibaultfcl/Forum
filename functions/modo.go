package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func ReportPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	token := GetSessionToken(r)
	if token == "" {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}
	row := db.QueryRow("SELECT isModerator FROM users WHERE UUID=?", token)
	var isModerator bool
	err := row.Scan(&isModerator)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if !isModerator {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}

	var data struct {
		UserID     int `json:"userID"`
		PostID int `json:"postID"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`INSERT INTO posts_reported (user_id, post_id) VALUES (?, ?)`, data.UserID, data.PostID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func UnreportPost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	token := GetSessionToken(r)
	if token == "" {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}
	row := db.QueryRow("SELECT isModerator FROM users WHERE UUID=?", token)
	var isModerator bool
	err := row.Scan(&isModerator)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if !isModerator {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}

	var data struct {
		UserID     int `json:"userID"`
		PostID int `json:"postID"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`DELETE FROM posts_reported WHERE user_id = ? AND post_id = ?`, data.UserID, data.PostID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func ReportComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	token := GetSessionToken(r)
	if token == "" {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}
	row := db.QueryRow("SELECT isModerator FROM users WHERE UUID=?", token)
	var isModerator bool
	err := row.Scan(&isModerator)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if !isModerator {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}

	var data struct {
		UserID     int `json:"userID"`
		CommentID int `json:"commentID"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`INSERT INTO comments_reported (user_id, comment_id) VALUES (?, ?)`, data.UserID, data.CommentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func UnreportComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	token := GetSessionToken(r)
	if token == "" {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}
	row := db.QueryRow("SELECT isModerator FROM users WHERE UUID=?", token)
	var isModerator bool
	err := row.Scan(&isModerator)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if !isModerator {
		http.Error(w, "You aren't authorized", http.StatusUnauthorized)
		return
	}

	var data struct {
		UserID     int `json:"userID"`
		CommentID int `json:"commentID"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`DELETE FROM comments_reported WHERE user_id = ? AND comment_id = ?`, data.UserID, data.CommentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}