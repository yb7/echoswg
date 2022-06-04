package echoswg

import (
	"net/http"
  "strings"

  "github.com/labstack/echo/v4"
)

// GenApiDoc generate v3 api https://swagger.io/docs/specification/basic-structure/
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
      "servers": []map[string]string{
        {"url": "/"},
      },
      "host":    c.Request().Host,
      "openapi": "3.0.0",
      "info": map[string]interface{}{
        "title":       title,
        "description": description,
        "version":     docVersion,
      },
      "paths": SwaggerPaths,

      "tags": tags,
      "security": []map[string]interface{}{
        {
          "BearerAuth": []string{},
        },
      },

      "components": map[string]any{
        "schemas": GlobalTypeDefBuilder.StructDefinitions,
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
