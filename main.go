package main

import (
	"log"
	"net/http"

	"github.com/jasonli0226/wef-submission-go-server/api/handlers"
	"github.com/jasonli0226/wef-submission-go-server/api/middleware"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/upload", handlers.HandleUpload)

	routerWithMiddleware := middleware.LoggingMiddleware(router)
	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", routerWithMiddleware)
}
