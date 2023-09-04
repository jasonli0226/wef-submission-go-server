package handlers

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Response struct {
	Message string `json:"message"`
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(25 << 20) // 25 MB
	if err != nil {
		RespJSON(w, http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("file")
	if err != nil {
		RespJSON(w, http.StatusBadRequest, Response{Message: "Error retrieving file from form data"})
		return
	}
	defer file.Close()

	// Create a temporary directory to extract the zip file
	tempDir := "./uploads"
	err = os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		RespJSON(w, http.StatusInternalServerError, Response{Message: "Error creating temporary directory"})
		return
	}

	// Save the uploaded zip file to the temporary directory
	filePath := filepath.Join(tempDir, handler.Filename)
	defer os.Remove(filePath)

	outFile, err := os.Create(filePath)
	if err != nil {
		RespJSON(w, http.StatusInternalServerError, Response{Message: "Error saving uploaded file"})
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, file)
	if err != nil {
		RespJSON(w, http.StatusInternalServerError, Response{Message: "Error saving uploaded file"})
		return
	}

	// Unzip the file
	err = unzip(filePath, tempDir)
	if err != nil {
		RespJSON(w, http.StatusInternalServerError, Response{Message: "Error unzipping file"})
		return
	}

	RespJSON(w, http.StatusCreated, Response{Message: "File uploaded and unzipped successfully"})
}

func GetAllUploadLinks(w http.ResponseWriter, r *http.Request) {
	RespJSON(w, http.StatusOK, Response{Message: "Dummy message"})
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
