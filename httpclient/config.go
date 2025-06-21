package httpclient

import "time"

// Config represents configuration settings for network connections.
// It includes timeouts and dialer-specific parameters that can be
// unmarshalled from either JSON or YAML formats.
type Config struct {
	// Timeout specifies the maximum duration for the entire connection
	// process. Zero means no timeout.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// TLSHandshakeTimeout specifies the maximum duration to wait for
	// a TLS handshake to complete. Zero means no timeout.
	TLSHandshakeTimeout time.Duration `json:"tls_handshake_timeout" yaml:"tls_handshake_timeout"`

	// Dialer contains configuration specific to the connection dialer.
	Dialer struct {
		// Timeout is the maximum duration for dialing a connection.
		// This includes name resolution if required.
		Timeout time.Duration `json:"timeout" yaml:"timeout"`

		// Deadline specifies an absolute time point after which
		// dial operations will fail.
		// Zero means no deadline.
		Deadline time.Time `json:"deadline" yaml:"deadline"`

		// FallbackDelay specifies the length of time to wait before
		// spawning a fallback connection, when dual-stack IPv4/IPv6
		// is enabled.
		FallbackDelay time.Duration `json:"fallback_delay" yaml:"fallback_delay"`

		// KeepAlive specifies the keep-alive period for network
		// connections. If zero, keep-alives are not enabled.
		KeepAlive time.Duration `json:"keep_alive" yaml:"keep_alive"`
	} `json:"dialer" yaml:"dialer"`
}
