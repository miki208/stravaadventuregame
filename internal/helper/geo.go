package helper

import (
	"errors"

	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/twpayne/go-polyline"
)

// GetPointFromPolylineAndDistance calculates a point on a polyline at a specified distance from the start and returns its longitude and latitude.
func GetPointFromPolylineAndDistance(coursePolylineEncoded string, reverseDirection bool, distanceMeters float32) (float64, float64, error) {
	coords, notDecodedBytes, err := polyline.DecodeCoords([]byte(coursePolylineEncoded)) // returns lat, lon pairs
	if err != nil {
		return 0, 0, err
	}

	if len(notDecodedBytes) > 0 {
		return 0, 0, errors.New("not all bytes were decoded from polyline")
	}

	course := make(orb.LineString, 0, len(coords))
	for _, coord := range coords {
		course = append(course, orb.Point{coord[1], coord[0]})
	}

	if reverseDirection {
		course.Reverse()
	}

	targetPoint, _ := geo.PointAtDistanceAlongLine(course, float64(distanceMeters))

	return targetPoint.Lon(), targetPoint.Lat(), nil
}

func GetPreferedLocationName(features []model.ReverseGeocodeFeature) string {
	if len(features) == 0 {
		return "unknown location"
	}

	layers := make(map[string]*model.ReverseGeocodeFeature)
	for _, feature := range features {
		layers[feature.Properties.Layer] = &feature
	}

	if len(layers) == 0 {
		return "unknown location"
	}

	result, localityAvailable := layers["locality"]
	if localityAvailable {
		return result.Properties.Label
	}

	result, localAdminAvailable := layers["localadmin"]
	if localAdminAvailable {
		return result.Properties.Label
	}

	result, regionAvailable := layers["region"]
	if regionAvailable {
		return result.Properties.Label
	}

	result, countryAvailable := layers["country"]
	if countryAvailable {
		return result.Properties.Label
	}

	return "unknown location"
}
