package wstunnel

import (
	_ "golang.org/x/mobile/bind"
)

/*
This project uses gomobile to build android and ios libraries used in Windscribe apps for WStunnel support.
*/

// Channel Host app > Library
var channel = make(chan string)

// Callback for Library > Host app
var tunnelCallBack TunnelCallBack

func Initialise(development bool, logFilePath string) {
	InitLogger(development, logFilePath)
}

// StartWSTunnel Builds and start a http client (Tcp server + Handles Bi directional connection between clients and Websocket server)
// This Function blocks until exit signal is sent by host app.
func StartWSTunnel(listenAddress string, wsAddress string) bool {
	err := NewHTTPClient(listenAddress, wsAddress, func(fd int) {
		if tunnelCallBack != nil {
			tunnelCallBack.Protect(fd)
		} else {
			Logger.Info("Host app has not registered callback.")
		}
	}, channel).Run()
	if err != nil {
		return false
	}
	return true
}

// RegisterTunnelCallback is called from the host app.
func RegisterTunnelCallback(callback TunnelCallBack) {
	if callback != nil {
		Logger.Info("Connecting to host app.")
		tunnelCallBack = callback
	} else {
		Logger.Info("Disconnecting from host app.")
		channel <- "done"
	}
}

// TunnelCallBack Host app should implement this interface and register.
type TunnelCallBack interface {
	// Protect Web socket's underlying file descriptor sent to host app for protecting it from VPN Service.
	Protect(fd int)
}
