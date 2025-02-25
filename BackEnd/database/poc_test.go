package database

import (
	"PLIC/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_POCDB(t *testing.T) {
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := &Service{}
			s.InitServiceTest()

			ctx := context.Background()

			err := s.db.CreateUser(ctx, models.DBUser{
				Id:    id,
				Name:  "A name",
				Email: "a@a.com",
			})
			require.NoError(t, err)
			user, err := s.db.GetUser(ctx, id)
			require.NoError(t, err)
			require.Equal(t, user.Id, c.param)
		})
	}
}
