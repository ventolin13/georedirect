package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"gitlab.com/bitboardexchange.com/georedirect/internal/config"
	"gitlab.com/bitboardexchange.com/georedirect/internal/handlers"
	"gitlab.com/bitboardexchange.com/georedirect/internal/repositories"
	"gitlab.com/bitboardexchange.com/georedirect/internal/services"
)

func main() {
	path := os.Getenv("CONFIG")
	if path == "" {
		path = "deployments/local.yaml"
	}
	cfg, err := config.Load(path)
	if err != nil {
		log.Fatal(err)
	}

	repo := repositories.New()
	repo.Use(openClients(cfg))
	defer repo.Close()

	svc := services.New(repo)
	svc.SetOffer(cfg.OfferURL)

	srv := &http.Server{
		Handler:           withFingerprint(handlers.New(svc, cfg.RedisTimeout).Routes()),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}

	ln, err := listen(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go watchConfig(ctx, path, cfg, svc, repo)

	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(stopCtx); err != nil {
			log.Fatal(err)
		}
		repo.Close()
		if err := <-errCh; err != nil {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	}
}

func openClients(cfg config.Config) (*redis.Client, *redis.Client, *kafka.Writer) {
	return redisClient(cfg.RedisAddr, cfg.RedisTimeout),
		redisClient(cfg.ReferenceRedisAddr, cfg.RedisTimeout),
		&kafka.Writer{
			Addr:         kafka.TCP(cfg.KafkaAddr),
			Topic:        cfg.KafkaTopic,
			Async:        true,
			RequiredAcks: kafka.RequireOne,
		}
}

func redisClient(addr string, timeout time.Duration) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		PoolSize:     32,
	})
}

func watchConfig(ctx context.Context, path string, cur config.Config, svc *services.Service, repo *repositories.Repository) {
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	defer signal.Stop(hup)
	for {
		select {
		case <-ctx.Done():
			return
		case <-hup:
			next, err := config.Load(path)
			if err != nil {
				log.Printf("config: %v", err)
				continue
			}
			svc.SetOffer(next.OfferURL)
			if next.RedisAddr != cur.RedisAddr || next.ReferenceRedisAddr != cur.ReferenceRedisAddr || next.RedisTimeout != cur.RedisTimeout || next.KafkaAddr != cur.KafkaAddr || next.KafkaTopic != cur.KafkaTopic {
				repo.Use(openClients(next))
			}
			cur = next
		}
	}
}

func withFingerprint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := handlers.WithFingerprint(r.Context(), takeFingerprint(r.RemoteAddr))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
