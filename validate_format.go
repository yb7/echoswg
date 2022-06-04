package echoswg

import (
  "github.com/labstack/gommon/log"
  "strconv"
  "strings"
)

func parseValidateTag(fieldType, validateTag string) map[string]any {
  result := make(map[string]any)
  if len(validateTag) == 0 {
    return result
  }
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
    switch name {
    case "gte":
      handleMin(result, fieldType, value, false)
      break
    case "gt":
      handleMin(result, fieldType, value, true)
      break
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
    break
  case "number":
    result["minimum"] = intValue
    if exclusive {
      result["exclusiveMinimum"] = true
    }
    break
  }
}
