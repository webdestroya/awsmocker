package awsmocker

import "sync/atomic"

var requestCounter = &atomic.Uint64{}

func getRequestNum() uint64 {
	return requestCounter.Add(1)
}
