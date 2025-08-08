package helper

import (
	"errors"

	"github.com/miki208/stravaadventuregame/internal/model"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/twpayne/go-polyline"
)

func PointAndIndexAtDistanceAlongLine(ls orb.LineString, distance float64) (orb.Point, int) {
	if len(ls) == 0 {
		panic("empty LineString")
	}

	if distance < 0 || len(ls) == 1 {
		return ls[0], 0
	}

	var (
		travelled = 0.0
		from, to  orb.Point
	)

	for i := 1; i < len(ls); i++ {
		from, to = ls[i-1], ls[i]

		actualSegmentDistance := geo.DistanceHaversine(from, to)
		expectedSegmentDistance := distance - travelled

		if expectedSegmentDistance < actualSegmentDistance {
			bearing := geo.Bearing(from, to)
			return geo.PointAtBearingAndDistance(from, bearing, expectedSegmentDistance), i - 1
		}
		travelled += actualSegmentDistance
	}

	return to, len(ls) - 1
}

func DecodePolyline(coursePolylineEncoded string, reverse bool) (orb.LineString, error) {
	coords, notDecodedBytes, err := polyline.DecodeCoords([]byte(coursePolylineEncoded)) // returns lat, lon pairs
	if err != nil {
		return nil, err
	}

	if len(notDecodedBytes) > 0 {
		return nil, errors.New("not all bytes were decoded from polyline")
	}

	course := make(orb.LineString, 0, len(coords))
	for _, coord := range coords {
		course = append(course, orb.Point{coord[1], coord[0]})
	}

	if reverse {
		course.Reverse()
	}

	return course, nil
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
