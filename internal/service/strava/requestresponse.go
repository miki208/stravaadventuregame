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

type TokenRefreshRequest struct {
	ClientId     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type DeauthorizationRequest struct {
	AccessToken string `json:"access_token"`
}

type SubscriptionCreationRequest struct {
	ClientId     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CallbackUrl  string `json:"callback_url"`
	VerifyToken  string `json:"verify_token"`
}

type SubscriptionCreationResponse struct {
	Id int `json:"id"`
}

type CallbackValidationResponse struct {
	HubChallenge string `json:"hub.challenge"`
}
