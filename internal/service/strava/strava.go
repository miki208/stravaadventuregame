package strava

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/miki208/stravaadventuregame/internal/model"
)

type Strava struct {
	clientId              int
	clientSecret          string
	authorizationCallback string
	baseUrl               string
	scope                 string
}

func CreateService(clientId int, clientSecret, authorizationCallback, scope string) *Strava {
	return &Strava{
		clientId:              clientId,
		clientSecret:          clientSecret,
		authorizationCallback: authorizationCallback,
		baseUrl:               "https://www.strava.com/api/v3",
		scope:                 scope,
	}
}

func (svc *Strava) GetClientId() int {
	return svc.clientId
}

func (svc *Strava) GetAuthorizationCallback() string {
	return svc.authorizationCallback
}

func (svc *Strava) GetScope() string {
	return svc.scope
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

func (svc *Strava) Deauthorize(athleteId int, db *sql.DB, tx *sql.Tx) error {
	// retrieve athlete and credentials to ensure they exist
	credentials, err := svc.GetCredentialsForAthlete(athleteId, db, tx)
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	var athlete model.Athlete
	exists, err := athlete.LoadById(athleteId, db, tx)
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if !exists {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: http.StatusNotFound, err: errors.New("athlete not found")}
	}

	// deauthorize the athlete
	deauthorizationBody, err := json.Marshal(DeauthorizationRequest{
		AccessToken: credentials.AccessToken,
	})
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	deauthorizationResponse, err := http.Post(svc.baseUrl+"/oauth/deauthorize", "application/json", bytes.NewBuffer(deauthorizationBody))
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	defer deauthorizationResponse.Body.Close()

	_, err = io.ReadAll(deauthorizationResponse.Body)
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if deauthorizationResponse.StatusCode != http.StatusOK {
		if tx != nil {
			tx.Rollback()
		}

		return &StravaError{statusCode: deauthorizationResponse.StatusCode, err: errors.New("deauthorization failed")}
	}

	// delete the athlete (and everything related to it)
	err = athlete.Delete(db, tx)
	if tx != nil {
		tx.Rollback()

		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	return nil
}

func (svc *Strava) refreshTokenIfNeeded(cred *model.StravaCredential) (bool, error) {
	expiresAt := time.Unix(int64(cred.ExpiresAt), 0)
	if !expiresAt.Before(time.Now()) && time.Until(expiresAt) >= 5*time.Minute {
		return false, nil
	}

	tokenRefreshBody, err := json.Marshal(TokenRefreshRequest{
		ClientId:     svc.clientId,
		ClientSecret: svc.clientSecret,
		GrantType:    "refresh_token",
		RefreshToken: cred.RefreshToken,
	})
	if err != nil {
		return false, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	tokenRefreshResponse, err := http.Post(svc.baseUrl+"/oauth/token", "application/json", bytes.NewBuffer(tokenRefreshBody))
	if err != nil {
		return false, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	defer tokenRefreshResponse.Body.Close()

	tokenRefreshResponseBody, err := io.ReadAll(tokenRefreshResponse.Body)
	if err != nil {
		return false, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if tokenRefreshResponse.StatusCode != http.StatusOK {
		return false, &StravaError{statusCode: tokenRefreshResponse.StatusCode, err: errors.New("token refresh failed")}
	}

	var tokenRefreshResponseObj TokenRefreshResponse
	err = json.Unmarshal(tokenRefreshResponseBody, &tokenRefreshResponseObj)
	if err != nil {
		return false, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	cred.AccessToken = tokenRefreshResponseObj.AccessToken
	cred.RefreshToken = tokenRefreshResponseObj.RefreshToken
	cred.ExpiresAt = tokenRefreshResponseObj.ExpiresAt

	return true, nil
}

func (svc *Strava) GetCredentialsForAthlete(athleteId int, db *sql.DB, tx *sql.Tx) (*model.StravaCredential, error) {
	cred := &model.StravaCredential{}

	exists, err := cred.LoadByAthleteId(athleteId, db, tx)
	if err != nil {
		return nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if !exists {
		return nil, nil
	}

	refreshed, err := svc.refreshTokenIfNeeded(cred)
	if err != nil {
		return nil, err
	}

	if refreshed {
		err = cred.Save(db, tx)
		if err != nil {
			return nil, err
		}
	}

	return cred, nil
}

func (svc *Strava) ValidateScope(scopeGiven string) bool {
	scopeGivenParts := strings.Split(scopeGiven, ",")
	for _, part := range strings.Split(svc.scope, ",") {
		if !slices.Contains(scopeGivenParts, part) {
			return false
		}
	}

	return true
}
