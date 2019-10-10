package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/shadow1163/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var counter int

var log = logger.NewLogger()

type JMes struct {
	Counter int
	Message string
}

func main() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
		counter++
		if err := conn.WriteJSON(JMes{Counter: counter, Message: ""}); err != nil {
			log.Error(err)
			return
		}
		for {
			// Read message from browser
			// msgType, msg, err := conn.ReadMessage()
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			log.Info(conn.RemoteAddr().String() + " sent: " + string(msg))

			// Write message back to browser
			if err = conn.WriteJSON(JMes{Counter: 0, Message: string(msg)}); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})

	log.Error(http.ListenAndServe(":8888", nil))

}
