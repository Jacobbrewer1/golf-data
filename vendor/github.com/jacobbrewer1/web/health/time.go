package health

import "time"

// timestamp is a variable that holds a function returning the current UTC time.
//
// This function is designed to return the current time in UTC format.
// It is implemented as a variable to allow for mocking in tests, enabling
// controlled and predictable behavior during testing scenarios.
var timestamp = func() time.Time {
	return time.Now().UTC()
}
