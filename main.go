package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/miki208/stravaadventuregame/internal/application"
	"github.com/miki208/stravaadventuregame/internal/handler"
	"github.com/miki208/stravaadventuregame/internal/handler/auth"
	"github.com/miki208/stravaadventuregame/internal/handler/noauth"
	"github.com/miki208/stravaadventuregame/internal/scheduledjobs"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	configFileName := flag.String("config", "config.ini", "Path to the configuration file")
	flag.Parse()

	app := application.MakeApp(*configFileName)
	defer app.Close()

	slog.Info("Application started.", "hostname", app.Hostname)

	slog.Info("Registering scheduled jobs...")
	for _, job := range scheduledjobs.GetScheduledJobs() {
		app.CronSvc.AddJob(job)
	}
	slog.Info("Scheduled jobs registered.")

	app.CronSvc.Start()
	defer app.CronSvc.Stop()

	// http mux and server setup (needed for Let's Encrypt certificate management)
	domains := []string{app.Hostname, "www." + app.Hostname}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache(app.PathToCertCache),
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", certManager.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	})))

	httpServer := &http.Server{
		Addr:    ":http",
		Handler: httpMux,
	}

	// create main (https) httpsMux and server
	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc(app.DefaultPageLoggedOutUsers, handler.MakeHandlerWoutSession(app, noauth.Authorize))
	httpsMux.HandleFunc(app.StravaSvc.GetAuthorizationCallback(), handler.MakeHandlerWoutSession(app, noauth.StravaAuthCallback))
	httpsMux.HandleFunc(app.DefaultPageLoggedInUsers, handler.MakeHandlerWSession(app, auth.Welcome))
	httpsMux.HandleFunc("/start-adventure", handler.MakeHandlerWSession(app, auth.StartAdventure))
	httpsMux.HandleFunc("/logout", handler.MakeHandlerWSession(app, auth.Logout))
	httpsMux.HandleFunc("/deauthorize", handler.MakeHandlerWSession(app, auth.Deauthorize))
	httpsMux.HandleFunc(app.AdminPanelPage, handler.MakeHandlerWSession(app, auth.AdminPanel))
	httpsMux.HandleFunc("/stravawebhook/delete", handler.MakeHandlerWSession(app, auth.DeleteStravaWebhookSubscription))
	httpsMux.HandleFunc("/stravawebhook/create", handler.MakeHandlerWSession(app, auth.CreateStravaWebhookSubscription))
	httpsMux.HandleFunc(app.StravaSvc.GetWebhookCallback(), handler.MakeHandlerWoutSession(app, noauth.StravaWebhookCallback))

	httpsServer := &http.Server{
		Addr:    ":https",
		Handler: httpsMux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	// termination handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting HTTP server for Let's Encrypt...")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe error.", "error", err)
		}
	}()

	go func() {
		slog.Info("Starting HTTPS server...")

		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServeTLS error.", "error", err)
		}
	}()

	<-quit

	// shutdown the server (gracefully or with timeout)
	slog.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpsServer.Shutdown(ctx); err != nil {
		slog.Error("HTTPS server shutdown error.", "error", err)
	}

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error.", "error", err)
	}

	slog.Info("Servers shutdown complete.")
}
