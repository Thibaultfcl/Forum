package functions

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
)

func EditProfile(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	err := row.Scan(&id)
	if err != nil {
		http.Error(w, fmt.Sprintf("User not found: %v", err), http.StatusNotFound)
		return
	}

	//we parse the form
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusInternalServerError)
		return
	}

	//we get the form values
	file, _, err := r.FormFile("File")
	if err != nil {
		if err != http.ErrMissingFile {
			http.Error(w, fmt.Sprintf("Error retrieving the file: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Read the file content and update the profile picture
	if file != nil {
		defer file.Close()
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading the file: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("UPDATE users SET pp=? WHERE id=?", fileBytes, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating the profile picture: %v", err), http.StatusInternalServerError)
			return
		}
	}

	//we get the username and update it
	username := r.FormValue("Username")
	if username != "" {
		_, err = db.Exec("UPDATE users SET username=? WHERE id=?", username, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating the username: %v", err), http.StatusInternalServerError)
			return
		}
	}

	//we get the email and update it
	email := r.FormValue("Email")
	if email != "" {
		row := db.QueryRow("SELECT id FROM users WHERE email=?", email)
		var id int
		err := row.Scan(&id)
		if err != sql.ErrNoRows {
			if err != nil {
				http.Error(w, fmt.Sprintf("Error checking the email: %v", err), http.StatusInternalServerError)
				return
			}
			http.Error(w, "This email has already been used, please select another one", http.StatusBadRequest)
			return
		}
		_, err = db.Exec("UPDATE users SET email=? WHERE id=?", email, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating the email: %v", err), http.StatusInternalServerError)
			return
		}
	}

	//we get the password and update it
	password := r.FormValue("Password")
	confirmpassword := r.FormValue("ConfirmPassword")
	if password != "" && password == confirmpassword {
		password, err = HashPassword(password)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("UPDATE users SET password=? WHERE id=?", password, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating the password: %v", err), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/user/" + fmt.Sprint(id), redirect)
}
