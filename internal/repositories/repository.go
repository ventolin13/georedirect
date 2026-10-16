package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"gitlab.com/bitboardexchange.com/georedirect/models"
)

type Repository struct {
	mu        sync.RWMutex
	rdb       *redis.Client
	reference *redis.Client
	writer    *kafka.Writer
	block     map[string]struct{}
}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) Use(rdb, reference *redis.Client, writer *kafka.Writer) {
	r.mu.Lock()
	oldRdb, oldReference, oldWriter := r.rdb, r.reference, r.writer
	r.rdb, r.reference, r.writer = rdb, reference, writer
	r.mu.Unlock()
	closeClients(oldRdb, oldReference, oldWriter)
}

func (r *Repository) Close() {
	r.Use(nil, nil, nil)
}

func closeClients(rdb, reference *redis.Client, writer *kafka.Writer) {
	if writer != nil {
		_ = writer.Close()
	}
	if rdb != nil {
		_ = rdb.Close()
	}
	if reference != nil {
		_ = reference.Close()
	}
}

func (r *Repository) Ping(ctx context.Context) error {
	r.mu.RLock()
	rdb, reference := r.rdb, r.reference
	r.mu.RUnlock()
	if rdb == nil || reference == nil {
		return errors.New("redis")
	}
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	return reference.Ping(ctx).Err()
}

func (r *Repository) Blocked(_ context.Context, ip models.ClientIP) (bool, error) {
	// TODO: чёрный список в памяти пода, обновление из redis_addr.
	_, ok := r.block[ip.IP]
	return ok, nil
}

func (r *Repository) Limited(_ context.Context, ip models.ClientIP) (bool, error) {
	// TODO: рейт-лимит в том же redis_addr. Попадание — отказ.
	_ = ip
	return false, nil
}

func (r *Repository) Save(_ context.Context, click models.Click) error {
	body, err := json.Marshal(click)
	if err != nil {
		return err
	}
	r.mu.RLock()
	writer := r.writer
	r.mu.RUnlock()
	if writer == nil {
		return nil
	}
	return writer.WriteMessages(context.Background(), kafka.Message{Value: body})
}
