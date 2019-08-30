package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	pb "../note/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

const maxUploadSize = 200 * 1024 * 1024 // 200 mb
const uploadPath = "/server/files/"
const jsPath = "/server/js/"
const cssPath = "/server/css/"
const appFolder = "/server"

type lFile struct {
	Flist []os.FileInfo
}

//Credentials Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Note note struct
type Note struct {
	ID          string `jsion:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	//2006-01-02T15:04:05.000Z
	CreatedOn time.Time `json:"createdon"`
}

func newNote() pb.NoteServiceServer {
	return new(Note)
}

// Get get all notes
func (s *Note) Get(ctx context.Context, msg *pb.Message) (*pb.MessageArray, error) {
	var messages []*pb.Message
	for _, v := range noteStore {
		message := &pb.Message{Id: v.ID, Title: v.Title, Description: v.Description}
		messages = append(messages, message)
	}
	return &pb.MessageArray{Messages: messages}, nil
}

// Post post a note
func (s *Note) Post(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Post: %s", msg))
	return msg, nil
}

// Put modify a known note
func (s *Note) Put(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Put: %s", msg))
	return msg, nil
}

// Delete delete a note
func (s *Note) Delete(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Delete: %s", msg))
	return msg, nil
}

//Message message struct
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{}

//Store for the Notes collection
var noteStore = make(map[string]Note)

var cache redis.Conn

var user = []byte("user")
var passwd = []byte("password")

func initCache() {
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	cache = conn
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// see @andreiavrammsd comment: often 307 > 301
		http.StatusTemporaryRedirect)
}

//GetNoteHandler HTTP Get - /api/notes
func GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	var notes []Note
	for _, v := range noteStore {
		notes = append(notes, v)
	}
	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(notes)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

//PutNoteHandler HTTP Put - /api/notes/{id}
func PutNoteHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	k := vars["id"]
	var noteToUpd Note
	// Decode the incoming Note json
	err = json.NewDecoder(r.Body).Decode(&noteToUpd)
	if err != nil {
		panic(err)
	}
	if note, ok := noteStore[k]; ok {
		noteToUpd.CreatedOn = note.CreatedOn
		//delete existing item and add the updated item
		delete(noteStore, k)
		noteStore[k] = noteToUpd
	} else {
		log.Printf("Could not find key of Note %s to delete", k)
	}
	w.WriteHeader(http.StatusNoContent)
}

//PostNoteHandler HTTP Post - /api/notes
func PostNoteHandler(w http.ResponseWriter, r *http.Request) {
	var note Note
	// Decode the incoming Note json
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		panic(err)
	}

	note.CreatedOn = time.Now()
	k := note.ID
	noteStore[k] = note

	j, err := json.Marshal(note)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(j)
}

//DeleteNoteHandler HTTP Delete - /api/notes/{id}
func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k := vars["id"]
	// Remove from Store
	if _, ok := noteStore[k]; ok {
		//delete existing item
		delete(noteStore, k)
	} else {
		log.Printf("Could not find key of Note %s to delete", k)
	}
	w.WriteHeader(http.StatusNoContent)
}

// BasicAuth basic auth function
func BasicAuth(f func(http.ResponseWriter, *http.Request), user, passwd []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		basicAuthPrefix := "Basic "
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			payload, err := base64.StdEncoding.DecodeString(
				auth[len(basicAuthPrefix):],
			)
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && bytes.Equal(pair[0], user) &&
					bytes.Equal(pair[1], passwd) {
					f(w, r)
					return
				}
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// Signin sign in
func Signin(w http.ResponseWriter, r *http.Request) {
	// var creds Credentials
	// Get the JSON body and decode into credentials
	//err := json.NewDecoder(r.Body).Decode(&creds)
	//if err != nil {
	// If the structure of the body is wrong, return an HTTP error
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}
	name := r.FormValue("username")
	pass := r.FormValue("password")

	// Get the expected password from our in memory map
	// expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if name != string(user) || string(passwd) != pass {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("<!DOCTYPE html><html><head> <meta http-equiv='refresh' content='5; URL=/'></head><body>Error 401<p></p><a href=/>Back to previous page</a></body></html>"))
		return
	}

	// Create a new random session token
	obj, _ := uuid.NewV4()
	sessionToken := obj.String()
	log.Println(sessionToken)
	log.Println(name)
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err := cache.Do("SETEX", sessionToken, "120", name)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<!DOCTYPE html><html><head> <meta http-equiv='refresh' content='5; URL=/'></head><body>Error 500<p></p><a href=/>Back to previous page</a></body></html>"))
		return
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds, the same as the cache
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
	http.Redirect(w, r, "/uploadpage", 302)
}

// Welcome welcome page
func UploadPage(w http.ResponseWriter, r *http.Request) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		// If the session token is not present in cache, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Finally, return the welcome message to the user
	tmpl := template.Must(template.ParseFiles("/server/upload.html"))
	files, err := ioutil.ReadDir(uploadPath)
	if err != nil {
		renderError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sfile := lFile{Flist: files}
	tmpl.Execute(w, sfile)
}

//Refresh refresh token
func Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// The code uptil this point is the same as the first part of the `Welcome` route

	// Now, create a new session token for the current user
	obj, _ := uuid.NewV4()
	newSessionToken := obj.String()
	_, err = cache.Do("SETEX", newSessionToken, "120", fmt.Sprintf("%s", response))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete the older session token
	_, err = cache.Do("DEL", sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
	http.Redirect(w, r, "/uploadpage", 302)
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	clients[ws] = true
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func chatroom(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/server/chatroom.html")
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		log.Println(fmt.Sprintf("Not found: %s", r.URL.Path))
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join("/note/proto/", p)
	http.ServeFile(w, r, p)
}

func main() {
	initCache()

	r := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir(uploadPath))
	jsfs := http.FileServer(http.Dir(jsPath))
	cssfs := http.FileServer(http.Dir(cssPath))

	// r.Handle("/files/", http.StripPrefix("/files/", fs))
	// r.Handle("/js/", http.StripPrefix("/js/", jsfs))
	// r.Handle("/css/", http.StripPrefix("/css/", cssfs))
	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", fs))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsfs))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssfs))

	r.HandleFunc("/upload", uploadFileHandler())
	r.Handle("/key", keyTestPageHandler())
	r.Handle("/", indexPageHandler())

	// API
	r.HandleFunc("/api/notes", GetNoteHandler).Methods("GET")
	r.HandleFunc("/api/notes", BasicAuth(PostNoteHandler, user, passwd)).Methods("POST")
	r.HandleFunc("/api/notes/{id}", BasicAuth(PutNoteHandler, user, passwd)).Methods("PUT")
	r.HandleFunc("/api/notes/{id}", BasicAuth(DeleteNoteHandler, user, passwd)).Methods("DELETE")

	//session demo
	r.HandleFunc("/signin", Signin)
	r.HandleFunc("/uploadpage", UploadPage)
	r.HandleFunc("/refresh", Refresh)
	r.HandleFunc("/logout", logoutHandler).Methods("POST")

	//ChatRoom
	r.HandleFunc("/chatroom", chatroom)
	r.HandleFunc("/ws", handleConnections)

	//swagger files
	swaggerFs := http.FileServer(http.Dir("/swagger-ui"))
	r.PathPrefix("/swagger/").Handler(http.HandlerFunc(serveSwagger))
	r.PathPrefix("/swaggerui/").Handler(http.StripPrefix("/swaggerui/", swaggerFs))

	//handle grpc server
	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln(err)
	}
	server := grpc.NewServer()
	pb.RegisterNoteServiceServer(server, newNote())
	go server.Serve(listen)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	muxS := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterNoteServiceHandlerFromEndpoint(ctx, muxS, "localhost:50051", dialOpts)
	if err != nil {
		log.Fatalln(err)
	}
	r.PathPrefix("/gapi/").Handler(muxS)

	go handleMessages()

	log.Print("Server started on localhost:80/443, use /upload for uploading files and /files/{fileName} for downloading")
	go http.ListenAndServe(":80", http.HandlerFunc(redirect))
	http.ListenAndServeTLS(":443", "/server/cert.pem", "/server/key.pem", r)
}

func keyTestPageHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/server/key.html")
	})
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
		w.Write([]byte("<!DOCTYPE html><html><head> <meta http-equiv='refresh' content='5; URL=/uploadpage'></head><body>SUCCESS<p></p><a href=/uploadpage>Back to previous page</a></body></html>"))
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
