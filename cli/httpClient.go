package cli

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	tls "github.com/refraction-networking/utls"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

//export WSTunnel wraps OpenVPN tcp traffic in to Websocket
const WSTunnel = 1

//export Stunnel wraps OpenVPN tcp traffic in to regular tcp.
const Stunnel = 2

// Channel holds one event, so Stop never blocks, also when the proxy isn't running.
//
//export Channel is used by host app to send events to http client.
var Channel = make(chan string, 1)

// httpClient
// sets up tcp server and remote connections.
// //////////////////////////////////////////////////////////////////////////////
type httpClient struct {
	listenTCP     string
	remoteServer  string
	tunnelType    int
	mtu           int
	callback      func(fd int)
	channel       chan string
	extraPadding  bool
	tlsServerName string
}

func NewHTTPClient(listenTCP, remoteServer string, tunnelType int, mtu int, callback func(fd int), channel chan string, extraPadding bool, tlsServerName string) Runner {
	return &httpClient{
		listenTCP:     listenTCP,
		remoteServer:  remoteServer,
		tunnelType:    tunnelType,
		mtu:           mtu,
		callback:      callback,
		channel:       channel,
		extraPadding:  extraPadding,
		tlsServerName: tlsServerName,
	}
}

// Run stars tcp server and connect to remote server.
func (h *httpClient) Run() error {
	if h.tunnelType != WSTunnel && h.tunnelType != Stunnel {
		err := fmt.Errorf("invalid tunnel type %d", h.tunnelType)
		Logger.Errorf("Not starting proxy: %s", err)
		return err
	}
	// A stop requested while no proxy was running belongs to an earlier run.
	select {
	case <-h.channel:
	default:
	}
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
	Logger.Infof("Listening on %s", h.listenTCP)
	doneMutex := sync.Mutex{}
	done := false
	isDone := func() bool {
		doneMutex.Lock()
		defer doneMutex.Unlock()
		return done
	}
	goSafely("stop listener", func() {
		select {
		case msg := <-h.channel:
			if msg == "done" {
				doneMutex.Lock()
				defer doneMutex.Unlock()
				done = true
				_ = tcpConnection.Close()
			}
		}
	})
	for !isDone() {
		tcpConn, err := tcpConnection.Accept()
		if err != nil {
			continue
		}
		Logger.Infof("New connection from %s", tcpConn.RemoteAddr().String())
		h.handleConnection(tcpConn)
	}
	return err
}

// handleConnection connects one local client to the remote server. A panic while connecting closes only this client.
func (h *httpClient) handleConnection(tcpConn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			_ = tcpConn.Close()
			Logger.Errorf("Recovered from a panic while connecting %s: %v\n%s", tcpConn.RemoteAddr(), r, debug.Stack())
		}
	}()
	if h.tunnelType == WSTunnel {
		handleWsTunnelConnection(h, tcpConn)
	} else {
		handleStunnelConnection(h, tcpConn)
	}
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
	b := NewStunnelBiDirection(localConn, remoteConn, h.mtu)
	goSafely("stunnel connection", func() { _ = b.Run() })
}

func (h *httpClient) createRemoteConnection() (*tls.UConn, error) {
	customNetDialer := h.createDialer()
	remoteUrl, err := url.Parse(h.remoteServer)
	if err != nil {
		return nil, err
	}
	serverName := remoteUrl.Hostname()
	if h.tlsServerName != "" {
		serverName = h.tlsServerName
	}
	cfg := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
	}
	netConn, err := customNetDialer.Dial("tcp", remoteUrl.Host)
	if err != nil {
		return nil, err
	}
	if h.extraPadding {
		netConn = newTLSRecordSplitConn(netConn)
	}

	remoteConn := tls.UClient(netConn, cfg, tls.HelloCustom)
	clientHelloSpec, err := tls.UTLSIdToSpec(tls.HelloRandomizedALPN)
	if err != nil {
		return nil, fmt.Errorf("uTlsConn.generateRandomizedSpec error: %+v", err)
	}

	err = remoteConn.ApplyPreset(&clientHelloSpec)
	if err != nil {
		return nil, fmt.Errorf("uTlsConn.ApplyPreset error: %+v", err)
	}

	return remoteConn, nil
}

func handleWsTunnelConnection(h *httpClient, tcpConn net.Conn) {
	wsConn, wsErr := h.createWsConnection(tcpConn.RemoteAddr().String())
	if wsErr != nil || wsConn == nil {
		Logger.Errorf("%s - Ws connection > Error while dialing %s: %s", tcpConn.RemoteAddr(), h.remoteServer, wsErr)
		_ = tcpConn.Close()
		return
	}
	b := NewBidirConnection(tcpConn, wsConn, time.Second*10, h.mtu)
	goSafely("wstunnel connection", func() { _ = b.Run() })
}

func (h *httpClient) toUrl(asString string) (string, error) {
	asURL, err := url.Parse(asString)
	if err != nil {
		return asString, err
	}
	return asURL.String(), nil
}

// createDialer creates custom dialer which provides access to socket fd
func (h *httpClient) createDialer() *net.Dialer {
	customNetDialer := &net.Dialer{}
	// Access underlying socket fd before connecting to it.
	customNetDialer.Control = func(network, address string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			Logger.Infof("Received socket fd %d", fd)
			i := int(fd)
			h.callback(i)
		})
	}
	return customNetDialer
}

// createWsConnection creates a connection to websocket server.
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
		tlsServerName := h.tlsServerName
		if tlsServerName == "" {
			if u, e := url.Parse(wsConnectUrl); e == nil {
				tlsServerName = u.Hostname()
			}
		}
		tlsCfg := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         tlsServerName,
		}
		dialer.TLSClientConfig = tlsCfg
		customNetDialer := h.createDialer()
		dialer.NetDial = func(network, addr string) (net.Conn, error) {
			conn, dialErr := customNetDialer.Dial(network, addr)
			if dialErr == nil && h.extraPadding {
				return newTLSRecordSplitConn(conn), nil
			}
			return conn, dialErr
		}
		wsConn, httpResponse, err = dialer.Dial(wsURL, nil)
		if wsConn != nil {
			Logger.Info("Successfully connected to remote server.")
		} else if err != nil {
			Logger.Errorf("Failed to connect to remote server.. %s", err)
		}
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
