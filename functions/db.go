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
			isModerator BOOL NOT NULL DEFAULT FALSE,
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
			name VARCHAR(100) NOT NULL UNIQUE
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
			image BLOB,
			FOREIGN KEY(category) REFERENCES categories(id),
			FOREIGN KEY(author) REFERENCES users(id)
		)
	`)
	if err != nil {
		panic(err.Error())
	}
}

// function that creates a table UserLikedCategories
func CreateTableUserLikedCategories(db *sql.DB) {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS user_liked_categories (
            user_id INTEGER NOT NULL,
            category_id INTEGER NOT NULL,
            PRIMARY KEY(user_id, category_id),
            FOREIGN KEY(user_id) REFERENCES users(id),
            FOREIGN KEY(category_id) REFERENCES categories(id)
        )
    `)
    if err != nil {
        panic(err.Error())
    }
}

// function that creates a table UserLikedCategories
func CreateTableUserLikedPosts(db *sql.DB) {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS user_liked_posts (
            user_id INTEGER NOT NULL,
            post_id INTEGER NOT NULL,
            PRIMARY KEY(user_id, post_id),
            FOREIGN KEY(user_id) REFERENCES users(id),
            FOREIGN KEY(post_id) REFERENCES posts(id)
        )
    `)
    if err != nil {
        panic(err.Error())
    }
}

// function that creates a table Comments
func CreateTableComments(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			post_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(post_id) REFERENCES post(id)
		)
	`)
	if err != nil {
		panic(err.Error())
	}
}

// function that creates a table UserLikedComments
func CreateTableUserLikedComments(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_liked_comments (
			user_id INTEGER NOT NULL,
			comment_id INTEGER NOT NULL,
			PRIMARY KEY(user_id, comment_id),
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(comment_id) REFERENCES comments(id)
		)
	`)
	if err != nil {
		panic(err.Error())
	}
}

func CreateTablePostReported(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts_reported (
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			PRIMARY KEY(post_id, user_id),
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		panic(err.Error())
	}
}

// function that creates all the tables
func CreateTable(db *sql.DB) {
	CreateTableUser(db)
	CreateTableCategories(db)
	CreateTablePost(db)
	CreateTableUserLikedCategories(db)
	CreateTableUserLikedPosts(db)
	CreateTableComments(db)
	CreateTableUserLikedComments(db)
	CreateTablePostReported(db)
}