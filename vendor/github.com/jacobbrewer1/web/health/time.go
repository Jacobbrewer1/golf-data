package health

import "time"

var Timestamp = func() time.Time {
	return time.Now().UTC()
}
