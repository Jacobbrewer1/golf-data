package filters

import "github.com/jacobbrewer1/pagefilter"

type markerCourseID struct {
	courseID int64
}

// NewMarkerCourseID creates a new marker for course details ID.
func NewMarkerCourseID(courseID int64) pagefilter.Wherer {
	return &markerCourseID{
		courseID: courseID,
	}
}

func (m *markerCourseID) Where() (string, []any) {
	return "t.course_id = ?", []any{m.courseID}
}
