package test

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"wstunnel/wstunnel"
)

// EchoServer simple web socket server for testing.
type EchoServer struct {
	clients map[string]*websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (server *EchoServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	clientId := connection.RemoteAddr().String()
	server.clients[clientId] = connection
	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		go server.echoMessageBack(clientId, message)
	}
	connection.Close()
	delete(server.clients, clientId)
}

func startServer(address string, path string) *EchoServer {
	server := EchoServer{
		clients: map[string]*websocket.Conn{},
	}
	http.HandleFunc(path, server.handleRequest)
	go func() {
		err := http.ListenAndServe(address, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return &server
}

func (server *EchoServer) echoMessageBack(clientId string, message []byte) {
	err := server.clients[clientId].WriteMessage(websocket.BinaryMessage, message)
	if err != nil {
		wstunnel.Logger.Error("Error writing message to client: %s", err)
		return
	}
}
