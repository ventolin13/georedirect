package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr         string
	RedisAddr          string
	ReferenceRedisAddr string
	KafkaAddr          string
	KafkaTopic         string
	OfferURL           string
	ProxyProtocol      bool
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ShutdownTimeout    time.Duration
	RedisTimeout       time.Duration
}

type fileConfig struct {
	ListenAddr         string `yaml:"listen_addr"`
	RedisAddr          string `yaml:"redis_addr"`
	ReferenceRedisAddr string `yaml:"reference_redis_addr"`
	KafkaAddr          string `yaml:"kafka_addr"`
	KafkaTopic         string `yaml:"kafka_topic"`
	OfferURL           string `yaml:"offer_url"`
	ProxyProtocol      bool   `yaml:"proxy_protocol"`
	ReadTimeout        string `yaml:"read_timeout"`
	WriteTimeout       string `yaml:"write_timeout"`
	ShutdownTimeout    string `yaml:"shutdown_timeout"`
	RedisTimeout       string `yaml:"redis_timeout"`
}

func Load(path string) (Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var raw fileConfig
	if err := yaml.Unmarshal(buf, &raw); err != nil {
		return Config{}, err
	}
	cfg := Config{
		ListenAddr:         raw.ListenAddr,
		RedisAddr:          raw.RedisAddr,
		ReferenceRedisAddr: raw.ReferenceRedisAddr,
		KafkaAddr:          raw.KafkaAddr,
		KafkaTopic:         raw.KafkaTopic,
		OfferURL:           raw.OfferURL,
		ProxyProtocol:      raw.ProxyProtocol,
	}
	if cfg.ListenAddr == "" || cfg.RedisAddr == "" || cfg.ReferenceRedisAddr == "" || cfg.KafkaAddr == "" || cfg.KafkaTopic == "" || cfg.OfferURL == "" {
		return Config{}, errors.New("listen_addr, redis_addr, reference_redis_addr, kafka_addr, kafka_topic and offer_url are required")
	}
	if cfg.ReadTimeout, err = durationOr(raw.ReadTimeout, 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = durationOr(raw.WriteTimeout, 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = durationOr(raw.ShutdownTimeout, 15*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.RedisTimeout, err = durationOr(raw.RedisTimeout, 20*time.Millisecond); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func durationOr(raw string, fallback time.Duration) (time.Duration, error) {
	if raw == "" {
		return fallback, nil
	}
	return time.ParseDuration(raw)
}
