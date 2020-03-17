package fileserver

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

const maxUploadSize = 2 * 1024 * 1024 * 1024 // 2Gb

var (
	UploadPath = "files/"
)

type lFile struct {
	Flist []os.FileInfo
	Ftype []string
}

func FileserverIndex(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "public/html/index.html")
	tmpl := template.Must(template.ParseFiles("public/html/index.html"))
	// files, err := ioutil.ReadDir(uploadPath)
	files, err := ioutil.ReadDir(UploadPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var fileType []string
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == "" {
			fileType = append(fileType, "unknown")
		} else {
			fileType = append(fileType, ext)
		}
	}
	sfile := lFile{Flist: files, Ftype: fileType}
	tmpl.Execute(w, sfile)
}

func FileserverUpload(w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFiles("public/html/upload.html"))
	files, err := ioutil.ReadDir(UploadPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var fileType []string
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == "" {
			fileType = append(fileType, "unknown")
		} else {
			fileType = append(fileType, ext)
		}
	}
	sfile := lFile{Flist: files, Ftype: fileType}
	tmpl.Execute(w, sfile)
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	reader, err := r.MultipartReader()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//copy each part to destination.
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		//if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			continue
		}
		dst, err := os.Create(UploadPath + part.FileName())
		defer dst.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("<!DOCTYPE html><html><head> <meta http-equiv='refresh' content='5; URL=/uploadpage'></head><body>SUCCESS<p></p><a href=/uploadpage>Back to previous page</a></body></html>"))
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	err := os.Remove(fmt.Sprintf("%s%s", UploadPath, name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write([]byte(fmt.Sprintf("Delete file '%s' success", name)))
}

func ReceiveFile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if _, err := os.Stat(fmt.Sprintf("%s%s", UploadPath, name)); err == nil {
		http.Error(w, "file exists", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f, err := os.OpenFile(fmt.Sprintf("%s%s", UploadPath, name), os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(f, file)
}
