package awsmocker

import (
	"log"
	_ "unsafe"
)

// GO actually caches proxy env vars which totally breaks our test flow
// so this hacks in a call to Go's internal method... This is pretty janky

//BROKENgo:linkname resetProxyConfig net/http.resetProxyConfig
// func resetProxyConfig()

func resetProxyConfig() {
	log.Print("RESET_PROXY_CONFIG CALLED!")
}

// Force call it just to make sure it works
// if Go updates this, this will make it very obvious
// func init() {
// 	resetProxyConfig()
// }
