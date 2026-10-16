package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pires/go-proxyproto"

	"gitlab.com/bitboardexchange.com/georedirect/internal/config"
)

func listen(cfg config.Config) (net.Listener, error) {
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return nil, err
	}
	if cfg.ProxyProtocol {
		ln = &proxyproto.Listener{Listener: ln}
	}

	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")
	if certFile == "" && keyFile == "" {
		return ln, nil
	}
	if certFile == "" || keyFile == "" {
		_ = ln.Close()
		return nil, errors.New("TLS_CERT and TLS_KEY must be set together")
	}

	store, err := newCertStore(certFile, keyFile)
	if err != nil {
		_ = ln.Close()
		return nil, err
	}
	return tls.NewListener(ln, &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: store.get,
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			if hello != nil && hello.Conn != nil {
				fingerprints.Store(hello.Conn.RemoteAddr().String(), fingerprint(hello))
			}
			return nil, nil
		},
	}), nil
}

type certStore struct {
	certFile string
	keyFile  string
	mu       sync.RWMutex
	cert     *tls.Certificate
	certTime time.Time
	keyTime  time.Time
}

func newCertStore(certFile, keyFile string) (*certStore, error) {
	s := &certStore{certFile: certFile, keyFile: keyFile}
	if _, err := s.get(nil); err != nil {
		return nil, err
	}
	return s, nil
}

// На рукопожатии смотрит время TLS_CERT и TLS_KEY. Новая пара не читается — остаётся прежняя.
func (s *certStore) get(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certInfo, certErr := os.Stat(s.certFile)
	keyInfo, keyErr := os.Stat(s.keyFile)

	s.mu.RLock()
	current := s.cert
	same := current != nil && certErr == nil && keyErr == nil &&
		s.certTime.Equal(certInfo.ModTime()) && s.keyTime.Equal(keyInfo.ModTime())
	s.mu.RUnlock()
	if same {
		return current, nil
	}
	if certErr != nil || keyErr != nil {
		if current != nil {
			return current, nil
		}
		if certErr != nil {
			return nil, certErr
		}
		return nil, keyErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cert != nil && s.certTime.Equal(certInfo.ModTime()) && s.keyTime.Equal(keyInfo.ModTime()) {
		return s.cert, nil
	}
	pair, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		if s.cert != nil {
			return s.cert, nil
		}
		return nil, err
	}
	s.cert = &pair
	s.certTime = certInfo.ModTime()
	s.keyTime = keyInfo.ModTime()
	return s.cert, nil
}

var fingerprints sync.Map

func takeFingerprint(addr string) string {
	v, ok := fingerprints.LoadAndDelete(addr)
	if !ok {
		return ""
	}
	fp, _ := v.(string)
	return fp
}

func fingerprint(hello *tls.ClientHelloInfo) string {
	ver := "00"
	for _, v := range hello.SupportedVersions {
		if v == tls.VersionTLS13 {
			ver = "13"
		} else if ver != "13" && v == tls.VersionTLS12 {
			ver = "12"
		}
	}
	sni := "i"
	if hello.ServerName != "" {
		sni = "d"
	}
	ciphers := make([]string, 0, len(hello.CipherSuites))
	for _, c := range hello.CipherSuites {
		if c&0x0f0f == 0x0a0a {
			continue
		}
		ciphers = append(ciphers, strconv.FormatUint(uint64(c), 10))
	}
	sort.Strings(ciphers)
	sum := sha256.Sum256([]byte(strings.Join(ciphers, ",")))
	alpn := "00"
	if len(hello.SupportedProtos) > 0 {
		alpn = hello.SupportedProtos[0]
		if len(alpn) > 2 {
			alpn = alpn[:2]
		}
	}
	return "t" + ver + sni + strconv.Itoa(len(ciphers)) + "_" + hex.EncodeToString(sum[:6]) + "_" + alpn
}
