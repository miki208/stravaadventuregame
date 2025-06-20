package externalmodel

type DirectionsSummary struct {
	Distance float32 `json:"distance"`
}

type DirectionsRoute struct {
	Summary  DirectionsSummary `json:"summary"`
	Geometry string            `json:"geometry"`
}

type ReverseGeocodeProperties struct {
	Layer      string `json:"layer"`
	Country    string `json:"country"`
	Region     string `json:"region"`
	LocalAdmin string `json:"localadmin"`
	Label      string `json:"label"`
}

type ReverseGeocodeFeature struct {
	Properties ReverseGeocodeProperties `json:"properties"`
}
