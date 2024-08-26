package cli

import (
	tls "github.com/refraction-networking/utls"
	"io"
	"net"
	"os"
)

// StunnelBiDirection
// creates an object to transfer data between the TCP clients and remote server in bidirectional way
type StunnelBiDirection struct {
	localConn  net.Conn
	remoteConn *tls.UConn
	mtu        int
}

func NewStunnelBiDirection(localConn net.Conn, remoteConn *tls.UConn, mtu int) Runner {
	return &StunnelBiDirection{
		localConn, remoteConn, mtu,
	}
}

func (s *StunnelBiDirection) Run() error {
	go s.sendTCPToStunnel()
	s.sendStunnelToTCP()
	return nil
}

// sendTCPToStunnel copies tcp traffic to remote server
func (s *StunnelBiDirection) sendTCPToStunnel() {
	defer s.close()
	data := make([]byte, s.mtu)
	for {
		readSize, err := s.localConn.Read(data)
		if err != nil && !os.IsTimeout(err) {
			if err != io.EOF {
				Logger.Errorf("TCPToStunnel - Error while reading from TCP: %s", err)
			}
			return
		}
		_, writeErr := s.remoteConn.Write(data[:readSize])
		if err != nil {
			Logger.Errorf("TCPToStunnel - Error while writing to Stunnel: %s", writeErr)
			return
		}
	}
}

// sendStunnelToTCP copies remote server traffic to tcp connection.
func (s *StunnelBiDirection) sendStunnelToTCP() {
	defer s.close()
	data := make([]byte, s.mtu)
	for {
		readSize, err := s.remoteConn.Read(data)
		if err != nil && !os.IsTimeout(err) {
			if err != io.EOF {
				Logger.Errorf("StunnelToTCP - Error while reading from Stunnel: %s", err)
			}
			break
		}
		_, writeErr := s.localConn.Write(data[:readSize])
		if err != nil {
			Logger.Errorf("StunnelToTCP - Error while writing to TCP: %s", writeErr)
			return
		}
	}
}

// close closes connections.
func (s *StunnelBiDirection) close() {
	_ = s.remoteConn.Close()
	_ = s.localConn.Close()
}
