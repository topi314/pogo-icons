package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "getVersion")

	if _, err := fmt.Fprintf(w, "Dashboard %s (Go %s)", s.version, s.goVersion); err != nil {
		Error(r.Context(), w, "failed to write response", http.StatusInternalServerError)
	}
}

func (s *Server) page(w http.ResponseWriter, r *http.Request) {

}

func Error(ctx context.Context, w http.ResponseWriter, error string, code int) {
	slog.ErrorContext(ctx, error)
	http.Error(w, error, code)
}
