package game

import "net/http"

func MiniGame(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/html/key.html")
}
