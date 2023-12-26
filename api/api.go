package api

import (
	"encoding/json"
	"log"
	"net/http"

	handlers "main/api/handlers"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type ApiError struct {
	Error string
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			//handle the error
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func NewApiServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

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
	router.HandleFunc("/ws", ApiHandler.WebsocketHandler)
	router.HandleFunc("/search", ApiHandler.SearchWebsocketHandler)
	http.ListenAndServe(s.listenAddr, router)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws://localhost:8080")
}
