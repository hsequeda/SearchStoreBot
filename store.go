package main

type Location struct {
	Latitude  float64
	Longitude float64
}

type Store struct {
	ID           int64
	Municipality string
	Name         string
	Address      string
	Department   string
	Geolocation  Location
	MapUrl       string
	Phone        string
	Open         string
	Close        string
}
