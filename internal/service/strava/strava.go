package strava

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/strava/externalmodel"
)

type Strava struct {
	clientId                     int
	clientSecret                 string
	authorizationCallback        string
	baseUrl                      string
	scope                        string
	webhookCallback              string
	verifyToken                  string // TODO: this should be a random string in future
	deleteOldActivitiesAfterDays int
}

func CreateService(clientId int, clientSecret, authorizationCallback, scope, webhookCallback, verifyToken string, deleteOldActivitiesAfterDays int) *Strava {
	return &Strava{
		clientId:                     clientId,
		clientSecret:                 clientSecret,
		authorizationCallback:        authorizationCallback,
		baseUrl:                      "https://www.strava.com/api/v3",
		scope:                        scope,
		webhookCallback:              webhookCallback,
		verifyToken:                  verifyToken,
		deleteOldActivitiesAfterDays: deleteOldActivitiesAfterDays,
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

func (svc *Strava) GetWebhookCallback() string {
	return svc.webhookCallback
}

func (svc *Strava) GetVerifyToken() string {
	return svc.verifyToken
}

func (svc *Strava) GetDeleteOldActivitiesAfterDays() int {
	return svc.deleteOldActivitiesAfterDays
}

func (svc *Strava) ExchangeToken(authorizationCode string) (*model.Athlete, *model.StravaCredential, error) {
	tokenExchangeBody, err := json.Marshal(externalmodel.TokenExchangeRequest{
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

	var tokenExchangeResponseObj externalmodel.TokenExchangeResponse
	err = json.Unmarshal(tokenExchangeResponseBody, &tokenExchangeResponseObj)
	if err != nil {
		return nil, nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	var internalStravaCredential model.StravaCredential
	internalStravaCredential.FromExternalModel(&tokenExchangeResponseObj)

	var internalAthlete model.Athlete
	internalAthlete.FromExternalModel(&tokenExchangeResponseObj.Athl)

	return &internalAthlete, &internalStravaCredential, nil
}

func (svc *Strava) Deauthorize(athleteId int64, sendDeauthRequest bool, db *sql.DB, tx *sql.Tx) error {
	// retrieve athlete and credentials to ensure they exist
	credentials, err := svc.GetCredentialsForAthlete(athleteId, db, tx)
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	athlete := model.NewAthlete()
	exists, err := athlete.Load(athleteId, db, tx)
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if !exists {
		return &StravaError{statusCode: http.StatusNotFound, err: errors.New("athlete not found")}
	}

	if sendDeauthRequest {
		// deauthorize the athlete
		deauthorizationBody, err := json.Marshal(externalmodel.DeauthorizationRequest{
			AccessToken: credentials.AccessToken,
		})
		if err != nil {
			return &StravaError{statusCode: http.StatusInternalServerError, err: err}
		}

		deauthorizationResponse, err := http.Post(svc.baseUrl+"/oauth/deauthorize", "application/json", bytes.NewBuffer(deauthorizationBody))
		if err != nil {
			return &StravaError{statusCode: http.StatusInternalServerError, err: err}
		}

		defer deauthorizationResponse.Body.Close()

		_, err = io.ReadAll(deauthorizationResponse.Body)
		if err != nil {
			return &StravaError{statusCode: http.StatusInternalServerError, err: err}
		}

		if deauthorizationResponse.StatusCode != http.StatusOK {
			return &StravaError{statusCode: deauthorizationResponse.StatusCode, err: errors.New("deauthorization failed")}
		}
	}

	// delete the athlete (and everything related to it)
	err = athlete.Delete(db, tx)
	if tx != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	return nil
}

func (svc *Strava) refreshTokenIfNeeded(cred *model.StravaCredential) (bool, error) {
	expiresAt := time.Unix(int64(cred.ExpiresAt), 0)
	if !expiresAt.Before(time.Now()) && time.Until(expiresAt) >= 5*time.Minute {
		return false, nil
	}

	tokenRefreshBody, err := json.Marshal(externalmodel.TokenRefreshRequest{
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

	var tokenRefreshResponseObj externalmodel.TokenRefreshResponse
	err = json.Unmarshal(tokenRefreshResponseBody, &tokenRefreshResponseObj)
	if err != nil {
		return false, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	cred.AccessToken = tokenRefreshResponseObj.AccessToken
	cred.RefreshToken = tokenRefreshResponseObj.RefreshToken
	cred.ExpiresAt = tokenRefreshResponseObj.ExpiresAt

	return true, nil
}

func (svc *Strava) GetCredentialsForAthlete(athleteId int64, db *sql.DB, tx *sql.Tx) (*model.StravaCredential, error) {
	cred := &model.StravaCredential{}

	exists, err := cred.Load(athleteId, db, tx)
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

func (svc *Strava) CreateSubscription(fullCallbackUrl string) (int, error) {
	subscriptionCreationBody, err := json.Marshal(externalmodel.SubscriptionCreationRequest{
		ClientId:     svc.clientId,
		ClientSecret: svc.clientSecret,
		CallbackUrl:  fullCallbackUrl,
		VerifyToken:  svc.verifyToken,
	})
	if err != nil {
		return 0, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	subscriptionCreationResponse, err := http.Post(svc.baseUrl+"/push_subscriptions", "application/json", bytes.NewBuffer(subscriptionCreationBody))
	if err != nil {
		return 0, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	defer subscriptionCreationResponse.Body.Close()

	subscriptionCreationResponseBody, err := io.ReadAll(subscriptionCreationResponse.Body)
	if err != nil {
		return 0, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	if subscriptionCreationResponse.StatusCode != http.StatusCreated {
		return 0, &StravaError{statusCode: subscriptionCreationResponse.StatusCode, err: errors.New("subscription creation failed")}
	}

	var subscriptionCreationResponseObj externalmodel.SubscriptionCreationResponse
	err = json.Unmarshal(subscriptionCreationResponseBody, &subscriptionCreationResponseObj)
	if err != nil {
		return 0, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	return subscriptionCreationResponseObj.Id, nil
}

func (svc *Strava) DeleteSubscription(subscriptionId int) error {
	u, err := url.Parse(svc.baseUrl + "/push_subscriptions/" + strconv.Itoa(subscriptionId))
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	query := u.Query()
	query.Set("client_id", strconv.Itoa(svc.clientId))
	query.Set("client_secret", svc.clientSecret)
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return &StravaError{statusCode: resp.StatusCode, err: errors.New("subscription deletion failed")}
	}

	return nil
}

func (svc *Strava) GetActivity(athleteId, activityId int64, db *sql.DB, tx *sql.Tx) (*model.Activity, error) {
	req, err := http.NewRequest(http.MethodGet, svc.baseUrl+"/activities/"+strconv.FormatInt(activityId, 10), nil)
	if err != nil {
		return nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	credential, err := svc.GetCredentialsForAthlete(athleteId, db, tx)
	if err != nil {
		return nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	req.Header.Set("Authorization", "Bearer "+credential.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &StravaError{statusCode: resp.StatusCode, err: errors.New("failed to get activity from strava")}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	var activity externalmodel.Activity
	err = json.Unmarshal(respBody, &activity)
	if err != nil {
		return nil, &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	var internalActivity model.Activity
	internalActivity.FromExternalModel(&activity)

	internalActivity.AthleteId = athleteId

	return &internalActivity, nil
}

func SendCallbackValidationResponse(hubChallenge string, resp http.ResponseWriter) error {
	challengeResponse := externalmodel.CallbackValidationResponse{HubChallenge: hubChallenge}
	challengeResponseJson, err := json.Marshal(&challengeResponse)
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)

	_, err = resp.Write(challengeResponseJson)
	if err != nil {
		return &StravaError{statusCode: http.StatusInternalServerError, err: err}
	}

	return nil
}
