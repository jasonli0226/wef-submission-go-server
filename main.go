package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jasonli0226/wef-submission-go-server/api/handlers"
	"github.com/jasonli0226/wef-submission-go-server/api/middleware"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/projects", handlers.HandleUpload).Methods("POST")
	router.HandleFunc("/api/links", handlers.GetAllUploadLinks).Methods("GET")

	// Serve static files from the "uploads" folder
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./uploads")))

	routerWithMiddleware := middleware.CorsMiddleware(middleware.LoggingMiddleware(router))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", routerWithMiddleware)
}
