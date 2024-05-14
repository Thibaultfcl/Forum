package functions

import "database/sql"

// function that creat a table User
func CreateTableUser(db *sql.DB) {
	//creating the user table if not already created
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username VARCHAR(12) NOT NULL,
            password VARCHAR(12) NOT NULL,
			email TEXT NOT NULL,
			isAdmin BOOL NOT NULL DEFAULT FALSE,
			isBanned BOOL NOT NULL DEFAULT FALSE,
			pp BLOB,
			UUID VARCHAR(36) NOT NULL
        )
    `)
	if err != nil {
		panic(err.Error())
	}
}

// function that creat a table Categories
func CreateTableCategories(db *sql.DB) {
	//creating the user table if not already created
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL UNIQUE,
			number_of_posts INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		panic(err.Error())
	}
}

// function that creat a table Post
func CreateTablePost(db *sql.DB) {
	//creating the post table if not already created
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title VARCHAR(100) NOT NULL,
			content TEXT NOT NULL,
			date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			category INTEGER NOT NULL,
			author INTEGER NOT NULL,
			FOREIGN KEY(category) REFERENCES categories(id),
			FOREIGN KEY(author) REFERENCES users(id)
		)
	`)
	if err != nil {
		panic(err.Error())
	}
}

func CreateTable(db *sql.DB) {
	CreateTableUser(db)
	CreateTableCategories(db)
	CreateTablePost(db)
}