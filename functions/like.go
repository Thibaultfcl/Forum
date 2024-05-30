package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func AddLikedCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		UserID     int `json:"userID"`
		CategoryID int `json:"categoryID"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`INSERT INTO user_liked_categories (user_id, category_id) VALUES (?, ?)`, data.UserID, data.CategoryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func RemoveLikedCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		UserID     int `json:"userID"`
		CategoryID int `json:"categoryID"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`DELETE FROM user_liked_categories WHERE user_id = ? AND category_id = ?`, data.UserID, data.CategoryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func AddLikedPost(w http.ResponseWriter, db *sql.DB, userID int, postID int) {
	_, err := db.Exec(`INSERT INTO user_liked_posts (user_id, post_id) VALUES (?, ?)`, userID, postID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}

func RemoveLikedPost(w http.ResponseWriter, db *sql.DB, userID int, postID int) {
	_, err := db.Exec(`DELETE FROM user_liked_posts WHERE user_id = ? AND post_id = ?`, userID, postID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}
