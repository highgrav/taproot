package taproot

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/highgrav/taproot/logging"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"strings"
)

func (srv *AppServer) startACME() error {
	srv.autocert = &autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       autocert.DirCache(srv.Config.HttpServer.TLS.ACMEDirectory),
		HostPolicy:  srv.acmeHostPolicy,
		RenewBefore: 0,
		Client:      nil,
	}
	srv.Server.Server.TLSConfig = &tls.Config{
		GetCertificate: srv.autocert.GetCertificate,
	}
	go http.ListenAndServe(":80", srv.autocert.HTTPHandler(nil))
	return nil
}

func (srv *AppServer) acmeHostPolicy(ctx context.Context, host string) error {
	for _, h := range srv.Config.HttpServer.TLS.ACMEAllowedHosts {
		if strings.ToLower(host) == strings.ToLower(h) {
			return nil
		}
	}
	logging.LogToDeck(ctx, "warning", "ACME", "alert", "invalid attempt to generate ACME cert for "+host)
	return errors.New("invalid attempt to generate ACME cert for " + host)
}
