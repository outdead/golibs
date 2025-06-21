package httpclient

import "time"

const (
	// DefaultTimeout is the default timeout for general network operations.
	DefaultTimeout = 10 * time.Second

	// DefaultTLSHandshakeTimeout is the default timeout for TLS handshake operations.
	DefaultTLSHandshakeTimeout = 5 * time.Second

	// DefaultDialerTimeout is the default timeout for establishing new connections.
	DefaultDialerTimeout = 5 * time.Second
)

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

// SetDefaults initializes the configuration with default values for any unset fields.
// If Timeout, TLSHandshakeTimeout, or Dialer.Timeout are zero (unset), they will be
// populated with their respective default values.
//
// This ensures the configuration is always valid and prevents zero-values from causing
// unexpected behavior in network operations.
func (cfg *Config) SetDefaults() {
	// Set default general operation timeout if not specified.
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}

	// Set default TLS handshake timeout if not specified.
	if cfg.TLSHandshakeTimeout == 0 {
		cfg.TLSHandshakeTimeout = DefaultTLSHandshakeTimeout
	}

	// Set default dialer connection timeout if not specified.
	if cfg.Dialer.Timeout == 0 {
		cfg.Dialer.Timeout = DefaultDialerTimeout
	}
}
