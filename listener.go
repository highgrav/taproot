package taproot

import (
	"crypto/tls"
	"net"
)

/*
A custom net.Listener override that provides some additional logging and tracing capabilities.
Used in the TLS ListenAndServe methods so we can do a better job capturing and tracking TLS
errors than the usual stdout. (Given the amount of port scanning on the net, this gets ... verbose)
*/

type taplistener struct {
	net.Listener
	onTlsHandshakeFailure func(error)
}

func (tl *taplistener) Accept() (net.Conn, error) {
	c, err := tl.Listener.Accept()
	if err != nil {
		return c, err
	}

	// Capture
	if err := c.(*tls.Conn).Handshake(); err != nil {
		tl.onTlsHandshakeFailure(err)
	}

	return c, nil
}
