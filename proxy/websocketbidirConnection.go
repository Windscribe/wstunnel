package proxy

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

//WebSocketBiDirection
//Creates an object to transfer data between the TCP clients and remote server in bidirectional way
type WebSocketBiDirection struct {
	tcpConn        net.Conn
	wsConn         *websocket.Conn
	tcpReadTimeout time.Duration
}

func NewBidirConnection(tcpConn net.Conn, wsConn *websocket.Conn, tcpReadTimeout time.Duration) Runner {
	return &WebSocketBiDirection{
		tcpConn:        tcpConn,
		wsConn:         wsConn,
		tcpReadTimeout: tcpReadTimeout,
	}
}

//sendTCPToWS copies tcp traffic to web socket connection.
func (b *WebSocketBiDirection) sendTCPToWS() {
	defer b.close()
	data := make([]byte, BufferSize)
	for {
		if b.tcpReadTimeout > 0 {
			_ = b.tcpConn.SetReadDeadline(time.Now().Add(b.tcpReadTimeout))
		}
		readSize, err := b.tcpConn.Read(data)
		if err != nil && !os.IsTimeout(err) {
			if err != io.EOF {
				Logger.Errorf("TCPToWS - Error while reading from TCP: %s", err)
			}
			return
		}

		if err := b.wsConn.WriteMessage(websocket.BinaryMessage, data[:readSize]); err != nil {
			Logger.Errorf("TCPToWS - Error while writing to WS: %s", err)
			return
		}
	}
}

//sendWSToTCP copies web socket traffic to tcp connection.
func (b *WebSocketBiDirection) sendWSToTCP() {
	defer b.close()
	data := make([]byte, BufferSize)
	for {
		messageType, wsReader, err := b.wsConn.NextReader()
		if err != nil {
			Logger.Errorf("WSToTCP - Error while reading from WS: %s", err)
			return
		}
		if messageType != websocket.BinaryMessage {
			Logger.Infof("WSToTCP - Got wrong message type from WS: %s", messageType)
			return
		}

		for {
			readSize, err := wsReader.Read(data)
			if err != nil {
				if err != io.EOF {
					Logger.Errorf("WSToTCP - Error while reading from WS: %s", err)
				}
				break
			}

			if _, err := b.tcpConn.Write(data[:readSize]); err != nil {
				Logger.Errorf("WSToTCP - Error while writing to TCP: %s", err)
				return
			}
		}
	}
}

func (b *WebSocketBiDirection) Run() error {
	go b.sendTCPToWS()
	b.sendWSToTCP()
	return nil
}

// close closes connections.
func (b *WebSocketBiDirection) close() {
	_ = b.wsConn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
	_ = b.wsConn.Close()
	_ = b.tcpConn.Close()
}
