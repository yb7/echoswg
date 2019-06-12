package echoswg

import (
	"github.com/labstack/echo"
	"reflect"
)

type ApiGroup interface {
	SetDescription(desc string)
	EchoGroup() *echo.Group
	GET(url string, actions ...interface{})
	POST(url string, actions ...interface{})
	PUT(url string, actions ...interface{})
	DELETE(url string, actions ...interface{})
  Use(middleware ...echo.MiddlewareFunc)
}

type internalApiGroup struct {
	tag string
	urlPrefix string
	summary string
	description string
	echoGroup *echo.Group
}
func (g *internalApiGroup) SetDescription(desc string) {g.description = desc}
func (g *internalApiGroup) EchoGroup() *echo.Group {return g.echoGroup}
type CanGroup interface {
	Group(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group
}
func NewApiGroup(canGroup CanGroup, tag string, prefix string) ApiGroup {
	echoGroup := canGroup.Group(prefix)
	apiGroup := &internalApiGroup{
		tag: tag,
		urlPrefix: prefix,
		echoGroup: echoGroup,
	}
	return apiGroup
}

func (g *internalApiGroup) Use(middleware ...echo.MiddlewareFunc) {
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
func (g *internalApiGroup) wrapper(method string, url string, actions []interface{}) (string, echo.HandlerFunc) {
	var summary, description string
	var handlers []interface{}
	for _, a := range actions {
		if reflect.TypeOf(a).Kind() == reflect.String {
			if len(summary) == 0 {
				summary = a.(string)
			} else {
				description = a.(string)
			}
		} else {
			handlers = append(handlers, a)
		}
	}
	SwaggerTags[g.tag] = g.description
	fullPath := g.urlPrefix + url
	MountSwaggerPath(&SwaggerPathDefine{Tag: g.tag, Method: method, Path: fullPath,
		Summary: summary, Description: description, Handlers: handlers})
	echoHandler := BuildEchoHandler(fullPath, handlers)
	return url, echoHandler
}
