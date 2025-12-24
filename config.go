package echoswg

import "os"

var HttpTraceEnabled = false
var AliyunApiGatewayEnabled = false

func init() {
	if "on" == os.Getenv("ECHOSWG_HTTP_TRACE") {
		HttpTraceEnabled = true
	}
	if "on" == os.Getenv("ECHOSWG_ALIYUN_GATEWAY") {
		AliyunApiGatewayEnabled = true
	}
}
