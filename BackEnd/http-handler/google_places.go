package main

import (
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

func (s *Service) SyncGooglePlaces(ctx context.Context, latitude, longitude float64, apiKey string) error {
	places, err := s.db.GetPlaces(latitude, longitude, apiKey)
	log.Println(err)
	log.Println(places)
	if err != nil {
		return fmt.Errorf("échec de la récupération des lieux via Google Maps API : %w", err)
	}

	// Boucle sur les résultats
	for _, place := range places {
		id := uuid.NewString()
		err := s.db.InsertTerrain(ctx, id, place)
		log.Println(err)
		if err != nil {
			// On log l’erreur mais on continue pour les autres
			log.Printf("Erreur lors de l'insertion du terrain \"%s\": %v", place.Name, err)
		}
	}

	return nil
}

func (s *Service) HandleSyncGooglePlaces(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	lat := 48.8566
	lng := 2.3522
	apiKey := s.configuration.Google.ApiKey
	log.Println(apiKey)

	err := s.SyncGooglePlaces(ctx, lat, lng, apiKey)
	if err != nil {
		log.Println(err)
		return httpx.WriteError(w, http.StatusInternalServerError, "erreur lors de la synchro des terrains")
	}
	return httpx.Write(w, http.StatusCreated, nil)

}
