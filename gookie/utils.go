package gookie

import "time"

func expiresNtTimeEpochToTime(value int64) time.Time {
	expirationTime := time.Unix(value, 0)
	return expirationTime
}

func millisecondsToTime(ms int64) time.Time {
	// Convert milliseconds to seconds (since Unix epoch is in seconds)
	seconds := ms / 1000

	// Convert the milliseconds remainder to nanoseconds
	nanoseconds := (ms % 1000) * int64(time.Millisecond)

	// Create a time.Time value using time.Unix() and nanoseconds
	expirationTime := time.Unix(seconds, nanoseconds)

	return expirationTime
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
