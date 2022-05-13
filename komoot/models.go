package komoot

import "time"

type ToursResponse struct {
	Embedded Embedded `json:"_embedded"`
	Links    Links    `json:"_links"`
	Page     Page     `json:"page"`
}
type Creator struct {
	Href string `json:"href"`
}
type Self struct {
	Href string `json:"href"`
}
type Coordinates struct {
	Href string `json:"href"`
}
type Participants struct {
	Href string `json:"href"`
}
type Timeline struct {
	Href string `json:"href"`
}
type Translations struct {
	Href string `json:"href"`
}
type CoverImages struct {
	Href string `json:"href"`
}

type StartPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}
type Avatar struct {
	Src       string `json:"src"`
	Templated bool   `json:"templated"`
	Type      string `json:"type"`
}
type Relation struct {
	Href      string `json:"href"`
	Templated bool   `json:"templated"`
}
type Links struct {
	Self     Self     `json:"self"`
	Relation Relation `json:"relation"`
}

type MapImage struct {
	Src         string `json:"src"`
	Templated   bool   `json:"templated"`
	Type        string `json:"type"`
	Attribution string `json:"attribution"`
}
type MapImagePreview struct {
	Src         string `json:"src"`
	Templated   bool   `json:"templated"`
	Type        string `json:"type"`
	Attribution string `json:"attribution"`
}
type Activity struct {
	Status          string          `json:"status"`
	Type            string          `json:"type"`
	DateString      string          `json:"date"`
	Name            string          `json:"name"`
	Source          string          `json:"source"`
	Distance        float64         `json:"distance"`
	Duration        int             `json:"duration"`
	Sport           string          `json:"sport"`
	Links           Links           `json:"_links"`
	KcalActive      int             `json:"kcal_active"`
	KcalResting     int             `json:"kcal_resting"`
	StartPoint      StartPoint      `json:"start_point"`
	ElevationUp     float64         `json:"elevation_up"`
	ElevationDown   float64         `json:"elevation_down"`
	TimeInMotion    int             `json:"time_in_motion"`
	Embedded        Embedded        `json:"_embedded"`
	ID              int             `json:"id"`
	ChangedAt       time.Time       `json:"changed_at"`
	MapImage        MapImage        `json:"map_image"`
	MapImagePreview MapImagePreview `json:"map_image_preview"`
	Date            time.Time
	Private         bool
}
type Embedded struct {
	Tours []Activity `json:"tours"`
}

type Page struct {
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Number        int `json:"number"`
}
