package echoswg

import "os"

var HttpTraceEnabled = true
func init() {
  if "off" == os.Getenv("ECHOSWG_HTTP_TRACE") {
    HttpTraceEnabled = false
  }
}
