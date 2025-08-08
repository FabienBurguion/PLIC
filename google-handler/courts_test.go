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
		_ = json.NewEncoder(w).Encode(models.GooglePlacesResponse{
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
