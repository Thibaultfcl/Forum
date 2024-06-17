package functions

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

type AdminData struct {
	UserID             int
	ProfilePicture     string
	Users              []UserTable
	Categories         []CategoryData
	CategoriesFollowed []CategoryData
}

type UserTable struct {
	ID             int
	Username       string
	IsAdmin        bool
	IsModerator    bool
	IsBanned       bool
	ProfilePicture string
}

// It creates the admin account if it does not exist
func CreateAdminAccount(db *sql.DB) error {
	row := db.QueryRow("SELECT id FROM users WHERE isAdmin=?", true)
	var id int
	err := row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == nil {
		return fmt.Errorf("admin account already exists")
	}

	//hash the password
	password, err := HashPassword("password")
	if err != nil {
		return fmt.Errorf("error hashing the password: %v", err)
	}

	//open the default profile picture
	file, err := os.Open("img/profileDefault.jpg")
	if err != nil {
		return fmt.Errorf("error opening image: %v", err)
	}
	defer file.Close()

	//read the image
	ppDefault, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading image: %v", err)
	}

	//generate a new token
	newToken, err := GenerateSessionToken()
	if err != nil {
		return err
	}

	//insert the admin account in the database
	_, err = db.Exec("INSERT INTO users (username, email, password, isAdmin, isModerator, isBanned, pp, UUID) VALUES (?, ?, ?, TRUE, FALSE, FALSE, ?, ?)", "admin", "admin@gmail.com", password, ppDefault, newToken)
	if err != nil {
		return fmt.Errorf("error while creating the admin account: %v", err)
	}

	return nil
}

func Admin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	token := GetSessionToken(r)
	fmt.Println("token: ", token)

	//check if the user is logged in
	if token == "" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	//check if the user is an admin
	row := db.QueryRow("SELECT isAdmin FROM users WHERE UUID=?", token)
	var isAdmin bool
	err := row.Scan(&isAdmin)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	//get the admin data
	row = db.QueryRow("SELECT id, pp FROM users WHERE UUID=?", token)
	var adminID int
	var adminPPbyte []byte
	err = row.Scan(&adminID, &adminPPbyte)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	var adminPP string
	if adminPPbyte != nil {
		adminPP = base64.StdEncoding.EncodeToString(adminPPbyte)
	}

	//get the categories
	categories := getCategoriesByNumberOfPost(w, db)
	categoriesFollowed := getCategoriesFollowed(w, db, token)

	//get the users
	users := []UserTable{}
	rows, err := db.Query("SELECT id, username, isAdmin, isBanned, pp FROM users ORDER BY id DESC")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var user UserTable
		var pp []byte
		err := rows.Scan(&user.ID, &user.Username, &user.IsAdmin, &user.IsBanned, &pp)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		if pp != nil {
			user.ProfilePicture = base64.StdEncoding.EncodeToString(pp)
		}
		users = append(users, user)
	}

	// create the template
	adminData := AdminData{UserID: adminID, ProfilePicture: adminPP, Users: users, Categories: categories, CategoriesFollowed: categoriesFollowed}
	tmpl, err := template.ParseFiles("tmpl/admin.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, adminData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func SwitchAdminStatus(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
		UserID int `json:"userID"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	row = db.QueryRow("SELECT isAdmin FROM users WHERE id=?", data.UserID)
	err = row.Scan(&isAdmin)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if isAdmin {
		_, err = db.Exec("UPDATE users SET isAdmin = FALSE WHERE id=?", data.UserID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.Exec("UPDATE users SET isAdmin = TRUE WHERE id=?", data.UserID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func SwitchBanStatus(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
		UserID int `json:"userID"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	row = db.QueryRow("SELECT isBanned FROM users WHERE id=?", data.UserID)
	var isBanned bool
	err = row.Scan(&isBanned)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
	if isBanned {
		_, err = db.Exec("UPDATE users SET isBanned = FALSE WHERE id=?", data.UserID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		_, err = db.Exec("UPDATE users SET isBanned = TRUE WHERE id=?", data.UserID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func ResetPP(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
		UserID int `json:"userID"`
	}
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	file, err := os.Open("img/profileDefault.jpg")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening image: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	ppDefault, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading image: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET pp = ? WHERE id=?", ppDefault, data.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}
}