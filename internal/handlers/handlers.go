package handlers

import (
	"context"
	"net"
	"net/http"
	"time"

	apperr "gitlab.com/bitboardexchange.com/georedirect/errors"
	"gitlab.com/bitboardexchange.com/georedirect/internal/handlers/messages"
	"gitlab.com/bitboardexchange.com/georedirect/internal/services"
	"gitlab.com/bitboardexchange.com/georedirect/internal/services/structures"
)

type fingerprintKey struct{}

func WithFingerprint(ctx context.Context, fp string) context.Context {
	return context.WithValue(ctx, fingerprintKey{}, fp)
}

type Handler struct {
	svc          *services.Service
	redisTimeout time.Duration
}

func New(svc *services.Service, redisTimeout time.Duration) *Handler {
	return &Handler{svc: svc, redisTimeout: redisTimeout}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("GET /r", h.redirect)
	mux.HandleFunc("GET /r/", h.redirect)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.redisTimeout)
	defer cancel()
	if err := h.svc.Ready(ctx); err != nil {
		apperr.Write(w, apperr.Unavailable())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	fp, _ := r.Context().Value(fingerprintKey{}).(string)
	req := messages.RedirectRequest{
		ClientIP:    host,
		UserAgent:   r.UserAgent(),
		Fingerprint: fp,
		Path:        r.URL.Path,
		Query:       r.URL.RawQuery,
	}
	decision, err := h.svc.Redirect(r.Context(), structures.Click{
		ClientIP:    req.ClientIP,
		UserAgent:   req.UserAgent,
		Fingerprint: req.Fingerprint,
		Path:        req.Path,
		Query:       req.Query,
	})
	if err != nil {
		apperr.Write(w, apperr.Internal())
		return
	}
	resp := messages.RedirectResponse{Location: decision.Location, Blocked: decision.Blocked}
	w.Header().Set("Connection", "close")
	if resp.Blocked {
		apperr.Write(w, apperr.NotFound())
		return
	}
	http.Redirect(w, r, resp.Location, http.StatusFound)
}
