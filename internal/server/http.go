package server

import (
	"net"
	"net/http"

	"github.com/BhaumikTalwar/Gama/config"
)

func NewHttpServer(handler http.Handler, httpCfg *config.HTTPServerConfig) *http.Server {
	srv := &http.Server{
		Addr:              net.JoinHostPort(httpCfg.Host, httpCfg.Port),
		Handler:           handler,
		ReadTimeout:       httpCfg.ReadTimeout,
		ReadHeaderTimeout: httpCfg.ReadHeaderTimeout,
		WriteTimeout:      httpCfg.WriteTimeout,
		IdleTimeout:       httpCfg.IdleTimeout,

		MaxHeaderBytes: httpCfg.MaxHeaderBytes,
	}

	srv.SetKeepAlivesEnabled(httpCfg.KeepAlive)
	return srv
}
