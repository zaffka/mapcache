package mapcache

import "time"

//getTimeDurationFunc helper func to get time.Duration is seconds
func getTimeDurationFunc(intervalInSecond int) time.Duration {
	return time.Second * time.Duration(intervalInSecond)
}
