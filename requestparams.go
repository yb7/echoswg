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
	PathParams  []Param
	QueryParams []Param
	RequestBody reflect.Type
}

// Param struct
type Param struct {
	Name        string
	Type        reflect.Type
	Description string
	Required    bool
}

func (p *Param) String() string {
	return fmt.Sprintf("name[%s] type:%s, required:%t\n", p.Name, p.Type, p.Required)
}

// ToSwaggerJSON func
func (p *Param) ToSwaggerJSON(position string) map[string]interface{} {
	typ, format := GoTypeToSwaggerType(p.Type)
	return map[string]interface{}{
		"name":        lowCamelStr(p.Name),
		"in":          position,
		"format":      format,
		"required":    p.Required,
		"type":        typ,
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
			param := Param{Name: typeField.Name, Type: typeField.Type, Required: typeField.Type.Kind() != reflect.Ptr}
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
func findRequestBody(inTypes []reflect.Type) reflect.Type {
	var requestBody reflect.Type
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
				if requestBody != nil {
					panic("only one request parameter can have `Body`")
				}
				if typeField.Type.Kind() == reflect.Ptr {
					requestBody = typeField.Type.Elem()
				} else {
					requestBody = typeField.Type
				}
			}
		}
	}
	return requestBody
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
	printParams(pathParams, queryParams)

	requestBody := findRequestBody(inTypes)
	return &RequestParam{PathParams: pathParams, QueryParams: queryParams, RequestBody: requestBody}
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
func (req *RequestParam) ToSwaggerJSON() []map[string]interface{} {
	var parameters []map[string]interface{}
	for _, pathParam := range req.PathParams {
		parameters = append(parameters, pathParam.ToSwaggerJSON("path"))
	}
	for _, queryParam := range req.QueryParams {
		parameters = append(parameters, queryParam.ToSwaggerJSON("query"))
	}
	if req.RequestBody != nil {
		parameters = append(parameters, map[string]interface{}{
			"in":       "body",
			"name":     "body",
			"required": true,
			"schema":   SwaggerEntitySchemaRef(req.RequestBody),
			// map[string]string{
			// 	"$ref": "#/definitions/" + req.RequestBody.Name(),
			// },
		})
	}
	return parameters
}
