package server

import (
	"net/http"
	"os"

	"../handlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func StartServer() {
	log.Info("Hendlers are started.")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", handlers.HomeHandlerEmpty)
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/register", handlers.RegisterHandler)
	router.HandleFunc("/bookmarks/{user}", handlers.LoggedHandler)
	router.HandleFunc("/bookmarks/", handlers.LoggedHandler)
	router.HandleFunc("/bookmarks/{user}/page={page:[0-9]+}", handlers.LoggedHandler)
	router.HandleFunc("/bookmarks/{user}/page={page:[0-9]+}/tagcloud={tagcloud:[0-9]+}", handlers.LoggedHandler)
	router.HandleFunc("/bookmarks/{user}/page={page:[0-9]+}/tag={tag}", handlers.TagsHandler)
	router.HandleFunc("/bookmarks/{user}/page={page:[0-9]+}/tag={tag}/tagcloud={tagcloud:[0-9]+}", handlers.TagsHandler)
	router.HandleFunc("/new_bookmark/", handlers.NewBookmarkHandler)
	router.HandleFunc("/edit_bookmark/bookmarkID={bookmarkID:[0-9]+}", handlers.EditBookmarkHandler)
	router.HandleFunc("/delete_bookmark/bookmarkID={bookmarkID:[0-9]+}", handlers.DeleteBookmarkHandler)
	router.HandleFunc("/search/page={page:[0-9]+}", handlers.SearchHandler)
	router.HandleFunc("/search/string={search}/page={page:[0-9]+}/", handlers.SearchHandler)
	router.HandleFunc("/search/string={search}/page={page:[0-9]+}/tagcloud={tagcloud:[0-9]+}", handlers.SearchHandler)
	router.HandleFunc("/search/page={page:[0-9]+}/", handlers.SearchHandler)
	router.PathPrefix("/Required/").Handler(http.StripPrefix("/Required/", http.FileServer(http.Dir("./Required/"))))
	http.ListenAndServe(":8080", router)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.WarnLevel)
}
