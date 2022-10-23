package wstunnel

import (
	"Ws/utils"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// httpClient
////////////////////////////////////////////////////////////////////////////////

// httpClient implements the Runner interface
type httpClient struct {
	connectWS string
	listenTCP string
	callback  func(fd int)
	channel   chan string
}

func NewHTTPClient(listenTCP, connectWS string, callback func(fd int), channel chan string) Runner {
	return &httpClient{
		connectWS: connectWS,
		listenTCP: listenTCP,
		callback:  callback,
		channel:   channel,
	}
}

func (h *httpClient) Run() error {
	tcpAdr, err := net.ResolveTCPAddr("tcp", h.listenTCP)
	if err != nil {
		utils.Logger.Errorf("Error resolving tcp address: %s", err)
		return err
	}
	tcpConnection, err := net.ListenTCP("tcp", tcpAdr)
	if err != nil {
		return err
	}
	defer tcpConnection.Close()
	utils.Logger.Infof("Listening on 127.0.0.1:%s", h.listenTCP)
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
			utils.Logger.Error("Error: could not accept the connection: ", err)
			continue
		}
		utils.Logger.Infof("New connection from %s", tcpConn.RemoteAddr().String())

		wsConn, wsErr := h.createWsConnection(tcpConn.RemoteAddr().String())
		if wsErr != nil || wsConn == nil {
			utils.Logger.Errorf("%s - Ws connection > Error while dialing %s: %s", tcpConn.RemoteAddr(), h.connectWS, wsErr)
			_ = tcpConn.Close()
			continue
		}
		b := NewBidirConnection(tcpConn, wsConn, time.Second*10)
		go b.Run()
	}
	return err
}

func (h *httpClient) toWsURL(asString string) (string, error) {
	asURL, err := url.Parse(asString)
	if err != nil {
		return asString, err
	}

	switch asURL.Scheme {
	case "http":
		asURL.Scheme = "ws"
	case "https":
		asURL.Scheme = "wss"
	}
	return asURL.String(), nil
}

// Creates a connection to websocket server.
func (h *httpClient) createWsConnection(remoteAddr string) (wsConn *websocket.Conn, err error) {
	wsConnectUrl := h.connectWS
	for {
		var wsURL string
		wsURL, err = h.toWsURL(wsConnectUrl)
		if err != nil {
			return
		}
		utils.Logger.Infof("%s - Connecting to %s", remoteAddr, wsURL)
		var httpResponse *http.Response
		dialer := *websocket.DefaultDialer
		// Access underlying socket fd before connecting to it.
		customNetDialer := &net.Dialer{}
		customNetDialer.Control = func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				i := int(fd)
				if err != nil {
					return
				}
				h.callback(i)
			})
		}
		dialer.NetDial = func(network, addr string) (net.Conn, error) {
			return customNetDialer.Dial(network, addr)
		}
		//Connect
		wsConn, httpResponse, err = dialer.Dial(wsURL, nil)
		if httpResponse != nil {
			switch httpResponse.StatusCode {
			case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
				wsConnectUrl = httpResponse.Header.Get("Location")
				utils.Logger.Infof("%s - Redirect to %s", remoteAddr, wsConnectUrl)
				continue
			}
		}
		return
	}
}
