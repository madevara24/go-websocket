package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	WEBSOCKET_URL = "ws://localhost:8080/ws?username="
)

type SocketPayload struct {
	Message string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
}

func main() {
	localUUID := uuid.NewString()
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%v%v", WEBSOCKET_URL, localUUID), nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to WebSocket. Type 'exit' to close the connection.")

	go func() {
		for {
			var msg SocketResponse
			err := conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println("Error reading from WebSocket:", err)
				return
			}

			switch msg.Type {
			case "New User":
				fmt.Printf("User %v joined!\n", msg.From)
			case "Chat":
				if msg.From != localUUID {
					fmt.Printf("%v: %v\n", msg.From, msg.Message)
				}
			case "Leave":
				fmt.Printf("User %v disconnected!\n", msg.From)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()

		if strings.ToLower(input) == "exit" {
			fmt.Println("Closing connection...")
			return
		}

		// fmt.Printf("You: %v\n", input)

		payload := SocketPayload{
			Message: input,
		}

		err := conn.WriteJSON(payload)
		if err != nil {
			fmt.Println("Error sending message to WebSocket:", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from terminal:", err)
	}
}
