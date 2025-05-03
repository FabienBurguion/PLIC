package models

import (
	"github.com/google/uuid"
	"time"
)

type DBCourt struct {
	Id        string    `db:"id"`
	Address   string    `db:"address"`
	Longitude float64   `db:"longitude"`
	Latitude  float64   `db:"latitude"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewDBCourtFixture() DBCourt {
	return DBCourt{
		Id:        uuid.NewString(),
		Address:   "address",
		Longitude: 0.0,
		Latitude:  0.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u DBCourt) WithId(id string) DBCourt {
	u.Id = id
	return u
}

func (u DBCourt) WithLongitude(longitude float64) DBCourt {
	u.Longitude = longitude
	return u
}

func (u DBCourt) WithLatitude(latitude float64) DBCourt {
	u.Latitude = latitude
	return u
}
