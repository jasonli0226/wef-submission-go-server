package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/upload", handleUpload)

	routerWithMiddleware := loggingMiddleware(router)
	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", routerWithMiddleware)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, Response{Message: "Method not allowed"})
		return
	}

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{Message: "Error retrieving file from form data"})
		return
	}
	defer file.Close()

	// Create a temporary directory to extract the zip file
	tempDir := "./uploads"
	err = os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{Message: "Error creating temporary directory"})
		return
	}

	// Save the uploaded zip file to the temporary directory
	filePath := filepath.Join(tempDir, handler.Filename)
	defer os.Remove(filePath)

	outFile, err := os.Create(filePath)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{Message: "Error saving uploaded file"})
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, file)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{Message: "Error saving uploaded file"})
		return
	}

	// Unzip the file
	err = unzip(filePath, tempDir)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, Response{Message: "Error unzipping file"})
		return
	}

	sendJSONResponse(w, http.StatusCreated, Response{Message: "File uploaded and unzipped successfully"})
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
		} else {
			os.MkdirAll(filepath.Dir(path), os.ModePerm)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(f, rc)
			if err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
		rc.Close()
	}

	return nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request details
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
