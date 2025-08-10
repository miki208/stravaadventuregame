package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// interface for server operations
type Server interface {
	ListenAndServe()
	AddRoute(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// concrete implementations of the Server interface
type HTTPServerOptions struct {
	InsecurePort    int
	ProxyPathPrefix string
}

func NewHTTPServerOptions(insecurePort int, proxyPathPrefix string) *HTTPServerOptions {
	return &HTTPServerOptions{
		InsecurePort:    insecurePort,
		ProxyPathPrefix: proxyPathPrefix,
	}
}

type HTTPServer struct {
	httpMux *http.ServeMux

	httpServer *http.Server
}

func (s *HTTPServer) AddRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.httpMux.HandleFunc(pattern, handler)
}

func (s *HTTPServer) ListenAndServe() {
	// termination handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting HTTP server...")

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe error.", "error", err)
		}
	}()

	<-quit

	// shutdown the server (gracefully or with timeout)
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error.", "error", err)
	}

	slog.Info("Server shutdown complete.")
}

// concrete implementation of the HTTPSServer
type HTTPSServerOptions struct {
	PathToCertCache string
	ProxyPathPrefix string
}

func NewHTTPSServerOptions(pathToCertCache string, proxyPathPrefix string) *HTTPSServerOptions {
	return &HTTPSServerOptions{
		PathToCertCache: pathToCertCache,
		ProxyPathPrefix: proxyPathPrefix,
	}
}

type HTTPSServer struct {
	httpsMux *http.ServeMux

	httpServer  *http.Server
	httpsServer *http.Server
}

func (s *HTTPSServer) AddRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.httpsMux.HandleFunc(pattern, handler)
}

func (s *HTTPSServer) ListenAndServe() {
	// termination handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting HTTP server for Let's Encrypt...")

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe error.", "error", err)
		}
	}()

	go func() {
		slog.Info("Starting HTTPS server...")

		if err := s.httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServeTLS error.", "error", err)
		}
	}()

	<-quit

	// shutdown the server (gracefully or with timeout)
	slog.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpsServer.Shutdown(ctx); err != nil {
		slog.Error("HTTPS server shutdown error.", "error", err)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error.", "error", err)
	}

	slog.Info("Servers shutdown complete.")
}

// ServerFactory is responsible for creating server instances
type ServerFactory struct {
	Hostname string
}

func (factory *ServerFactory) CreateServer(options any) Server {
	switch opts := options.(type) {
	case *HTTPServerOptions:
		result := &HTTPServer{
			httpMux: http.NewServeMux(),
		}

		result.httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.InsecurePort),
			Handler: result.httpMux,
		}

		return result

	case *HTTPSServerOptions:
		result := &HTTPSServer{
			httpsMux: http.NewServeMux(),
		}

		domains := []string{factory.Hostname, "www." + factory.Hostname}

		certManager := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
			Cache:      autocert.DirCache(opts.PathToCertCache),
		}

		httpMux := http.NewServeMux()
		httpMux.Handle("/", certManager.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+opts.ProxyPathPrefix+r.RequestURI, http.StatusMovedPermanently)
		})))

		result.httpServer = &http.Server{
			Addr:    ":http",
			Handler: httpMux,
		}

		result.httpsServer = &http.Server{
			Addr:    ":https",
			Handler: result.httpsMux,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
				MinVersion:     tls.VersionTLS12,
			},
		}

		return result
	default:
		return nil
	}
}
