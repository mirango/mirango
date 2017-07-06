package mirango

import (
	"net/http"

	"github.com/mirango/framework"
)

type Server struct {
	*http.Server
	Logger framework.LogWriter
}

func DefaultServer() *Server {
	return &Server{
		Server: &http.Server{},
	}
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	return srv.Server.ListenAndServe()
}

func (srv *Server) SetHandler(h http.Handler) {
	srv.Handler = h
}

func (srv *Server) SetAddr(addr string) {
	srv.Addr = addr
}

func (srv *Server) SetLogger(logger framework.LogWriter) {
	srv.Logger = logger
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	return srv.Server.ListenAndServeTLS(certFile, keyFile)
}

func (srv *Server) log(fmt string, v ...interface{}) {
	if srv.Logger != nil {
		srv.Logger.Print(fmt, v...)
	}
}
