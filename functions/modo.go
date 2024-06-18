package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type Report struct {
	ID   int
	Modo string
}

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
		UserID int `json:"userID"`
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
		UserID int `json:"userID"`
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

func getReport(w http.ResponseWriter, db *sql.DB) []Report {
	rows, err := db.Query("SELECT post_id, user_id FROM posts_reported")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	var Reports []Report
	for rows.Next() {
		var Report Report
		var modoID int
		err = rows.Scan(&Report.ID, &modoID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		row := db.QueryRow("SELECT username FROM users WHERE id=?", modoID)
		err = row.Scan(&Report.Modo)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return nil
		}
		Reports = append(Reports, Report)
	}

	return Reports
}

func DeleteReport(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	token := GetSessionToken(r)
	if token == "" {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	row := db.QueryRow("SELECT isAdmin FROM users WHERE UUID=?", token)
	var isAdmin bool
	err := row.Scan(&isAdmin)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var data struct {
		PostID int `json:"postID"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM posts_reported WHERE post_id=?", data.PostID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}