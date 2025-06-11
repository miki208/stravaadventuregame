package strava

import "github.com/miki208/stravaadventuregame/internal/model"

type TokenExchangeRequest struct {
	ClientId     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
}

type TokenExchangeResponse struct {
	TokenType    string        `json:"token_type"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    int           `json:"expires_at"`
	ExpiresIn    int           `json:"expires_in"`
	Athl         model.Athlete `json:"athlete"`
}
