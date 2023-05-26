package main

import "net/http"

func healthHandler(downstreamUrl string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform healthcheck on the downstream URL
		// If an error is returned or the status code is not 200, we consider
		// for now that the downstream is not healthy
		if res, err := http.Get(downstreamUrl); err != nil || res.StatusCode != 200 {
			JSONResponseWithStatusCode(w, http.StatusServiceUnavailable, &HealthResponse{
				Status: "DOWN",
			})
			return
		}

		JSONResponse(w, &HealthResponse{
			Status: "UP",
		})
	}
}
