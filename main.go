package main

import (
	"database/sql"
	"fmt"
	"forum/functions"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// port of the server
const port = ":8080"

func main() {
	//open the database with sqlite3
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		panic(err.Error())
	}
	//creat the tables
	functions.CreateTable(db)
	functions.CreateAdminAccount(db)
	defer db.Close()

	//handle the different pages
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) { functions.Home(w, r, db) })
	http.HandleFunc("/category/", func(w http.ResponseWriter, r *http.Request) { functions.Category(w, r, db) })
	http.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) { functions.Post(w, r, db) })
	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) { functions.User(w, r, db) })
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) { functions.Admin(w, r, db) })
	http.HandleFunc("/createPost", func(w http.ResponseWriter, r *http.Request) { functions.CreatePost(w, r, db) })
	http.HandleFunc("/deletePost", func(w http.ResponseWriter, r *http.Request) { functions.DeletePost(w, r, db) })
	http.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) { functions.CreateComment(w, r, db) })
	http.HandleFunc("/deleteComment", func(w http.ResponseWriter, r *http.Request) { functions.DeleteComment(w, r, db) })
	http.HandleFunc("/editProfile", func(w http.ResponseWriter, r *http.Request) { functions.EditProfile(w, r, db) })
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) { functions.GetSuggestions(w, r, db) })
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) { functions.Signin(w, r, db) })
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) { functions.Signup(w, r, db) })
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) { functions.Logout(w, r, db) })
	http.HandleFunc("/add-liked-category", func(w http.ResponseWriter, r *http.Request) { functions.AddLikedCategory(w, r, db) })
	http.HandleFunc("/remove-liked-category", func(w http.ResponseWriter, r *http.Request) { functions.RemoveLikedCategory(w, r, db) })
	http.HandleFunc("/add-liked-post", func(w http.ResponseWriter, r *http.Request) { functions.AddLikedPost(w, r, db) })
	http.HandleFunc("/remove-liked-post", func(w http.ResponseWriter, r *http.Request) { functions.RemoveLikedPost(w, r, db) })
	http.HandleFunc("/report-post", func(w http.ResponseWriter, r *http.Request) { functions.ReportPost(w, r, db) })
	http.HandleFunc("/unreport-post", func(w http.ResponseWriter, r *http.Request) { functions.UnreportPost(w, r, db) })
	http.HandleFunc("/add-liked-comment", func(w http.ResponseWriter, r *http.Request) { functions.AddLikedComment(w, r, db) })
	http.HandleFunc("/remove-liked-comment", func(w http.ResponseWriter, r *http.Request) { functions.RemoveLikedComment(w, r, db) })
	http.HandleFunc("/report-comment", func(w http.ResponseWriter, r *http.Request) { functions.ReportComment(w, r, db) })
	http.HandleFunc("/unreport-comment", func(w http.ResponseWriter, r *http.Request) { functions.UnreportComment(w, r, db) })
	http.HandleFunc("/switchAdminStatus", func(w http.ResponseWriter, r *http.Request) { functions.SwitchAdminStatus(w, r, db) })
	http.HandleFunc("/switchModoStatus", func(w http.ResponseWriter, r *http.Request) { functions.SwitchModoStatus(w, r, db) })
	http.HandleFunc("/switchBanStatus", func(w http.ResponseWriter, r *http.Request) { functions.SwitchBanStatus(w, r, db) })
	http.HandleFunc("/resetPP", func(w http.ResponseWriter, r *http.Request) { functions.ResetPP(w, r, db) })

	//load the CSS, the JS and the images
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))
	http.Handle("/script/", http.StripPrefix("/script/", http.FileServer(http.Dir("./script"))))

	//start the local host
	fmt.Println("\n(http://localhost:8080/home) - Server started on port", port)
	http.ListenAndServe(port, nil)
}
