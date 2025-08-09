package google_handler

import (
	"PLIC/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPlaces(t *testing.T) {
	type testCase struct {
		name           string
		lat, lng       float64
		keyword        string
		expectedPlaces []models.Place
	}

	place := models.Place{
		Name:    "name",
		Address: "address",
		Geometry: struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		}{
			Location: struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}{
				Lat: 48.8566,
				Lng: 2.3522,
			},
		},
	}

	testCases := []testCase{
		{
			name:           "basic valid response",
			lat:            48.8566,
			lng:            2.3522,
			keyword:        "test",
			expectedPlaces: []models.Place{place},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(models.GooglePlacesResponse{
					Results: c.expectedPlaces,
				})
			})

			mockServer := httptest.NewServer(mux)
			defer mockServer.Close()

			result, err := GetPlaces(mockServer.URL, c.lat, c.lng, c.keyword)
			require.NoError(t, err)
			require.Equal(t, c.expectedPlaces, result)
		})
	}
}
