package echoswg

import (
	"net/http"
  "strings"

  "github.com/labstack/echo/v4"
)

func GenApiDoc(title, description, version string) func(echo.Context) error {
  return func(c echo.Context) error {
    var tags []map[string]string
    for tag, desc := range SwaggerTags {
      tags = append(tags, map[string]string{
        "name":        tag,
        "description": desc,
      })
    }

    docVersion := strings.TrimSpace(version)
    if len(docVersion) == 0 {
      docVersion = "0.0.0"
    }
    return c.JSON(http.StatusOK, map[string]interface{}{
      "basePath": "/",
      "host":     c.Request().Host,
      "openapi":  "3.0.0",
      "info": map[string]interface{}{
        "title":       title,
        "description": description,
        "version":     docVersion,
      },
      "paths":       SwaggerPaths,
      "definitions": GlobalTypeDefBuilder.StructDefinitions,
      "tags":        tags,
      "security": []map[string]interface{}{
        {
          "BearerAuth": []string{},
        },
      },

      "components": map[string]any{
        "securitySchemes": map[string]any{
          "BearerAuth": map[string]any{
            "type":   "http",
            "scheme": "bearer",
          },
        },
      },
    })
  }
}
