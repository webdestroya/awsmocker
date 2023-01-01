package awsmocker

import (
	"io"
	"os"
)

var (
	// Will Print out all the Request/Response traffic from the proxy
	GlobalDebugMode = false

	// where debugging output will go if requested
	DebugOutputWriter io.Writer = os.Stdout
)

const (
	envGlobalDebug = "AWSMOCKER_DEBUG"
)

func init() {
	val, ok := os.LookupEnv(envGlobalDebug)
	if ok && val != "false" {
		GlobalDebugMode = true
	}
}

func getDebugMode() bool {

	// // this is "faster", but it breaks test caching, so keep the slow version
	// if GlobalDebugMode {
	// 	return true
	// }

	val, ok := os.LookupEnv(envGlobalDebug)
	if ok && val != "false" {
		return true
	}

	return GlobalDebugMode
}
