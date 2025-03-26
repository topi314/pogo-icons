package server

import (
	"net/http"
)

func (s *Server) Routes() http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /version", s.getVersion)
	r.HandleFunc("/", s.getControl)

	return r
}
