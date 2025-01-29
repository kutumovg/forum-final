package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// initDB initializes the database connection and creates tables if they don't exist.
func initDB() (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	createTables(db)

	return db, nil
}

// createTables defines the SQL schema for the forum database and creates tables if they don't exist.
func createTables(db *sql.DB) {
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE,
        username TEXT UNIQUE,
        password TEXT
    );`

	createPostsTable := `
    CREATE TABLE IF NOT EXISTS posts (
        id TEXT PRIMARY KEY,
        user_id TEXT,
        content TEXT,
        created_at DATETIME,
        likes INTEGER DEFAULT 0,
        dislikes INTEGER DEFAULT 0,
		image_path TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

	createPostLikesTable := `
    CREATE TABLE IF NOT EXISTS post_likes (
        id TEXT PRIMARY KEY,
        user_id TEXT,
        post_id TEXT,
        is_like BOOLEAN,
		created_at DATETIME,
        FOREIGN KEY (user_id) REFERENCES users(id),
        FOREIGN KEY (post_id) REFERENCES posts(id),
        UNIQUE (user_id, post_id)
    );`

	createCommentsTable := `
    CREATE TABLE IF NOT EXISTS comments (
        id TEXT PRIMARY KEY,
        post_id TEXT,
        user_id TEXT,
        content TEXT,
        created_at DATETIME,
        likes INTEGER DEFAULT 0,
        dislikes INTEGER DEFAULT 0,
        FOREIGN KEY (post_id) REFERENCES posts(id),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

	createCommentLikesTable := `
    CREATE TABLE IF NOT EXISTS comment_likes (
        id TEXT PRIMARY KEY,
        user_id TEXT,
        comment_id TEXT,
        is_like BOOLEAN,
		created_at DATETIME,
        FOREIGN KEY (user_id) REFERENCES users(id),
        FOREIGN KEY (comment_id) REFERENCES comments(id),
        UNIQUE (user_id, comment_id)
    );`

	createCategoriesTable := `
    CREATE TABLE IF NOT EXISTS categories (
        id TEXT PRIMARY KEY,
        name TEXT UNIQUE
    );`

	createPostCategoriesTable := `
	CREATE TABLE IF NOT EXISTS post_categories (
		post_id TEXT,
		category_id TEXT,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);`

	createNotificationsTable := `
	CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	author_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    post_id TEXT,
    comment_id TEXT,
    action TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	is_read BOOLEAN DEFAULT FALSE,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(post_id) REFERENCES posts(id),
    FOREIGN KEY(comment_id) REFERENCES comments(id)
	);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at);
	`

	// Execute the table creation commands
	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createCategoriesTable)
	if err != nil {
		log.Fatal(err)
	}
	seedCategories(db)

	_, err = db.Exec(createPostCategoriesTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createPostsTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createPostLikesTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createCommentsTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createCommentLikesTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createNotificationsTable)
	if err != nil {
		log.Fatal(err)
	}
}

// seedCategories inserts default categories into the categories table.
func seedCategories(db *sql.DB) {
	categories := []string{"Autobiography", "Comedy", "Science Fiction", "Fantasy", "Mystery", "Other", "NSFW"}

	for _, category := range categories {
		_, err := db.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			log.Fatal(err)
		}
	}
}
