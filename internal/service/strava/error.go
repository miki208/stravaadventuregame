package strava

import "fmt"

type StravaError struct {
	statusCode int
	err        error
}

func (stravaError *StravaError) StatusCode() int {
	return stravaError.statusCode
}

func (stravaError *StravaError) Error() string {
	return fmt.Sprintf("Strava error (%d): %v", stravaError.statusCode, stravaError.err)
}
