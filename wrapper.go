package echoswg

import (
	"reflect"

	"github.com/labstack/echo/v4"
)

type ApiGroup interface {
	SetDescription(desc string)
	EchoGroup() *echo.Group
	GET(url string, actions ...interface{})
	POST(url string, actions ...interface{})
	PUT(url string, actions ...interface{})
	DELETE(url string, actions ...interface{})
	Any(url string, actions ...interface{})
	USE(middleware ...echo.MiddlewareFunc)
}

type internalApiGroup struct {
	tag         string
	urlPrefix   string
	summary     string
	description string
	echoGroup   *echo.Group
}

func (g *internalApiGroup) SetDescription(desc string) { g.description = desc }
func (g *internalApiGroup) EchoGroup() *echo.Group     { return g.echoGroup }

type CanGroup interface {
	Group(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group
}

func NewApiGroup(canGroup CanGroup, tag string, prefix string) ApiGroup {
	echoGroup := canGroup.Group(prefix)
	apiGroup := &internalApiGroup{
		tag:       tag,
		urlPrefix: prefix,
		echoGroup: echoGroup,
	}
	return apiGroup
}

func (g *internalApiGroup) USE(middleware ...echo.MiddlewareFunc) {
	g.echoGroup.Use(middleware...)
}

func (g *internalApiGroup) GET(url string, actions ...interface{}) {
	g.echoGroup.GET(g.wrapper("GET", url, actions))
}
func (g *internalApiGroup) POST(url string, actions ...interface{}) {
	g.echoGroup.POST(g.wrapper("POST", url, actions))
}
func (g *internalApiGroup) PUT(url string, actions ...interface{}) {
	g.echoGroup.PUT(g.wrapper("PUT", url, actions))
}
func (g *internalApiGroup) DELETE(url string, actions ...interface{}) {
	g.echoGroup.DELETE(g.wrapper("DELETE", url, actions))
}
func (g *internalApiGroup) Any(url string, actions ...interface{}) {
	g.echoGroup.Any(g.wrapper("Any", url, actions))
}
func (g *internalApiGroup) wrapper(method string, url string, actions []interface{}) (string, echo.HandlerFunc) {
	var summary, description string
	var handlers []interface{}
	internalHttpTraceEnabled := HttpTraceEnabled
	for _, a := range actions {
		if reflect.TypeOf(a).Kind() == reflect.String {
			strValue := a.(string)
			if strValue == "__LOG_OFF" {
				internalHttpTraceEnabled = false
			} else if strValue == "__LOG_ON" {
				internalHttpTraceEnabled = true
			} else if len(summary) == 0 {
				summary = a.(string)
			} else {
				description = a.(string)
			}
		} else {
			handlers = append(handlers, a)
		}
	}
	// TODO: 此处直接关闭，以后优化
	internalHttpTraceEnabled = false

	SwaggerTags[g.tag] = g.description
	fullPath := g.urlPrefix + url
	MountSwaggerPath(&SwaggerPathDefine{Tag: g.tag, Method: method, Path: fullPath,
		Summary: summary, Description: description, Handlers: handlers})
	echoHandler := BuildEchoHandler(fullPath, HandlerConfig{
		DisableLog: !internalHttpTraceEnabled,
	}, handlers)
	return url, echoHandler
}
