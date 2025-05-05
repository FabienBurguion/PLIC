package google_handler

import (
	"PLIC/models"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
	func TestDatabase_GetPlaces(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mockResponse := `{
			"results": [
				{
					"name": "Mock Sports Center",
					"vicinity": "789 Mock Blvd",
					"geometry": {
						"location": {
							"lat": 48.8566,
							"lng": 2.3522
						}
					}
				}
			]
		}`

		url := "https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=48.856600,2.352200&radius=1000&type=sports_complex&key=fake-key"
		httpmock.RegisterResponder("GET", url,
			httpmock.NewStringResponder(200, mockResponse),
		)

		db := Database{}
		//ctx := context.Background()

		places, err := db.GetPlaces(48.8566, 2.3522, "fake-key")
		require.NoError(t, err)
		require.Len(t, places, 1)
		require.Equal(t, "Mock Sports Center", places[0].Name)
		require.Equal(t, "789 Mock Blvd", places[0].Address)
	}
*/
func TestGetPlaces(t *testing.T) {
	mux := http.NewServeMux()
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
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.GooglePlacesResponse{
			Results: []models.Place{place},
		})
	})

	mockServer := httptest.NewServer(mux)
	defer mockServer.Close()

	result, err := GetPlaces(mockServer.URL, 48.8566, 2.3522, "test")
	if err != nil {
		t.Fatalf("Erreur: %v", err)
	}

	expected := place
	require.Equal(t, expected, result[0])
}
