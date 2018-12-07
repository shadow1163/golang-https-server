package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 200 * 1024 * 1024 // 200 mb
const uploadPath = "/server/files/"
const jsPath = "/server/js/"
const cssPath = "/server/css/"
const appFolder = "/server"

type lFile struct {
	Flist []os.FileInfo
}

func main() {
	http.HandleFunc("/upload", uploadFileHandler())

	fs := http.FileServer(http.Dir(uploadPath))
	jsfs := http.FileServer(http.Dir(jsPath))
	cssfs := http.FileServer(http.Dir(cssPath))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
	http.Handle("/js/", http.StripPrefix("/js/", jsfs))
	http.Handle("/css/", http.StripPrefix("/css/", cssfs))
	http.Handle("/", indexPageHandler())

	log.Print("Server started on localhost:80, use /upload for uploading files and /files/{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func indexPageHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("/server/index.html"))
		files, err := ioutil.ReadDir(uploadPath)
		if err != nil {
			renderError(w, err.Error(), http.StatusBadRequest)
			return
		}
		sfile := lFile{Flist: files}
		tmpl.Execute(w, sfile)
	})
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		//fileType := r.PostFormValue("type")
		file, handler, err := r.FormFile("uploadFile")
		if err != nil {
			log.Println(err)
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println(err)
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		//filetype := http.DetectContentType(fileBytes)
		//switch filetype {
		//case "image/jpeg", "image/jpg":
		//case "image/gif", "image/png":
		//case "application/pdf":
		//	break
		//default:
		//	renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
		//	return
		//}
		//fileName := randToken(12)
		//fileEndings, err := mime.ExtensionsByType(fileType)
		//if err != nil {
		//	renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		//	return
		//}
		newPath := filepath.Join(uploadPath, handler.Filename)
		fmt.Printf("File: %s\n", newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			log.Println(err)
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			log.Println(err)
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SUCCESS"))
	})
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
