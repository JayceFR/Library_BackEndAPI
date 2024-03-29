package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	handlers "main/api/handlers"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string //The port to run the API on.
}

type apiFunc func(http.ResponseWriter, *http.Request) error

// Utility function to convert object to JSON
func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// Hold any write/read errors
type ApiError struct {
	Error string
}

// Convert the function signature to the required.
func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			//handle the error
			fmt.Println(err)
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// Constructor to create a new object of type ApiServer
func NewApiServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

// Procedure to add Cors support to the API.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow credentials (cookies, headers, etc.) to be sent
		w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws://localhost:8080")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Links endpoints to their respective functions.
func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.Use(enableCORS)
	ApiHandler := handlers.New()
	log.Println("Api running on port :", s.listenAddr)
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/account", makeHttpHandleFunc(ApiHandler.HandleAccount))
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(ApiHandler.HandleSpecificAccount))
	router.HandleFunc("/login", makeHttpHandleFunc(ApiHandler.HandleLogin))
	router.HandleFunc("/community", makeHttpHandleFunc(ApiHandler.HandleComms))
	router.HandleFunc("/community/{id}", makeHttpHandleFunc(ApiHandler.HandleSpecificComm))
	router.HandleFunc("/ws", ApiHandler.WebsocketHandler)
	router.HandleFunc("/search", ApiHandler.SearchWebsocketHandler)
	router.HandleFunc("/messages", makeHttpHandleFunc(ApiHandler.HandleMessages))
	router.HandleFunc("/chat/{id}", makeHttpHandleFunc(ApiHandler.HandleSpecificMessage))
	router.HandleFunc("/active", ApiHandler.ActiveWebsocketHandler)
	router.HandleFunc("/books", makeHttpHandleFunc(ApiHandler.HandleBooks))
	router.HandleFunc("/book/{id}", makeHttpHandleFunc(ApiHandler.HandleSpecificBook))
	router.HandleFunc("/images", makeHttpHandleFunc(ApiHandler.HandleImages))
	router.HandleFunc("/bookuser/{id}", makeHttpHandleFunc(ApiHandler.HandleBookUser))
	router.HandleFunc("/notifications/{id}", makeHttpHandleFunc(ApiHandler.HandleNotifications))
	//Run the API on the port
	http.ListenAndServe(s.listenAddr, router)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws://localhost:8080")
}
