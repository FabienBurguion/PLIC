package main

import "time"

type Clock struct {
	location *time.Location
}

func (c Clock) Now() time.Time {
	return time.Now().In(c.location)
}
