package echoswg

import (
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
)

func validateTagToMap(validateTag string) map[string]string {
	result := make(map[string]string)
	tags := strings.Split(validateTag, ",")
	for _, _tag := range tags {
		tag := strings.TrimSpace(_tag)
		eqIdx := strings.Index(tag, "=")
		var name, value string
		if eqIdx == -1 {
			name = tag
		} else {
			name = tag[:eqIdx]
			value = tag[eqIdx+1:]
		}
		result[name] = value
	}
	return result
}
func parseValidateTag(fieldType, validateTag string) map[string]any {
	result := make(map[string]any)
	if len(validateTag) == 0 {
		return result
	}
	tags := validateTagToMap(validateTag)
	for name, value := range tags {
		switch name {
		case "gte":
			handleMin(result, fieldType, value, false)
		case "gt":
			handleMin(result, fieldType, value, true)
		}
	}
	return result
}

func handleMin(result map[string]any, fieldType, value string, exclusive bool) {
	intValue, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		log.Warnf("echoswg.handleMin: %s", err.Error())
		return
	}
	switch fieldType {
	case "array":
		result["minItems"] = intValue
	case "number":
		result["minimum"] = intValue
		if exclusive {
			result["exclusiveMinimum"] = true
		}
	}
}
