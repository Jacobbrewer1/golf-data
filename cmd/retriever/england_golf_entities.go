package main

import (
	"github.com/jacobbrewer1/golf-data/pkg/models"
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
