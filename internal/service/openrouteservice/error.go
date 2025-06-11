package openrouteservice

import "fmt"

type OpenRouteServiceError struct {
	statusCode int
	err        error
}

func (orsError *OpenRouteServiceError) StatusCode() int {
	return orsError.statusCode
}

func (orsError *OpenRouteServiceError) Error() string {
	return fmt.Sprintf("OpenRouteService error (%d): %v", orsError.statusCode, orsError.err)
}
