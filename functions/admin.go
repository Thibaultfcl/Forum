package functions

import (
	"database/sql"
	"fmt"
	"io"
	"os"
)

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
	_, err = db.Exec("INSERT INTO users (username, email, password, isAdmin, isBanned, pp, UUID) VALUES (?, ?, ?, TRUE, FALSE, ?, ?)", "admin", "admin@gmail.com", password, ppDefault, newToken)
	if err != nil {
		return fmt.Errorf("error while creating the admin account: %v", err)
	}

	return nil
}
