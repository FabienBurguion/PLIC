package main

import (
	"PLIC/models"
	"context"
	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) GetPlaces(lat, lng float64, apiKey string) ([]models.Place, error) {
	args := m.Called(lat, lng, apiKey)
	return args.Get(0).([]models.Place), args.Error(1)
}

func (m *mockDB) InsertTerrain(ctx context.Context, id string, place models.Place) error {
	args := m.Called(ctx, id, place)
	return args.Error(0)
}

/*func TestService_HandleSyncGooglePlaces(t *testing.T) {
	mockedDB := new(mockDB)

	s := &Service{}
	s.InitServiceTest()
	s.db = mockedDB
	place := models.Place{
		Name:    "Test Place",
		Address: "123 Street",
	}
	place.Geometry.Location.Lat = 48.8566
	place.Geometry.Location.Lng = 2.3522

	mockedDB.On("GetPlaces", 48.8566, 2.3522, "fake-api-key").Return([]models.Place{place}, nil)
	mockedDB.On("InsertTerrain", mock.Anything, mock.AnythingOfType("string"), place).Return(nil)

	req := httptest.NewRequest("POST", "/place", nil)
	w := httptest.NewRecorder()
	os.Setenv("GOOGLE_APIKEY", "fake-api-key")

	err := s.HandleSyncGooglePlaces(w, req, models.AuthInfo{})
	require.NoError(t, err)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	mockedDB.AssertExpectations(t)
}*/
