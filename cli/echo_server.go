package cli

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

// EchoServer simple web socket server for testing.
type EchoServer struct {
	clients map[string]*websocket.Conn
	lock    sync.Mutex
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
	server.lock.Lock()
	server.clients[clientId] = connection
	server.lock.Unlock()
	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		go server.echoMessageBack(clientId, message)
	}
	connection.Close()
	server.lock.Lock()
	delete(server.clients, clientId)
	server.lock.Unlock()
}

func startServer(address string, path string) *EchoServer {
	server := EchoServer{
		clients: map[string]*websocket.Conn{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc(path, server.handleRequest)
	go func() {
		err := http.ListenAndServe(address, mux)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return &server
}

func (server *EchoServer) echoMessageBack(clientId string, message []byte) {
	server.lock.Lock()
	defer server.lock.Unlock()
	connection, ok := server.clients[clientId]
	if !ok {
		return
	}
	_ = connection.WriteMessage(websocket.BinaryMessage, message)
}
