package application

import (
	"encoding/json"
	"os"
)

type stravaConfig struct {
	ClientId              int    `json:"client_id"`
	ClientSecret          string `json:"client_secret"`
	AuthorizationCallback string `json:"authorization_callback"`
}

type openRouteServiceConfig struct {
	ApiKey string `json:"api_key"`
}

type config struct {
	Hostname                  string                  `json:"hostname"`
	DefaultPageLoggedInUsers  string                  `json:"default_page_logged_in"`
	DefaultPageLoggedOutUsers string                  `json:"default_page_logged_out"`
	PathToTemplates           string                  `json:"path_to_templates"`
	SqliteDbPath              string                  `json:"sqlite_db_path"`
	FileDbPath                string                  `json:"file_db_path"`
	StravaConf                *stravaConfig           `json:"strava_config"`
	OrsConf                   *openRouteServiceConfig `json:"open_route_service_config"`
}

func (conf *config) loadFromFile(fileName string) error {
	confContent, err := os.ReadFile(fileName)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(confContent, &conf)
	if err != nil {
		return nil
	}

	return nil
}

func (conf *config) validate() bool {
	//TODO: add real bulletproof validation

	if conf.Hostname == "" {
		return false
	}

	if conf.DefaultPageLoggedInUsers == "" || conf.DefaultPageLoggedOutUsers == "" {
		return false
	}

	if conf.PathToTemplates == "" {
		return false
	}

	if conf.SqliteDbPath == "" || conf.FileDbPath == "" {
		return false
	}

	if conf.StravaConf == nil || conf.OrsConf == nil {
		return false
	}

	if conf.StravaConf.AuthorizationCallback == "" || conf.StravaConf.ClientId == 0 || conf.StravaConf.ClientSecret == "" {
		return false
	}

	if conf.OrsConf.ApiKey == "" {
		return false
	}

	return true
}
