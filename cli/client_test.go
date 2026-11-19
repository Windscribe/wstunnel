package cli

import (
	"fmt"
	"net"
	"testing"
	"time"
)

var protocol = "ws://"
var echoServerAddress = "localhost:8080"
var path = "/ws"
var webSocketServerAddress = fmt.Sprintf("%s%s%s", protocol, echoServerAddress, path)
var tcpServerAddress = "localhost:1194"
var dataToSend = []byte("Send me this message back.")

func TestEndToEndConnection(t *testing.T) {
	// Free ports, so the test doesn't depend on what else runs on the machine (8080 is taken on CI runners).
	echoServerAddress = freeAddress(t)
	webSocketServerAddress = fmt.Sprintf("%s%s%s", protocol, echoServerAddress, path)
	tcpServerAddress = freeAddress(t)
	//Ws server
	startServer(echoServerAddress, path)
	//Tcp server
	go func() {
		err := NewHTTPClient(tcpServerAddress, webSocketServerAddress, 1, 1600, func(fd int) {
			t.Log(fd)
		}, Channel, false, "").Run()
		if err != nil {
			t.Fail()
			return
		}
	}()
	time.Sleep(time.Millisecond * 100)
	//Client 1
	_, client1Err := mockClientConnection()
	if client1Err != nil {
		t.Fail()
		return
	}
	//Client 2
	_, client2Err := mockClientConnection()
	if client2Err != nil {
		t.Fail()
		return
	}
	//Exit
	time.Sleep(time.Millisecond * 100)
	Channel <- "done"
	//Client 3
	_, client3Err := mockClientConnection()
	if client3Err == nil {
		t.Fail()
		return
	}
	t.Log("Test is successful.")
}

func freeAddress(t *testing.T) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().String()
}

func mockClientConnection() (string, error) {
	var conn, connErr = net.Dial("tcp", tcpServerAddress)
	if connErr != nil {
		return "", connErr
	}
	_, writeErr := conn.Write(dataToSend)
	if writeErr != nil {
		return "", writeErr
	}
	time.Sleep(time.Second * 1)
	data := make([]byte, 30)
	_, err := conn.Read(data)
	return string(data), err
}
