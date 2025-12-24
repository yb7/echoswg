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

// AliyunBackendConfig holds Aliyun API Gateway backend configuration
type AliyunBackendConfig struct {
	ServiceType    string // HTTP, FUNCTION, MOCK, etc.
	ServiceAddress string // Backend service address
	ServicePath    string // Backend service path
	ServiceMethod  string // Backend service HTTP method
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
	AliyunBackend            *AliyunBackendConfig
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
	
	// Add Aliyun API Gateway extensions if enabled
	if AliyunApiGatewayEnabled {
		// Add basic request config
		methodDef["x-aliyun-apigateway-request-config"] = map[string]interface{}{
			"requestProtocol": "HTTP",
			"requestHttpMethod": pathDefine.Method,
			"requestPath": resultPath,
			"requestMode": "PASSTHROUGH",
		}
		
		// Add backend configuration if provided
		if pathDefine.AliyunBackend != nil {
			serviceType := pathDefine.AliyunBackend.ServiceType
			if serviceType == "" {
				serviceType = "HTTP"
			}
			backendConfig := map[string]interface{}{
				"serviceType": serviceType,
			}
			if pathDefine.AliyunBackend.ServiceAddress != "" {
				backendConfig["serviceAddress"] = pathDefine.AliyunBackend.ServiceAddress
			}
			if pathDefine.AliyunBackend.ServicePath != "" {
				backendConfig["servicePath"] = pathDefine.AliyunBackend.ServicePath
			}
			if pathDefine.AliyunBackend.ServiceMethod != "" {
				backendConfig["serviceMethod"] = pathDefine.AliyunBackend.ServiceMethod
			} else {
				backendConfig["serviceMethod"] = pathDefine.Method
			}
			methodDef["x-aliyun-apigateway-backend"] = backendConfig
		} else {
			// Default backend configuration
			methodDef["x-aliyun-apigateway-backend"] = map[string]interface{}{
				"serviceType": "HTTP",
				"serviceMethod": pathDefine.Method,
			}
		}
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
