package handler

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strings"
)

func InitRouter() http.Handler {
	if len(mySigningKey) < 32 {
		panic("Set JWT_SECRET to a random value of at least 32 bytes")
	}
	router := mux.NewRouter()
	router.Handle("/upload", requireAuth(uploadHandler)).Methods("POST")
	router.Handle("/search", requireAuth(searchHandler)).Methods("GET")
	router.Handle("/events", requireAuth(eventsHandler)).Methods("GET")
	router.Handle("/post/{id}", requireAuth(deleteHandler)).Methods("DELETE")
	router.Handle("/post/{id}", requireAuth(editHandler)).Methods("PATCH")
	router.Handle("/post/{id}/like", requireAuth(likeHandler)).Methods("PUT")
	router.Handle("/generate", requireAuth(imageHandler)).Methods("POST")
	router.Handle("/signout", requireAuth(signoutHandler)).Methods("POST")
	router.HandleFunc("/signup", signupHandler).Methods("POST")
	router.HandleFunc("/signin", signinHandler).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok", "version": "socialai-teal-v1"})
	}).Methods("GET")
	origins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if value := os.Getenv("ALLOWED_ORIGINS"); value != "" {
		origins = strings.Split(value, ",")
	}
	return handlers.CORS(handlers.AllowedOrigins(origins), handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}), handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PATCH", "PUT"}))(router)
}
