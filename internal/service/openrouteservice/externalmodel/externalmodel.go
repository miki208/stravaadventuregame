package externalmodel

type DirectionsSummary struct {
	Distance float32 `json:"distance"`
}

type DirectionsRoute struct {
	Summary  DirectionsSummary `json:"summary"`
	Geometry string            `json:"geometry"`
}
