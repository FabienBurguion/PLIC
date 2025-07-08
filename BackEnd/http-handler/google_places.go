package main

import (
	google_handler "PLIC/google-handler"
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

func (s *Service) SyncGooglePlaces(ctx context.Context, latitude, longitude float64, apiKey string) error {
	places, err := google_handler.GetPlaces("https://maps.googleapis.com/maps/api/place/nearbysearch", latitude, longitude, apiKey)
	if err != nil {
		return fmt.Errorf("échec de la récupération des lieux via Google Maps API : %w", err)
	}

	for _, place := range places {
		id := uuid.NewString()
		err := s.db.InsertTerrain(ctx, id, place, s.clock.Now())
		if err != nil {
			log.Printf("Erreur lors de l'insertion du terrain \"%s\": %v", place.Name, err)
		}
	}

	return nil
}

// HandleSyncGooglePlaces godoc
// @Summary      Synchronise les terrains depuis l'API Google Places
// @Description  Appelle l'API Google Places pour synchroniser les terrains autour d'une position donnée (Paris en dur pour l'instant)
// @Tags         google
// @Produce      json
// @Success      201 {object} nil "Synchro réussie"
// @Failure      500 {object} models.Error "Erreur lors de la synchronisation"
// @Router       /place [post]
func (s *Service) HandleSyncGooglePlaces(w http.ResponseWriter, r *http.Request, _ models.AuthInfo) error {
	ctx := r.Context()
	lat := 48.8566
	lng := 2.3522
	apiKey := s.configuration.Google.ApiKey

	err := s.SyncGooglePlaces(ctx, lat, lng, apiKey)
	if err != nil {
		log.Println(err)
		return httpx.WriteError(w, http.StatusInternalServerError, "erreur lors de la synchro des terrains")
	}
	return httpx.Write(w, http.StatusCreated, nil)
}
