package proxy

import (
	"context"
	"github.com/gorilla/websocket"
	tls "github.com/refraction-networking/utls"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"
	"time"
)

//httpClient
//sets up tcp server and remote connections.
////////////////////////////////////////////////////////////////////////////////
type httpClient struct {
	listenTCP    string
	remoteServer string
	tunnelType   int
	callback     func(fd int)
	channel      chan string
}

func NewHTTPClient(listenTCP, remoteServer string, tunnelType int, callback func(fd int), channel chan string) Runner {
	return &httpClient{
		listenTCP:    listenTCP,
		remoteServer: remoteServer,
		tunnelType:   tunnelType,
		callback:     callback,
		channel:      channel,
	}
}

//Run stars tcp server and connect to remote server.
func (h *httpClient) Run() error {
	tcpAdr, err := net.ResolveTCPAddr("tcp", h.listenTCP)
	if err != nil {
		Logger.Errorf("Error resolving tcp address: %s", err)
		return err
	}
	tcpConnection, err := net.ListenTCP("tcp", tcpAdr)
	if err != nil {
		return err
	}
	defer tcpConnection.Close()
	Logger.Infof("Listening on 127.0.0.1:%s", h.listenTCP)
	doneMutex := sync.Mutex{}
	done := false
	isDone := func() bool {
		doneMutex.Lock()
		defer doneMutex.Unlock()
		return done
	}
	go func() {
		select {
		case msg := <-h.channel:
			if msg == "done" {
				doneMutex.Lock()
				defer doneMutex.Unlock()
				done = true
				_ = tcpConnection.Close()
			}
		}
	}()
	for !isDone() {
		tcpConn, err := tcpConnection.Accept()
		if err != nil {
			Logger.Error("Error: could not accept the connection: ", err)
			continue
		}
		Logger.Infof("New connection from %s", tcpConn.RemoteAddr().String())
		if h.tunnelType == WSTunnel {
			handleWsTunnelConnection(h, tcpConn)
		} else if h.tunnelType == Stunnel {
			handleStunnelConnection(h, tcpConn)
		} else {
			Logger.Fatal("Invalid tunnel type specified.")
		}
	}
	return err
}

func handleStunnelConnection(h *httpClient, localConn net.Conn) {
	remoteConn, err := h.createRemoteConnection()
	if err != nil {
		Logger.Errorf("%s - Remote server connection > Error while dialing %s: %s", localConn.RemoteAddr(), h.remoteServer, err)
		_ = localConn.Close()
		return
	}
	err = remoteConn.HandshakeContext(context.Background())
	if err != nil {
		_ = localConn.Close()
		Logger.Errorf("Error on handshake: %s", err)
		return
	}
	Logger.Info("Starting stunnel bi-direction connection.")
	b := NewStunnelBiDirection(localConn, remoteConn)
	go b.Run()
}

func (h *httpClient) createRemoteConnection() (*tls.UConn, error) {
	customNetDialer := h.createDialer()
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	remoteUrl, err := h.toUrl(h.remoteServer)
	if err != nil {
		return nil, err
	}
	netConn, err := customNetDialer.Dial("tcp", remoteUrl)
	if err != nil {
		return nil, err
	}
	cfg.ServerName = h.remoteServer
	remoteConn := tls.UClient(netConn, cfg, tls.HelloRandomized)
	return remoteConn, nil
}

func handleWsTunnelConnection(h *httpClient, tcpConn net.Conn) {
	wsConn, wsErr := h.createWsConnection(tcpConn.RemoteAddr().String())
	if wsErr != nil || wsConn == nil {
		Logger.Errorf("%s - Ws connection > Error while dialing %s: %s", tcpConn.RemoteAddr(), h.remoteServer, wsErr)
		_ = tcpConn.Close()
		return
	}
	b := NewBidirConnection(tcpConn, wsConn, time.Second*10)
	go b.Run()
}

func (h *httpClient) toUrl(asString string) (string, error) {
	asURL, err := url.Parse(asString)
	if err != nil {
		return asString, err
	}
	return asURL.String(), nil
}

//createDialer creates custom dialer which provides access to socket fd
func (h *httpClient) createDialer() *net.Dialer {
	customNetDialer := &net.Dialer{}
	// Access underlying socket fd before connecting to it.
	customNetDialer.Control = func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			i := int(fd)
			h.callback(i)
		})
	}
	return customNetDialer
}

//createWsConnection creates a connection to websocket server.
func (h *httpClient) createWsConnection(remoteAddr string) (wsConn *websocket.Conn, err error) {
	wsConnectUrl := h.remoteServer
	for {
		var wsURL string
		wsURL, err = h.toUrl(wsConnectUrl)
		if err != nil {
			return
		}
		Logger.Infof("%s - Connecting to %s", remoteAddr, wsURL)
		var httpResponse *http.Response
		dialer := *websocket.DefaultDialer
		customNetDialer := h.createDialer()
		dialer.NetDial = func(network, addr string) (net.Conn, error) {
			return customNetDialer.Dial(network, addr)
		}
		wsConn, httpResponse, err = dialer.Dial(wsURL, nil)
		if httpResponse != nil {
			switch httpResponse.StatusCode {
			case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
				wsConnectUrl = httpResponse.Header.Get("Location")
				Logger.Infof("%s - Redirect to %s", remoteAddr, wsConnectUrl)
				continue
			}
		}
		return
	}
}
