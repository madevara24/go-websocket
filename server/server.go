package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	gubrak "github.com/novalagung/gubrak/v2"
)

type M map[string]interface{}

const (
	MESSAGE_NEW_USER = "New User"
	MESSAGE_CHAT     = "Chat"
	MESSAGE_LEAVE    = "Leave"
	ERROR            = "ERROR"
	WEBSOCKET_CLOSE  = "websocket: close"
)

var conns = make([]*WebSocketConnection, 0)

type SocketPayload struct {
	Message string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func handleSoc(w http.ResponseWriter, r *http.Request) {
	gorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket conn", http.StatusBadRequest)
	}

	username := r.URL.Query().Get("username")
	currentConn := WebSocketConnection{Conn: gorillaConn, Username: username}

	conns = append(conns, &currentConn)

	go handleIO(&currentConn, conns)
}

func handleIO(currentConn *WebSocketConnection, conns []*WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(ERROR, fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), WEBSOCKET_CLOSE) {
				broadcastMessage(currentConn, MESSAGE_LEAVE, "")
				ejectConn(currentConn)
				return
			}

			log.Println(ERROR, err.Error())
			continue
		}

		broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
	}
}

func ejectConn(currentConn *WebSocketConnection) {
	filtered := gubrak.From(conns).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()

	conns = filtered.([]*WebSocketConnection)
}

func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
	switch kind {
	case MESSAGE_NEW_USER:
		fmt.Printf("User %v joined!\n", currentConn.Username)
	case MESSAGE_CHAT:
		fmt.Printf("%v: %v\n", currentConn.Username, message)
	case MESSAGE_LEAVE:
		fmt.Printf("User %v disconnected!\n", currentConn.Username)

	}
	fmt.Printf("%v: %v\n", currentConn.Username, message)
	for _, eachConn := range conns {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
		})
	}
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", handleSoc)

	fmt.Println("server listening...")
	http.ListenAndServe(":8080", nil)
}
