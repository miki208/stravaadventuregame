package externalmodel

type DirectionsRequest struct {
	Coordinates [][]float64 `json:"coordinates"`
	Units       string      `json:"units"`
}

type DirectionsResponse struct {
	Routes []DirectionsRoute `json:"routes"`
}

type ReverseGeocodeResponse struct {
	Features []ReverseGeocodeFeature `json:"features"`
}
