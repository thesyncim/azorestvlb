package main

import ()

type Maxmindrecord struct {
	Continent          Continent
	Country            Country
	Registered_country Country
}

const DefaultContinent = "EU"

type Country struct {
	Geoname_id int64
	Iso_code   string
	Names      map[string]string
}

type Continent struct {
	Code       string
	Geoname_id int64
	Names      map[string]string
}
