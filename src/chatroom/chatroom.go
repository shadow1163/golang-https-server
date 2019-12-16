package chatroom

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/shadow1163/logger"
)

//Message message struct
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var (
	log       = logger.NewLogger()
	upgrader  = websocket.Upgrader{}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
)

func init() {
	go handleMessages()
}

func ChatRoom(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/html/chatroom.html")
}

func HandleWSConnections(w http.ResponseWriter, r *http.Request) {
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
			log.Debug("error: ", err.Error())
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func GetChatRoomCounter(w http.ResponseWriter, r *http.Request) {
	counter := len(clients)

	w.Write([]byte(strconv.Itoa(counter)))
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
