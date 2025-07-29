package models

type Place struct {
	Name    string `json:"name"`
	Address string `json:"vicinity"`

	Geometry struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
}

type GooglePlacesResponse struct {
	Results []Place `json:"results"`
}
