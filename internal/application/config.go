package application

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type stravaConfig struct {
	ClientId                     int    `json:"client_id"`
	ClientSecret                 string `json:"client_secret"`
	AuthorizationCallback        string `json:"authorization_callback"`
	WebhookCallback              string `json:"webhook_callback"`
	Scope                        string `json:"scope"`
	VerifyToken                  string `json:"verify_token"`
	DeleteOldActivitiesAfterDays int    `json:"delete_old_activities_after_days"`
	ProcessWebhookEventsAfterSec int    `json:"process_webhook_events_after_sec"`
}

type openRouteServiceConfig struct {
	ApiKey string `json:"api_key"`
}

type config struct {
	LoggingLevel              string                  `json:"logging_level"`
	Hostname                  string                  `json:"hostname"`
	DefaultPageLoggedInUsers  string                  `json:"default_page_logged_in"`
	DefaultPageLoggedOutUsers string                  `json:"default_page_logged_out"`
	AdminPanelPage            string                  `json:"admin_panel_page"`
	PathToTemplates           string                  `json:"path_to_templates"`
	PathToCertCache           string                  `json:"path_to_cert_cache"`
	SqliteDbPath              string                  `json:"sqlite_db_path"`
	FileDbPath                string                  `json:"file_db_path"`
	StravaConf                *stravaConfig           `json:"strava_config"`
	OrsConf                   *openRouteServiceConfig `json:"open_route_service_config"`
	ScheduledJobIntervalSec   int                     `json:"scheduled_job_interval_sec"`
}

func (conf *config) loadFromFile(fileName string) error {
	confContent, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(confContent, &conf)
	if err != nil {
		return err
	}

	return nil
}

func (conf *config) getLoggingLevel() slog.Level {
	switch conf.LoggingLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (conf *config) validate() error {
	//TODO: add real bulletproof validation

	if conf.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if conf.DefaultPageLoggedInUsers == "" || conf.DefaultPageLoggedOutUsers == "" || conf.AdminPanelPage == "" {
		return fmt.Errorf("default pages cannot be empty")
	}

	if conf.PathToTemplates == "" {
		return fmt.Errorf("path to templates cannot be empty")
	}

	if conf.PathToCertCache == "" {
		return fmt.Errorf("path to certificate cache cannot be empty")
	}

	if conf.SqliteDbPath == "" || conf.FileDbPath == "" {
		return fmt.Errorf("database paths cannot be empty")
	}

	if conf.StravaConf == nil || conf.OrsConf == nil {
		return fmt.Errorf("strava and openrouteservice configurations cannot be nil")
	}

	if conf.StravaConf.AuthorizationCallback == "" || conf.StravaConf.ClientId == 0 || conf.StravaConf.ClientSecret == "" ||
		conf.StravaConf.Scope == "" || conf.StravaConf.WebhookCallback == "" || conf.StravaConf.VerifyToken == "" || conf.StravaConf.DeleteOldActivitiesAfterDays < 1 ||
		conf.StravaConf.ProcessWebhookEventsAfterSec < conf.ScheduledJobIntervalSec {

		return fmt.Errorf("strava configuration is invalid")
	}

	if conf.OrsConf.ApiKey == "" {
		return fmt.Errorf("openrouteservice configuration is invalid")
	}

	if conf.ScheduledJobIntervalSec < 60 {
		return fmt.Errorf("scheduled job interval must be at least 60 seconds")
	}

	return nil
}
