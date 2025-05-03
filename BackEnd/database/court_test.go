package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDatabase_InsertTerrain(t *testing.T) {
	type testCase struct {
		name  string
		param string
	}

	id := uuid.NewString()
	testCases := []testCase{
		{
			name:  "Basic test",
			param: id,
		},
	}

	testLat := 48.8566
	testLng := 2.3522

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()
			ctx := context.Background()

			err := s.db.InsertTerrain(ctx, id, models.Place{
				Name:    "Test Court",
				Address: "123 Rue Test",
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
						Lat: testLat,
						Lng: testLng,
					},
				},
			})
			require.NoError(t, err)
			court, err := s.db.GetTerrainByAddress(ctx, "123 Rue Test")
			require.NoError(t, err)
			require.Equal(t, court.Id, c.param)
		})
	}
}
