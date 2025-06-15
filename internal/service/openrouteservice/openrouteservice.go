package openrouteservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/miki208/stravaadventuregame/internal/service/openrouteservice/externalmodel"
)

type OpenRouteService struct {
	apiKey string

	httpClient http.Client
	baseUrl    string
}

func CreateService(apiKey string) *OpenRouteService {
	return &OpenRouteService{
		apiKey: apiKey,

		httpClient: http.Client{},
		baseUrl:    "https://api.openrouteservice.org/v2",
	}
}

func (ors *OpenRouteService) GetDirections(latStart, lonStart, latEnd, lonEnd float64, units string) (*model.DirectionsRoute, error) {
	directionsRequestJson, err := json.Marshal(&externalmodel.DirectionsRequest{
		Coordinates: [][]float64{{lonStart, latStart}, {lonEnd, latEnd}},
		Units:       units,
	})
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusInternalServerError, err: err}
	}

	directionsRequest, err := http.NewRequest(http.MethodPost, ors.baseUrl+"/directions/driving-car", bytes.NewBuffer(directionsRequestJson))
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusInternalServerError, err: err}
	}

	directionsRequest.Header.Set("Authorization", ors.apiKey)
	directionsRequest.Header.Set("Content-Type", "application/json")

	directionsResponse, err := ors.httpClient.Do(directionsRequest)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: err}
	}

	defer directionsResponse.Body.Close()

	directionsResponseBody, err := io.ReadAll(directionsResponse.Body)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: err}
	}

	if directionsResponse.StatusCode != http.StatusOK {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: errors.New("getting directions via external api failed")}
	}

	var directionsResponseObj externalmodel.DirectionsResponse
	err = json.Unmarshal(directionsResponseBody, &directionsResponseObj)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusInternalServerError, err: err}
	}

	if len(directionsResponseObj.Routes) < 1 {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: errors.New("routes returned from the external api are empty")}
	}

	internalDirectionsRoute := &model.DirectionsRoute{}
	internalDirectionsRoute.FromExternalModel(&directionsResponseObj.Routes[0])

	return internalDirectionsRoute, nil
}
