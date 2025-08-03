package httpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/outdead/golibs/httpserver/problemdetails"
	"github.com/outdead/golibs/httpserver/validator"
)

// ShutdownTimeOut is time to terminate queries when quit signal given.
const ShutdownTimeOut = 10 * time.Second

// ErrLockedServer returned on repeated call Close() the HTTP server.
var ErrLockedServer = errors.New("http server is locked")

// Logger describes Error and Info functions.
type Logger interface {
	Infof(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
	Writer() io.Writer
}

// Binder represents Bind and BindAndValidate functions.
type Binder interface {
	Bind(c echo.Context, req interface{}) error
	BindAndValidate(c echo.Context, req interface{}) error
}

// Responder represents ServeResult and ServeError functions.
type Responder interface {
	ServeResult(c echo.Context, i interface{}) error
	ServeError(c echo.Context, err error, logPrefix ...string) error
}

// Option infects params to Server.
type Option func(server *Server)

// WithBinder injects Binder to Server.
func WithBinder(binder Binder) Option {
	return func(s *Server) {
		s.Binder = binder
	}
}

// WithResponder injects Responder to Server.
func WithResponder(responser Responder) Option {
	return func(s *Server) {
		s.Responder = responser
	}
}

func WithRecover(rec bool) Option {
	return func(s *Server) {
		s.recover = rec
	}
}

// A Server defines parameters for running an HTTP server.
type Server struct {
	Binder
	Responder
	Echo *echo.Echo

	logger  Logger
	errors  chan error
	recover bool
	quit    chan bool
	wg      sync.WaitGroup
}

// NewServer allocates and returns a new Server.
func NewServer(log Logger, errs chan error, options ...Option) *Server {
	s := Server{
		logger: log,
		errors: errs,

		quit: make(chan bool),
	}

	for _, option := range options {
		option(&s)
	}

	if s.Binder == nil {
		s.Binder = problemdetails.NewBinder(s.logger)
	}

	if s.Responder == nil {
		s.Responder = problemdetails.NewResponder(s.logger)
	}

	s.Echo = s.newEcho()

	return &s
}

// Serve initializes HTTP Server and runs it on received port.
func (s *Server) Serve(port string) {
	go func() {
		if err := s.Echo.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Report error if server is not closed by Echo#Shutdown.
			s.ReportError(fmt.Errorf("start http server: %w", err))
		}
	}()

	s.logger.Infof("http server started on port %s", port)

	s.quit = make(chan bool)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		<-s.quit
		s.logger.Debug("stopping http server...")

		ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeOut)
		defer cancel()

		if s.Echo != nil {
			if err := s.Echo.Shutdown(ctx); err != nil {
				s.ReportError(fmt.Errorf("shutdown http server: %w", err))
			}
		}
	}()
}

// ReportError publishes error to the errors channel.
// if you do not read errors from the errors channel then after the channel
// buffer overflows the application exits with a fatal level and the
// os.Exit(1) exit code.
func (s *Server) ReportError(err error) {
	if err != nil {
		select {
		case s.errors <- err:
		default:
			s.logger.Fatalf("http server error channel is locked: %s", err)
		}
	}
}

// Close stops HTTP Server.
func (s *Server) Close() error {
	if s.quit == nil {
		return ErrLockedServer
	}

	select {
	case s.quit <- true:
		s.wg.Wait()
		s.logger.Debug("stop http server success")

		return nil
	default:
		return ErrLockedServer
	}
}

func (s *Server) newEcho() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORS())

	if s.recover {
		e.Use(middleware.Recover())
	}

	e.Validator = validator.New()

	e.Logger.SetOutput(s.logger.Writer())
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = s.httpErrorHandler

	return e
}

// httpErrorHandler customizes error response.
// @source: https://github.com/labstack/echo/issues/325
func (s *Server) httpErrorHandler(err error, c echo.Context) {
	if err = s.ServeError(c, err); err != nil {
		s.ReportError(fmt.Errorf("error handle: %w", err))
	}
}
