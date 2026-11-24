package cli

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"io"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"

	utls "github.com/refraction-networking/utls"
)

type recordCaptureConn struct {
	mu     sync.Mutex
	writes [][]byte
}

func (c *recordCaptureConn) Read(_ []byte) (int, error)         { return 0, io.EOF }
func (c *recordCaptureConn) Close() error                       { return nil }
func (c *recordCaptureConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *recordCaptureConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *recordCaptureConn) SetDeadline(_ time.Time) error      { return nil }
func (c *recordCaptureConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *recordCaptureConn) SetWriteDeadline(_ time.Time) error { return nil }

func (c *recordCaptureConn) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	c.mu.Lock()
	c.writes = append(c.writes, cp)
	c.mu.Unlock()
	return len(p), nil
}

func TestTLSRecordSplitRewritesClientHello(t *testing.T) {
	raw := &recordCaptureConn{}
	conn := newTLSRecordSplitConn(raw)

	hs := make([]byte, 40)
	hs[0] = 0x01
	hs[1] = 0x00
	hs[2] = 0x00
	hs[3] = 0x24
	for i := 4; i < len(hs); i++ {
		hs[i] = byte(i)
	}
	orig := make([]byte, 5+len(hs))
	orig[0] = 0x16
	orig[1] = 0x03
	orig[2] = 0x01
	binary.BigEndian.PutUint16(orig[3:5], uint16(len(hs)))
	copy(orig[5:], hs)

	n, err := conn.Write(orig)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != len(orig) {
		t.Fatalf("expected reported len %d, got %d", len(orig), n)
	}
	if len(raw.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(raw.writes))
	}

	r1 := raw.writes[0]
	r2 := raw.writes[1]
	wantR1 := []byte{0x16, 0x03, 0x01, 0x00, 0x02, hs[0], hs[1]}
	if !bytes.Equal(r1, wantR1) {
		t.Fatalf("unexpected first TLS record: got %x want %x", r1, wantR1)
	}
	if len(r2) != 5+len(hs)-2 {
		t.Fatalf("unexpected second record len %d", len(r2))
	}
	if r2[0] != 0x16 || r2[1] != 0x03 || r2[2] != 0x01 {
		t.Fatalf("unexpected second record header: %x", r2[:5])
	}
	if int(binary.BigEndian.Uint16(r2[3:5])) != len(hs)-2 {
		t.Fatalf("unexpected second record payload len: %d", binary.BigEndian.Uint16(r2[3:5]))
	}
	reassembled := append(append([]byte{}, r1[5:]...), r2[5:]...)
	if !bytes.Equal(reassembled, hs) {
		t.Fatalf("reassembled handshake mismatch")
	}

	if _, err := conn.Write([]byte("app-data")); err != nil {
		t.Fatalf("unexpected follow-up write error: %v", err)
	}
	if len(raw.writes) != 3 || !bytes.Equal(raw.writes[2], []byte("app-data")) {
		t.Fatalf("subsequent writes should pass through unchanged")
	}
}

func makeTestCert(t *testing.T) tls.Certificate {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("key gen: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		DNSNames:     []string{"api.windscribe.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("cert gen: %v", err)
	}
	return tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  priv,
	}
}

func TestTLSRecordSplitEndToEndHandshake(t *testing.T) {
	cert := makeTestCert(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	helloIDs := []utls.ClientHelloID{utls.HelloRandomizedALPN, utls.HelloRandomizedNoALPN}
	for _, helloID := range helloIDs {
		srvErr := make(chan error, 1)
		go func() {
			raw, aErr := ln.Accept()
			if aErr != nil {
				srvErr <- aErr
				return
			}
			tConn := tls.Server(raw, &tls.Config{Certificates: []tls.Certificate{cert}})
			defer tConn.Close()
			if hErr := tConn.Handshake(); hErr != nil {
				srvErr <- hErr
				return
			}
			buf := make([]byte, 4)
			if _, rErr := io.ReadFull(tConn, buf); rErr != nil {
				srvErr <- rErr
				return
			}
			_, wErr := tConn.Write(buf)
			srvErr <- wErr
		}()

		rawClient, dErr := net.Dial("tcp", ln.Addr().String())
		if dErr != nil {
			t.Fatalf("dial: %v", dErr)
		}
		splitClient := newTLSRecordSplitConn(rawClient)
		uConn := utls.UClient(splitClient, &utls.Config{
			InsecureSkipVerify: true,
			ServerName:         "api.windscribe.com",
		}, helloID)
		if hErr := uConn.Handshake(); hErr != nil {
			_ = uConn.Close()
			t.Fatalf("uTLS handshake (%v): %v", helloID.Client, hErr)
		}
		if _, wErr := uConn.Write([]byte("ping")); wErr != nil {
			_ = uConn.Close()
			t.Fatalf("write: %v", wErr)
		}
		echo := make([]byte, 4)
		if _, rErr := io.ReadFull(uConn, echo); rErr != nil {
			_ = uConn.Close()
			t.Fatalf("read: %v", rErr)
		}
		_ = uConn.Close()
		if sErr := <-srvErr; sErr != nil {
			t.Fatalf("server error: %v", sErr)
		}
		if string(echo) != "ping" {
			t.Fatalf("expected ping, got %q", string(echo))
		}
	}
}
