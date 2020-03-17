package account

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shadow1163/golang-https-server/src/fileserver"
	"github.com/shadow1163/logger"
)

var (
	user   = []byte("user")
	passwd = []byte("password")
	db     DB
	log    = logger.Log
)

func init() {
	rdb := redisdb{}
	rdb.connect()
	if rdb.cache == nil {
		log.Warning("did not connect redis database, using memory DB")
		mdb := memorydb{}
		ex := make(map[string]bool)
		mdb.Expires = ex
		db = mdb
	} else {
		db = rdb
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
	uuid, err := uuid.NewUUID()
	if err != nil {
		// If there is an error in setting the account, return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<!DOCTYPE html><html><head> <meta http-equiv='refresh' content='5; URL=/'></head><body>Error 500<p></p><a href=/>Back to previous page</a></body></html>"))
		return
	}
	db.save(uuid.String(), true)

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 300 seconds
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   uuid.String(),
		Expires: time.Now().Add(300 * time.Second),
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
	response, err := db.get(sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !response {
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

	response, err := db.get(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !response {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Now, create a new session token for the current user
	uuid, err := uuid.NewUUID()
	err = db.save(uuid.String(), true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete the older session token
	err = db.del(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   uuid.String(),
		Expires: time.Now().Add(300 * time.Second),
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
