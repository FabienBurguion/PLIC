package models

import (
	"time"

	"github.com/google/uuid"
)

type DBCourt struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Address   string    `db:"address"`
	Longitude float64   `db:"longitude"`
	Latitude  float64   `db:"latitude"`
	CreatedAt time.Time `db:"created_at"`
}

func NewDBCourtFixture() DBCourt {
	return DBCourt{
		Id:        uuid.NewString(),
		Name:      "a court",
		Address:   "an address",
		Longitude: 0.0,
		Latitude:  0.0,
		CreatedAt: time.Now(),
	}
}

func (u DBCourt) WithId(id string) DBCourt {
	u.Id = id
	return u
}

func (u DBCourt) WithName(name string) DBCourt {
	u.Name = name
	return u
}

func (u DBCourt) WithAddress(address string) DBCourt {
	u.Address = address
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
