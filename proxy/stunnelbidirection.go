package proxy

import (
	tls "github.com/refraction-networking/utls"
	"io"
	"net"
)

//StunnelBiDirection
//creates an object to transfer data between the TCP clients and remote server in bidirectional way
type StunnelBiDirection struct {
	localConn  net.Conn
	remoteConn *tls.UConn
}

func NewStunnelBiDirection(localConn net.Conn, remoteConn *tls.UConn) Runner {
	return &StunnelBiDirection{
		localConn, remoteConn,
	}
}

func (s *StunnelBiDirection) Run() error {
	go s.sendTCPToStunnel()
	s.sendStunnelToTCP()
	return nil
}

//sendTCPToStunnel copies tcp traffic to remote server
func (s *StunnelBiDirection) sendTCPToStunnel() {
	defer s.close()
	for {
		_, err := io.Copy(s.remoteConn, s.localConn)
		if err != nil {
			break
		}
	}
}

//sendStunnelToTCP copies remote server traffic to tcp connection.
func (s *StunnelBiDirection) sendStunnelToTCP() {
	defer s.close()
	for {
		_, err := io.Copy(s.localConn, s.remoteConn)
		if err != nil {
			break
		}
	}
}

// close closes connections.
func (s *StunnelBiDirection) close() {
	_ = s.remoteConn.Close()
	_ = s.localConn.Close()
}
