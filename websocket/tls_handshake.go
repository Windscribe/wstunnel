//go:build go1.17
// +build go1.17

package websocket

import (
	"context"
)
import tls "github.com/refraction-networking/utls"

func doHandshake(ctx context.Context, tlsConn *tls.UConn, cfg *tls.Config) error {
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return err
	}
	if !cfg.InsecureSkipVerify {
		if err := tlsConn.VerifyHostname(cfg.ServerName); err != nil {
			return err
		}
	}
	return nil
}
