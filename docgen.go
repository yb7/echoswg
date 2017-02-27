package echoswg

import (
	"net/http"

	"github.com/labstack/echo"
)


func GenApiDoc(c echo.Context) error {
		var tags []map[string]string
		for tag, desc := range SwaggerTags {
			tags = append(tags, map[string]string{
				"name":        tag,
				"description": desc,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"basePath": "/",
			"host":     c.Request().Host,
			"swagger":  "2.0",
			"info": map[string]interface{}{
				"title":          "Swagger Sample App",
				"description":    "This is a sample server Petstore server.",
				"termsOfService": "http://swagger.io/terms/",
				"contact": map[string]string{
					"name":  "API Support",
					"url":   "http://www.swagger.io/support",
					"email": "support@swagger.io",
				},
				"license": map[string]string{
					"name": "Apache 2.0",
					"url":  "http://www.apache.org/licenses/LICENSE-2.0.html",
				},
				"version": "1.0.1",
			},
			"paths":       SwaggerPaths,
			"definitions": SwaggerDefinitions,
			"tags":        tags,
		})
}
