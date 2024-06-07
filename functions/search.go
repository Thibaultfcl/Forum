package functions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type Suggestion struct {
	Name string `json:"name"`
}

func GetSuggestions(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Search string `json:"search"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT name FROM categories WHERE name LIKE ? LIMIT 5", "%"+data.Search+"%")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var suggestions []Suggestion
	for rows.Next() {
		var suggestion Suggestion
		err := rows.Scan(&suggestion.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		suggestions = append(suggestions, suggestion)
	}

	if len(suggestions) < 5 {
		rows, err := db.Query("SELECT username FROM users WHERE username LIKE ? LIMIT ?", "%"+data.Search+"%", 5-len(suggestions))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var suggestion Suggestion
			err := rows.Scan(&suggestion.Name)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
				return
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestions); err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	}
}
