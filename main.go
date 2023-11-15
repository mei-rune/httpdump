package httpdump

import (
	"net/http"
)

func NewServer(rootDir string) (*Server, error) {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented", http.StatusInternalServerError)
	})

	if rootDir != "" {
		var err error
		handler, err = ReadDir(rootDir)
		if err != nil {
			return nil, err
		}
	}

	return &Server{RootDir: rootDir, Handler: handler}, nil
}

type Server struct {
	RootDir            string
	Handler            http.Handler
	username, password string
}

func (srv *Server) SetAuth(username, password string) *Server {
	srv.username = username
	srv.password = password
	return srv
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	srv.Handler.ServeHTTP(w, req)
}
