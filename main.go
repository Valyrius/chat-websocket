package main

import (
	"html"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connected = make(map[*websocket.Conn]bool)
)

type Message struct {
	Username      string `json:"username"`
	UsernameColor string `json:"username_color"`
	Message       string `json:"message"`
	Timestamp     int64  `json:"timestamp"`
}

func main() {
	log.Println("starting server on :8765")
	http.HandleFunc("/", handleConnection)
	err := http.ListenAndServe(":8765", nil)
	if err != nil {
		log.Fatal("Error starting WebSocket server: ", err)
	}
}

func sendMessage(socket *websocket.Conn, message Message) {
	err := socket.WriteJSON(message)
	if err != nil {
		socket.Close()
		delete(connected, socket)
		log.Println(err)
	}
}

func broadcastMessage(message Message) {
	for client := range connected {
		go sendMessage(client, message)
	}
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	connected[conn] = true

	for {
		message := Message{}
		err := conn.ReadJSON(&message)
		if err != nil {
			delete(connected, conn)
			return
		}

		message.Username = html.EscapeString(message.Username)
		message.Message = html.EscapeString(message.Message)
		message.Timestamp = time.Now().Unix()

		broadcastMessage(message)
	}
}
