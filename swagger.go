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
	Tag                      string
	Method                   string
	Summary                  string
	Description              string
	OperationId              string
	Path                     string
	InternalHttpTraceEnabled bool
	Handlers                 []interface{}
}

// MountSwaggerPath func
func MountSwaggerPath(pathDefine *SwaggerPathDefine) {
	fmt.Printf("%-8s%s\n", pathDefine.Method, pathDefine.Path)
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

	resultPath := pathDefine.Path
	for _, pname := range ParsePathNames(pathDefine.Path) {
		replaceTo := pname
		// 替换成json tag的名字
		for _, pathParam := range requestParam.PathParams {
			if pname == pathParam.Name && len(pathParam.JsonFieldName) > 0 {
				replaceTo = pathParam.JsonFieldName
				//jsonName := strings.SplitN(pathParam.Tag.Get("json"), ",", 2)[0]
				//if len(jsonName) > 0 {
				//	name = jsonName
				//}
			}
		}
		resultPath = strings.Replace(resultPath, ":"+pname, "{"+replaceTo+"}", -1)
	}

	operationId := pathDefine.OperationId
	if len(operationId) == 0 {
		operationId = getOperationID(pathDefine.Tag, pathDefine.Handlers)
	}
	methodDef := map[string]interface{}{
		"tags":        []string{pathDefine.Tag},
		"summary":     pathDefine.Summary,
		"description": pathDefine.Description,
		//"produces":    []string{"application/json"},
		//"consumes":    []string{"application/json"},
		"operationId": operationId,
		"parameters":  requestParam.ParametersToSwaggerJSON(),
		"responses": map[string]interface{}{
			"200": successResponse,
			"500": map[string]interface{}{
				"description": "Interal Server Error",
			},
		},
	}
	if requestParam.RequestBody != nil {
		methodDef["requestBody"] = requestParam.RequestBodyToSwaggerJSON()
	}
	json := map[string]interface{}{
		strings.ToLower(pathDefine.Method): methodDef,
	}

	return &SwaggerPath{Path: resultPath, JSON: json}
}

func getRootOfPtr(typ reflect.Type) reflect.Type {
	if typ.Kind() == reflect.Ptr {
		return getRootOfPtr(typ.Elem())
	}
	return typ
}
