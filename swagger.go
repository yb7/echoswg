package echoswg

import (
	"fmt"
	"reflect"
	"strings"
)

// SwaggerPaths cache
var SwaggerPaths = make(map[string]interface{})

// SwaggerTags cache
var SwaggerTags = make(map[string]string)

// SwaggerPath struct
type SwaggerPath struct {
	Path string
	JSON map[string]interface{}
}

// SwaggerPathDefine struct
type SwaggerPathDefine struct {
	Tag         string
	Method      string
	Summary     string
	Description string
	Path        string
	Handlers    []interface{}
}

// MountSwaggerPath func
func MountSwaggerPath(pathDefine *SwaggerPathDefine) {
	fmt.Printf("\n\n%-8s%s\n", pathDefine.Method, pathDefine.Path)
	newPath := BuildSwaggerPath(pathDefine)

	if exist, ok := SwaggerPaths[newPath.Path]; !ok {
		SwaggerPaths[newPath.Path] = newPath.JSON
	} else {
		for k, v := range newPath.JSON {
			exist.(map[string]interface{})[k] = v
		}
	}
}

// BuildSwaggerPath func
func BuildSwaggerPath(pathDefine *SwaggerPathDefine) *SwaggerPath {
	resultPath := pathDefine.Path
	for _, pname := range ParsePathNames(pathDefine.Path) {
		resultPath = strings.Replace(resultPath, ":"+pname, "{"+pname+"}", -1)
	}

	inTypes, outType, err := validateChain(pathDefine.Handlers)

	if err != nil {
		panic(err)
	}

	successResponse := map[string]interface{}{
		"description": "successful operation",
	}
	if outType != nil {
		swaggerType := GlobalTypeDefBuilder.Build(outType, "")
		successResponse = map[string]interface{}{
      "description": "successful operation",
      "content": map[string]any{
        "application/json": map[string]any{
          "schema": swaggerType.ToSwaggerJSON(), //SwaggerEntitySchemaRef(outType),
        },
      },
    }
	}
	requestParam := BuildRequestParam(pathDefine.Path, inTypes)
	json := map[string]interface{}{
    strings.ToLower(pathDefine.Method): map[string]interface{}{
      "tags":        []string{pathDefine.Tag},
      "summary":     pathDefine.Summary,
      "description": pathDefine.Description,
      //"produces":    []string{"application/json"},
      //"consumes":    []string{"application/json"},
      "operationId": getOperationID(pathDefine.Handlers),
      "parameters":  requestParam.ToSwaggerJSON(),
      "responses": map[string]interface{}{
        "200": successResponse,
        "500": map[string]interface{}{
          "description": "Interal Server Error",
        },
      },
    },
  }
	return &SwaggerPath{Path: resultPath, JSON: json}
}

func getRootOfPtr(typ reflect.Type) reflect.Type {
  if typ.Kind() == reflect.Ptr {
    return getRootOfPtr(typ.Elem())
  }
  return typ
}
