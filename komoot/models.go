package komoot

import "time"

type ToursResponse struct {
	Embedded Embedded `json:"_embedded"`
	Links    Link     `json:"_links"`
	Page     Page
}

type Page struct {
	Size          int
	TotalElements int
	TotalPages    int
	Number        int
}
type Link struct {
	Self Href
	Next Href
}

type Href struct {
	Href string
}

type Embedded struct {
	Tours []Activity
}

type Activity struct {
	Id           int
	Status       string
	Type         string
	DateString   string `json:"Date"`
	Date         time.Time
	Name         string
	Source       string
	Distance     float64
	Duration     int
	Sport        string
	Query        string
	Constitution int
	Private      bool
}
