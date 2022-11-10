package echoswg

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// RequestParam struct
type RequestParam struct {
	PathParams     []Param
	QueryParams    []Param
	RequestBody    reflect.Type
	RequestBodyTag reflect.StructTag
}

// Param struct
type Param struct {
	Name        string
	Type        reflect.Type
	Tag         reflect.StructTag
	Required    bool
	Description string
}

func (p *Param) String() string {
	return fmt.Sprintf("name[%s] type:%s, required:%t\n", p.Name, p.Type, p.Required)
}

// ToSwaggerJSON func
func (p *Param) ToSwaggerJSON(position string) map[string]interface{} {
	//typ, format := GoTypeToSwaggerType(p.Type)
	t := GlobalTypeDefBuilder.ToSwaggerType(p.Type, p.Tag)

	//https://swagger.io/docs/specification/data-models/data-types/

	return map[string]interface{}{
		"name": p.Name,
		"in":   position,
		"schema": map[string]string{
			"type":   t.Type,
			"format": t.Format,
		},
		"required":    p.Required,
		"description": p.Description,
	}
}

func addPathAndQueryParams(path string, inType reflect.Type, pathParams *[]Param, queryParams *[]Param) {
	requestType := inType
	if requestType.Kind() == reflect.Ptr {
		requestType = requestType.Elem()
	}
	if requestType.Name() == "Context" {
		return
	}
	if requestType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("request type [%v] must be Struct, but is %v\n", requestType.Name(), requestType.Kind()))
	}
	pnames := ParsePathNames(path)
	for i := 0; i < requestType.NumField(); i++ {
		typeField := requestType.Field(i)

		if strings.ToUpper(typeField.Name) != "BODY" {
			if typeField.Anonymous {
				addPathAndQueryParams(path, typeField.Type, pathParams, queryParams)
			} else {
				param := Param{Name: typeField.Name, Type: typeField.Type, Required: typeField.Type.Kind() != reflect.Ptr,
					Tag: typeField.Tag, Description: typeField.Tag.Get("desc")}

				if !param.Required {
					param.Type = param.Type.Elem()
				}
				if containsIgnoreCase(pnames, typeField.Name) {
					appendToSet(pathParams, &param)
					// fmt.Printf("\tPath Params %s", param.String())
				} else {
					appendToSet(queryParams, &param)
					// fmt.Printf("\tQuery Params %s", param.String())
				}
			}
		}

	}
}
func findRequestBody(inTypes []reflect.Type) (reflect.Type, reflect.StructTag) {
	var requestBody reflect.Type
	var requestBodyTag reflect.StructTag
	for _, inType := range inTypes {
		requestType := inType
		if requestType.Kind() == reflect.Ptr {
			requestType = requestType.Elem()
		}
		if requestType.Name() == "Context" {
			continue
		}
		for i := 0; i < requestType.NumField(); i++ {
			typeField := requestType.Field(i)

			if strings.ToUpper(typeField.Name) == "BODY" {
				requestBodyTag = typeField.Tag
				if requestBody != nil {
					fmt.Sprintf("only last body will be show. %v ignored!\n", requestBody)
				}
				if typeField.Type.Kind() == reflect.Ptr {
					requestBody = typeField.Type.Elem()
				} else {
					requestBody = typeField.Type
				}
			}
		}
	}
	return requestBody, requestBodyTag
}
func appendToSet(set *[]Param, newOne *Param) {
	for _, e := range *set {
		if e.Name == newOne.Name {
			return
		}
	}
	*set = append(*set, *newOne)
}

// BuildRequestParam func
func BuildRequestParam(path string, inTypes []reflect.Type) *RequestParam {
	if len(inTypes) == 0 {
		return &RequestParam{}
	}
	var pathParams []Param
	var queryParams []Param
	for _, inType := range inTypes {
		addPathAndQueryParams(path, inType, &pathParams, &queryParams)
	}

	// TODO: 可以选择性的打印参数
	// printParams(pathParams, queryParams)

	requestBody, requestBodyTag := findRequestBody(inTypes)
	return &RequestParam{PathParams: pathParams, QueryParams: queryParams, RequestBody: requestBody, RequestBodyTag: requestBodyTag}
}

func printParams(pathParams []Param, queryParams []Param) {
	if len(pathParams)+len(queryParams) == 0 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	// table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"PATH", "type", "required", "QUERY", "type", "required"})
	for i := 0; i < len(pathParams) || i < len(queryParams); i++ {
		var data []string
		if i >= len(pathParams) {
			data = append(data, "", "", "")
		} else {
			p := pathParams[i]
			data = append(data, p.Name, p.Type.String(), strconv.FormatBool(p.Required))
		}
		if i >= len(queryParams) {
			data = append(data, "", "", "")
		} else {
			p := queryParams[i]
			data = append(data, p.Name, p.Type.String(), strconv.FormatBool(p.Required))
		}
		table.Append(data)
	}
	table.Render() // Send output
}

// ToSwaggerJSON func
func (req *RequestParam) ParametersToSwaggerJSON() []map[string]interface{} {
	var parameters = make([]map[string]interface{}, 0)
	for _, pathParam := range req.PathParams {
		parameters = append(parameters, pathParam.ToSwaggerJSON("path"))
	}
	for _, queryParam := range req.QueryParams {
		parameters = append(parameters, queryParam.ToSwaggerJSON("query"))
	}
	return parameters
}

func (req *RequestParam) RequestBodyToSwaggerJSON() map[string]interface{} {
	swaggerType := GlobalTypeDefBuilder.Build(req.RequestBody, req.RequestBodyTag)

	return map[string]interface{}{
		"description": req.RequestBodyTag.Get("desc"),
		"required":    true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": swaggerType.ToSwaggerJSON(),
			},
		},
	}
}
