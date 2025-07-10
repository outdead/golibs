package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockTransport struct {
	response *http.Response
	err      error
}

func (m *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return m.response, m.err
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func (f *failingReader) Close() error {
	return nil
}

func TestNewClient(t *testing.T) {
	cfg := &Config{
		Timeout:             0, // Sets default 10 * time.Second
		TLSHandshakeTimeout: 5 * time.Second,
		Dialer: struct {
			Timeout       time.Duration `json:"timeout" yaml:"timeout"`
			Deadline      time.Time     `json:"deadline" yaml:"deadline"`
			FallbackDelay time.Duration `json:"fallback_delay" yaml:"fallback_delay"`
			KeepAlive     time.Duration `json:"keep_alive" yaml:"keep_alive"`
		}{
			Timeout:       2 * time.Second,
			FallbackDelay: 300 * time.Millisecond,
			KeepAlive:     30 * time.Second,
		},
	}

	client := New(cfg)

	if client.Timeout != cfg.Timeout {
		t.Errorf("expected timeout %v, got %v", cfg.Timeout, client.Timeout)
	}
}

func TestSendRequest_Success(t *testing.T) {
	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Create client with default config
	client := New(&Config{
		Timeout: 5 * time.Second,
	})

	// Test successful request
	ctx := context.Background()
	body, err := client.SendRequest(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedBody := `{"status":"ok"}`
	if string(body) != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, string(body))
	}
}

func TestSendRequest_NilContext(t *testing.T) {
	// Create a client with default config
	client := New(&Config{
		Timeout: 5 * time.Second,
	})

	// Test with nil context
	_, err := client.SendRequest(nil, http.MethodGet, "http://example.com", nil)
	if err == nil {
		t.Fatal("expected error when passing nil context, got nil")
	}

	// Check for a specific error type or message
	expectedErr := "net/http: nil Context"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestSendRequest_EOF(t *testing.T) {
	// Setup test server that returns nil response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate connection drop by hijacking and closing connection
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("cannot hijack connection")
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("hijack failed: %v", err)
		}

		conn.Close()
	}))
	defer ts.Close()

	client := New(&Config{
		Timeout: 5 * time.Second,
	})

	ctx := context.Background()
	_, err := client.SendRequest(ctx, http.MethodGet, ts.URL, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expect := fmt.Errorf("Get %q: EOF", ts.URL)
	if err.Error() != expect.Error() {
		t.Errorf("expected error %v, got %v", expect, err)
	}
}

func TestSendRequest_ContextCanceled(t *testing.T) {
	// Setup test server that sleeps to simulate slow response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(&Config{
		Timeout: 5 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := client.SendRequest(ctx, http.MethodGet, ts.URL, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error %v, got %v", context.Canceled, err)
	}
}

func TestSendRequest_Timeout(t *testing.T) {
	// Setup test server that sleeps longer than client timeout
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(&Config{
		Timeout: 100 * time.Millisecond,
	})

	ctx := context.Background()
	_, err := client.SendRequest(ctx, http.MethodGet, ts.URL, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestSendRequest_WrongStatusCode(t *testing.T) {
	// Setup test server that returns 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := New(&Config{
		Timeout: 5 * time.Second,
	})

	ctx := context.Background()
	_, err := client.SendRequest(ctx, http.MethodGet, ts.URL, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrWrongStatusCode) {
		t.Errorf("expected error %v, got %v", ErrWrongStatusCode, err)
	}
}

func TestSendRequest_ReadBodyError(t *testing.T) {
	// Create a client with mock transport
	client := &Client{
		Client: http.Client{
			Transport: &mockTransport{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       &failingReader{},
				},
				err: nil,
			},
		},
	}

	ctx := context.Background()
	_, err := client.SendRequest(ctx, http.MethodGet, "http://example.com", nil)
	if err == nil {
		t.Fatal("expected read error, got nil")
	}

	expectedErr := "simulated read error"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}
