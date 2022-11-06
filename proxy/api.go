// Package proxy
//Library for Windscribe clients. Tunnel merely wraps and forwards traffic to remote server. Clients are responsible
//for encrypting data.
package proxy

import (
	_ "golang.org/x/mobile/bind"
)

// Channel is used by host app to send events to http client.
var channel = make(chan string)

// Callback is used by http client to send events to host app
var tunnelCallBack TunnelCallBack

// WSTunnel wraps OpenVPN tcp traffic in to Websocket
const WSTunnel = 1

// Stunnel wraps OpenVPN tcp traffic in to regular tcp.
const Stunnel = 2

func Initialise(development bool, logFilePath string) {
	InitLogger(development, logFilePath)
}

// StartProxy Builds and start a http client (Tcp server + Handles Bi directional connection between clients and remote server)
// This Function blocks until exit signal is sent by host app.
// listenAddress = ":LocalPort"
// remoteAddress = "wss://ip:port/path" or "ip:port"
// tunnelType = WSTunnel = 1 or Stunnel = 2
func StartProxy(listenAddress string, remoteAddress string, tunnelType int, mtu int) bool {
	Logger.Infof("Starting proxy with listenAddress: %s remoteAddress %s tunnelType: %d mtu %d", listenAddress, remoteAddress, tunnelType, mtu)
	err := NewHTTPClient(listenAddress, remoteAddress, tunnelType, mtu, func(fd int) {
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

// RegisterTunnelCallback is called from the host app to register for events from library.
func RegisterTunnelCallback(callback TunnelCallBack) {
	if callback != nil {
		Logger.Info("New connection from host app.")
		tunnelCallBack = callback
	} else {
		Logger.Info("Disconnect signal from host app.")
		channel <- "done"
	}
}

// TunnelCallBack Host app should implement this interface and register.
type TunnelCallBack interface {
	// Protect remote connection's underlying file descriptor sent to host app for protecting it from VPN Service.
	Protect(fd int)
}
