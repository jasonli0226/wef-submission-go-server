package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Response struct {
	Message string `json:"message"`
}

type Project struct {
	Name      string `json:"name"`
	Link      string `json:"link"`
	UpdatedAt string `json:"updated_at"`
}

var projectDirectory string = "uploads"

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(25 << 20) // 25 MB
	if err != nil {
		RespJSON(w, http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("project")
	if err != nil {
		RespJSON(w, http.StatusBadRequest, Response{Message: "Error retrieving file from form data"})
		return
	}
	defer file.Close()

	// Create a temporary directory to extract the zip file
	err = os.MkdirAll(projectDirectory, os.ModePerm)
	if err != nil {
		RespJSON(w, http.StatusInternalServerError, Response{Message: "Error creating temporary directory"})
		return
	}

	// Save the uploaded zip file to the temporary directory
	filePath := filepath.Join(projectDirectory, handler.Filename)
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
	err = unzip(filePath, projectDirectory)
	if err != nil {
		RespJSON(w, http.StatusInternalServerError, Response{Message: "Error unzipping file"})
		return
	}

	RespJSON(w, http.StatusCreated, Response{Message: "File uploaded and unzipped successfully"})
}

func GetAllUploadLinks(w http.ResponseWriter, r *http.Request) {
	dir, err := os.Open(projectDirectory)
	if err != nil {
		RespJSON(w, http.StatusBadRequest, Response{Message: "Error reading projects"})
		return
	}
	defer dir.Close()

	// Read the directory entries
	entries, err := dir.ReadDir(0)
	if err != nil {
		RespJSON(w, http.StatusBadRequest, Response{Message: "Error reading projects"})
		return
	}

	// Print the filenames
	var projects []Project
	var prefix = "WEF_Proj_"
	for _, entry := range entries {
		if !entry.Type().IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		schema := "http"
		if r.TLS != nil {
			schema = "https"
		}

		info, _ := entry.Info()
		filename := strings.Split(entry.Name(), prefix)[1]
		name := strings.Join(strings.Split(filename, "_"), " ")
		project := Project{
			Name:      name,
			Link:      fmt.Sprintf("%s://%s/%s", schema, r.Host, entry.Name()),
			UpdatedAt: info.ModTime().Format(time.RFC3339),
		}
		projects = append(projects, project)
	}

	RespJSON(w, http.StatusOK, projects)
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
