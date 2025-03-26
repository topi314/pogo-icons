package server

import (
	"errors"
	"log/slog"
	"net/http"
)

func New(cfg Config, version string, goVersion string) *Server {
	s := &Server{
		cfg:       cfg,
		version:   version,
		goVersion: goVersion,
	}

	s.server = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: s.Routes(),
	}

	return s
}

type Server struct {
	cfg       Config
	version   string
	goVersion string
	server    *http.Server
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", slog.Any("err", err))
		return
	}
}
