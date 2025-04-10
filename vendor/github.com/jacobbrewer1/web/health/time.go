package health

import "time"

// timestamp is a function that returns the current UTC time.
// This is purposefully a method to allow for mocking in tests.
var timestamp = func() time.Time {
	return time.Now().UTC()
}
