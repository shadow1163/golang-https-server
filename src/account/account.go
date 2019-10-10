package account

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"

	"github.com/shadow1163/logger"
	"github.com/shadow1163/new-server/src/fileserver"
)

var (
	user   = []byte("user")
	passwd = []byte("password")
	cache  redis.Conn
	log    = logger.NewLogger()
)

func init() {
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	cache = conn
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
	obj := uuid.NewV4()
	sessionToken := obj.String()
	log.Debug(sessionToken)
	log.Debug(name)
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

// UploadPage upload page
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
	fileserver.FileserverUpload(w)
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
	obj := uuid.NewV4()
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
	// log.Println(newSessionToken)
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

func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}
