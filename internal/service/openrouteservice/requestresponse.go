package openrouteservice

import "github.com/miki208/stravaadventuregame/internal/model"

type DirectionsRequest struct {
	Coordinates [][]float64 `json:"coordinates"`
	Units       string      `json:"units"`
}

type DirectionsResponse struct {
	Routes []model.DirectionsRoute `json:"routes"`
}
