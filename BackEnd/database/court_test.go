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

func TestDatabase_GetAllTerrains(t *testing.T) {
	type testCase struct {
		name     string
		expected models.DBCourt
	}

	testID := uuid.NewString()

	testCases := []testCase{
		{
			name: "Basic terrain retrieval",
			expected: models.DBCourt{
				Id:        testID,
				Address:   "123 Rue Test",
				Longitude: 2.3522,
				Latitude:  48.8566,
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := &Service{}
			s.InitServiceTest() // initialise DB test
			ctx := context.Background()

			// Insérer un terrain manuellement pour test
			_, err := s.db.Database.ExecContext(ctx, `
				INSERT INTO terrain (id, address, longitude, latitude)
				VALUES ($1, $2, $3, $4)`,
				c.expected.Id, c.expected.Address, c.expected.Longitude, c.expected.Latitude,
			)
			require.NoError(t, err)

			// Appel de la fonction à tester
			terrains, err := s.db.GetAllTerrains(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, terrains)

			var found bool
			for _, terr := range terrains {
				if terr.Id == c.expected.Id {
					require.Equal(t, c.expected.Address, terr.Address)
					require.Equal(t, c.expected.Longitude, terr.Longitude)
					require.Equal(t, c.expected.Latitude, terr.Latitude)
					found = true
					break
				}
			}
			require.True(t, found, "terrain not found in results")
		})
	}
}
