package cli

import (
	"net"
	"sync"
)

type tlsRecordSplitConn struct {
	net.Conn
	mu        sync.Mutex
	splitDone bool
}

func newTLSRecordSplitConn(c net.Conn) net.Conn {
	return &tlsRecordSplitConn{Conn: c}
}

func writeFull(w net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		if n <= 0 {
			return net.ErrClosed
		}
		data = data[n:]
	}
	return nil
}

func (c *tlsRecordSplitConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	shouldSplit := !c.splitDone && len(p) > 0
	if shouldSplit {
		c.splitDone = true
	}
	c.mu.Unlock()

	if shouldSplit && len(p) >= 8 && p[0] == 0x16 && p[1] == 0x03 {
		recLen := int(p[3])<<8 | int(p[4])
		if recLen > 2 && 5+recLen <= len(p) && p[5] == 0x01 {
			rec1 := []byte{0x16, p[1], p[2], 0x00, 0x02, p[5], p[6]}
			remLen := recLen - 2
			rec2 := make([]byte, 5+len(p)-7)
			rec2[0] = 0x16
			rec2[1] = p[1]
			rec2[2] = p[2]
			rec2[3] = byte(remLen >> 8)
			rec2[4] = byte(remLen)
			copy(rec2[5:], p[7:])
			if err := writeFull(c.Conn, rec1); err != nil {
				return 0, err
			}
			if err := writeFull(c.Conn, rec2); err != nil {
				return 0, err
			}
			return len(p), nil
		}
	}
	return c.Conn.Write(p)
}
