package echoswg

import "os"

var HttpTraceEnabled = false

func init() {
	if "on" == os.Getenv("ECHOSWG_HTTP_TRACE") {
		HttpTraceEnabled = true
	}
}
