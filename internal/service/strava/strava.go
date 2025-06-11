package strava

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/model"
)

type Strava struct {
	clientId              int
	clientSecret          string
	AuthorizationCallback string
	baseUrl               string
}

func CreateService(clientId int, clientSecret, authorizationCallback string) *Strava {
	return &Strava{
		clientId:              clientId,
		clientSecret:          clientSecret,
		AuthorizationCallback: authorizationCallback,
		baseUrl:               "https://www.strava.com/api/v3",
	}
}

func (svc *Strava) GetClientId() int {
	return svc.clientId
}

func (svc *Strava) ExchangeToken(authorizationCode string) (*model.Athlete, *model.StravaCredential, error) {
	tokenExchangeBody, err := json.Marshal(TokenExchangeRequest{
		ClientId:     svc.clientId,
		ClientSecret: svc.clientSecret,
		Code:         authorizationCode,
		GrantType:    "authorization_code",
	})
	if err != nil {
		return nil, nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	tokenExchangeResponse, err := http.Post(svc.baseUrl+"/oauth/token", "application/json", bytes.NewBuffer(tokenExchangeBody))
	if err != nil {
		return nil, nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	defer tokenExchangeResponse.Body.Close()

	tokenExchangeResponseBody, err := io.ReadAll(tokenExchangeResponse.Body)
	if err != nil {
		return nil, nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if tokenExchangeResponse.StatusCode != http.StatusOK {
		return nil, nil, &StravaError{statusCode: tokenExchangeResponse.StatusCode, err: errors.New("token exchange failed")}
	}

	var tokenExchangeResponseObj TokenExchangeResponse
	err = json.Unmarshal(tokenExchangeResponseBody, &tokenExchangeResponseObj)
	if err != nil {
		return nil, nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	stravaCredential := model.StravaCredential{
		AthleteId:    tokenExchangeResponseObj.Athl.Id,
		AccessToken:  tokenExchangeResponseObj.AccessToken,
		RefreshToken: tokenExchangeResponseObj.RefreshToken,
		ExpiresAt:    tokenExchangeResponseObj.ExpiresAt,
	}

	return &tokenExchangeResponseObj.Athl, &stravaCredential, nil
}
