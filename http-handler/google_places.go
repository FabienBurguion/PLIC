package main

import (
	googlehandler "PLIC/google-handler"
	"PLIC/httpx"
	"PLIC/models"
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (s *Service) SyncGooglePlaces(ctx context.Context, latitude, longitude float64, apiKey string) error {
	baseLogger := log.With().
		Str("method", "SyncGooglePlaces").
		Float64("latitude", latitude).
		Float64("longitude", longitude).
		Int("api_key_len", len(apiKey)).
		Logger()

	baseLogger.Info().Msg("starting Google Places sync")

	places, err := googlehandler.GetPlaces("https://maps.googleapis.com/maps/api/place/nearbysearch", latitude, longitude, apiKey)
	if err != nil {
		baseLogger.Error().Err(err).Msg("failed to fetch places from Google Maps API")
		return fmt.Errorf("échec de la récupération des lieux via Google Maps API : %w", err)
	}

	for _, place := range places {
		id := uuid.NewString()
		if err := s.db.InsertCourt(ctx, id, place, s.clock.Now()); err != nil {
			baseLogger.Error().
				Err(err).
				Str("place_name", place.Name).
				Msg("failed to insert court in database")
		} else {
			baseLogger.Info().
				Str("place_name", place.Name).
				Msg("court inserted successfully")
		}
	}

	baseLogger.Info().Int("places_count", len(places)).Msg("Google Places sync completed")
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
	baseLogger := log.With().
		Str("method", "HandleSyncGooglePlaces").
		Logger()

	logger := baseLogger.With().
		Str("remote_addr", r.RemoteAddr).
		Str("path", r.URL.Path).
		Logger()

	logger.Info().Msg("entering HandleSyncGooglePlaces")

	ctx := r.Context()
	lat := 48.8566
	lng := 2.3522
	apiKey := s.configuration.Google.ApiKey

	if err := s.SyncGooglePlaces(ctx, lat, lng, apiKey); err != nil {
		logger.Error().Err(err).Msg("Google Places sync failed")
		return httpx.WriteError(w, http.StatusInternalServerError, "erreur lors de la synchro des terrains")
	}

	logger.Info().Msg("Google Places sync succeeded")
	return httpx.Write(w, http.StatusCreated, nil)
}
