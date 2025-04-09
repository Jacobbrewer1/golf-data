package main

import (
	"strconv"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/goschema/usql"
)

type EnglandGolfClubResponse struct {
	ClubId      int     `json:"ClubId"`
	ClubName    string  `json:"ClubName"`
	LocAddress1 string  `json:"LocAddress1"`
	LocAddress2 string  `json:"LocAddress2"`
	LocAddress3 string  `json:"LocAddress3"`
	LocAddress4 string  `json:"LocAddress4"`
	PostalCode  string  `json:"PostalCode"`
	Email       *string `json:"Email"`
	Phone       string  `json:"Phone"`
}

func (c *EnglandGolfClubResponse) ToModel() *models.Club {
	dbClub := &models.Club{
		Id:         c.ClubId,
		Name:       c.ClubName,
		Address1:   c.LocAddress1,
		Address2:   c.LocAddress2,
		Address3:   c.LocAddress3,
		Address4:   c.LocAddress4,
		PostalCode: c.PostalCode,
	}
	return dbClub
}

type EnglandGolfCourseResponse struct {
	CourseId int    `json:"CourseId"`
	Name     string `json:"Name"`
}

func (c *EnglandGolfCourseResponse) ToModel(clubId int) *models.Course {
	dbCourse := &models.Course{
		Id:     c.CourseId,
		Name:   c.Name,
		ClubId: clubId,
	}
	return dbCourse
}

type EnglandGolfCourseDetailsResponse struct {
	MarkerId                int                        `json:"MarkerId"`
	DisplayMarkerName       string                     `json:"DisplayMarkerName"`
	CourseRating            float64                    `json:"UsgaNzcr"`
	SlopeRating             int                        `json:"SlopeRating"`
	MarkerColor             string                     `json:"MarkerColor"`
	Holes                   []*EnglandGolfResponseHole `json:"Holes"`
	FrontNineDistanceMetres int                        `json:"FrontNineDistanceMetres"`
	FrontNineDistanceYards  int                        `json:"FrontNineDistanceYards"`
	FrontNinePar            int                        `json:"FrontNinePar"`
	BackNineDistanceMetres  int                        `json:"BackNineDistanceMetres"`
	BackNineDistanceYards   int                        `json:"BackNineDistanceYards"`
	BackNinePar             int                        `json:"BackNinePar"`
	TotalPar                int                        `json:"TotalPar"`
}

func (c *EnglandGolfCourseDetailsResponse) ToModel(courseId int) *models.CourseDetails {
	dbCourseDetails := &models.CourseDetails{
		Id:              c.MarkerId,
		CourseId:        courseId,
		Marker:          *usql.NewNullString(c.DisplayMarkerName),
		Slope:           c.SlopeRating,
		CourseRating:    c.CourseRating,
		FrontNinePar:    c.FrontNinePar,
		BackNinePar:     c.BackNinePar,
		TotalPar:        c.TotalPar,
		FrontNineYards:  c.FrontNineDistanceYards,
		BackNineYards:   c.BackNineDistanceYards,
		TotalYards:      c.FrontNineDistanceYards + c.BackNineDistanceYards,
		FrontNineMeters: c.FrontNineDistanceMetres,
		BackNineMeters:  c.BackNineDistanceMetres,
		TotalMeters:     c.FrontNineDistanceMetres + c.BackNineDistanceMetres,
	}
	return dbCourseDetails
}

type EnglandGolfResponseHole struct {
	HoleNumStr     string `json:"Alias"`
	Par            int    `json:"Par"`
	Stroke         int    `json:"Stroke"`
	DistanceMetres int    `json:"DistanceMetres"`
	DistanceYards  int    `json:"DistanceYards"`
}

func (h *EnglandGolfResponseHole) ToModel(markerId int) *models.Hole {
	holeNum, err := strconv.Atoi(h.HoleNumStr)
	if err != nil {
		log.Printf("Error converting HoleNumStr to integer: %v", err)
		holeNum = 0 // or handle the error as appropriate
	}
	dbHole := &models.Hole{
		DetailsId:      markerId,
		Number:         holeNum,
		Par:            h.Par,
		Stroke:         h.Stroke,
		DistanceYards:  h.DistanceYards,
		DistanceMeters: h.DistanceMetres,
	}
	return dbHole
}
