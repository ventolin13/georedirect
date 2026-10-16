package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"

	"gitlab.com/bitboardexchange.com/georedirect/internal/repositories"
	"gitlab.com/bitboardexchange.com/georedirect/internal/services/structures"
	"gitlab.com/bitboardexchange.com/georedirect/models"
)

type Service struct {
	repo  *repositories.Repository
	offer atomic.Value
}

func New(repo *repositories.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetOffer(url string) {
	s.offer.Store(url)
}

func (s *Service) Ready(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) Redirect(ctx context.Context, in structures.Click) (structures.Decision, error) {
	blocked, err := s.repo.Blocked(ctx, models.ClientIP{IP: in.ClientIP})
	if err != nil {
		blocked = false
	}
	if fingerprintMismatch(in.Fingerprint, in.UserAgent) {
		blocked = true
	}
	limited, err := s.repo.Limited(ctx, models.ClientIP{IP: in.ClientIP})
	if err != nil {
		limited = false
	}
	refuse := blocked || limited
	var id [8]byte
	_, _ = rand.Read(id[:])
	_ = s.repo.Save(ctx, models.Click{
		ClickID:     hex.EncodeToString(id[:]),
		ClientIP:    in.ClientIP,
		UserAgent:   in.UserAgent,
		Fingerprint: in.Fingerprint,
		Path:        in.Path,
		Query:       in.Query,
		Blocked:     refuse,
		At:          time.Now().UTC(),
	})
	url, _ := s.offer.Load().(string)
	return structures.Decision{Location: url, Blocked: refuse}, nil
}

func fingerprintMismatch(fingerprint, userAgent string) bool {
	// TODO: сверка отпечатка с User-Agent. Несовпадение — отказ.
	_, _ = fingerprint, userAgent
	return false
}
