package client

import (
	"net"
	"net/http"
	"time"
)

// NewHTTPClient creates a custom http.Client with specified timeouts.
// This is crucial for preventing resource exhaustion and ensuring resilience
// in microservices communication.
func NewHTTPClient(
	totalTimeout,
	dialTimeout,
	tlsHandshakeTimeout,
	responseHeaderTimeout time.Duration,
) *http.Client {
	return &http.Client{
		Timeout: totalTimeout, // Overall request timeout
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: dialTimeout, // Connection establishment timeout
			}).DialContext,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,   // TLS handshake timeout
			ResponseHeaderTimeout: responseHeaderTimeout, // Time to wait for response headers
			MaxIdleConns:          100,                   // Max idle connections across all hosts
			IdleConnTimeout:       90 * time.Second,      // How long an idle connection is kept alive
			ForceAttemptHTTP2:     true,                  // Prefer HTTP/2
		},
	}
}
