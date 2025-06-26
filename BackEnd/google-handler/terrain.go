package google_handler

import (
	"PLIC/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GetPlaces(baseUrl string, latitude, longitude float64, apiKey string) ([]models.Place, error) {
	url := fmt.Sprintf(
		"%s/json?location=%f,%f&radius=1000&type=sports_complex&key=%s",
		baseUrl, latitude, longitude, apiKey)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	var data models.GooglePlacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}
