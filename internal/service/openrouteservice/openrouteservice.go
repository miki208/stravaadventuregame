package openrouteservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

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
		baseUrl:    "https://api.openrouteservice.org",
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

	directionsRequest, err := http.NewRequest(http.MethodPost, ors.baseUrl+"/v2/directions/driving-car", bytes.NewBuffer(directionsRequestJson))
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

func (ors *OpenRouteService) ReverseGeocode(lon, lat float64, numOfResults int, layers string) ([]model.ReverseGeocodeFeature, error) {
	u, err := url.Parse(ors.baseUrl + "/geocode/reverse")
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusInternalServerError, err: err}
	}

	query := u.Query()
	query.Set("point.lon", strconv.FormatFloat(lon, 'f', -1, 64))
	query.Set("point.lat", strconv.FormatFloat(lat, 'f', -1, 64))
	query.Set("size", strconv.Itoa(numOfResults))
	if layers != "" {
		query.Set("layers", layers)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusInternalServerError, err: err}
	}

	req.Header.Set("Authorization", ors.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ors.httpClient.Do(req)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: err}
	}

	defer resp.Body.Close()

	reverseGeocodeResponseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: err}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &OpenRouteServiceError{statusCode: http.StatusFailedDependency, err: errors.New("reverse geocoding via external api failed")}
	}

	var reverseGeocodeResponseObj externalmodel.ReverseGeocodeResponse
	err = json.Unmarshal(reverseGeocodeResponseBody, &reverseGeocodeResponseObj)
	if err != nil {
		return nil, &OpenRouteServiceError{statusCode: http.StatusInternalServerError, err: err}
	}

	var result []model.ReverseGeocodeFeature
	for _, feature := range reverseGeocodeResponseObj.Features {
		result = append(result, model.ReverseGeocodeFeature{
			ReverseGeocodeFeature: &feature,
		})
	}

	return result, nil
}
