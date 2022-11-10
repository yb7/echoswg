package echoswg

import (
  "embed"
  "fmt"
  "html/template"
  "strings"

  "github.com/labstack/echo/v4"
)

//go:embed swagger-ui-4.10.3
var SwaggerUiFS embed.FS

type SwaggerConfig struct {
  UrlPrefix   string
  Title       string
  Description string
  ApiDocUrl   string
  CdnPrefix   string
  Version     string
}

type ServeSwaggerResult struct {
  SwaggerPath string
}

func ServeSwagger(e *echo.Echo, config SwaggerConfig) *ServeSwaggerResult {
  cdnPrefix := strings.TrimSpace(config.CdnPrefix)
  cdnPrefix = strings.TrimSuffix(cdnPrefix, "/")
  if len(cdnPrefix) == 0 {
    panic(fmt.Sprintf("SwaggerConfig.CdnPrefx must be provided."))
  }

  t := template.Must(template.ParseFS(SwaggerUiFS, "swagger-ui-4.10.3/*.go.html"))
  // t := &Template {
  //     templates: template.Must(template.ParseGlob("public/views/*.html")),
  // }
  // http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
  // 	t.ExecuteTemplate(rw, "index.go.html", map[string]string{"title": "Golang Embed 测试"})
  // })
  fs := echo.MustSubFS(SwaggerUiFS, "swagger-ui-4.10.3")

  prefixed := func(orignalUrl string) string {
    if len(config.UrlPrefix) == 0 {
      return orignalUrl
    }
    urlPrefix := config.UrlPrefix
    if urlPrefix[0:1] != "/" {
      urlPrefix += "/" + config.UrlPrefix
    }
    return urlPrefix + orignalUrl
  }

  for _, tmpl := range t.Templates() {
    fmt.Printf("loading template >> %s\n", tmpl.Name())
  }

  indexHandler := func(c echo.Context) error {
    apiDocUrl := config.ApiDocUrl
    if len(config.ApiDocUrl) == 0 {
      apiDocUrl = prefixed("/swagger/api-docs")
    }
    params := map[string]string{"url": apiDocUrl, "cdnPrefix": cdnPrefix}
    if err := t.ExecuteTemplate(c.Response().Writer, "index.go.html", params); err != nil {
      return err
    }
    // c.Response().WriteHeader(http.StatusOK)
    return nil
  }
  e.GET(prefixed("/swagger/index.html"), indexHandler)
  e.GET(prefixed("/swagger/index"), indexHandler)
  e.GET(prefixed("/swagger/api-docs"), GenApiDoc(config.Title, config.Description, config.Version))
  e.GET(prefixed("/swagger"), func(c echo.Context) error {
    c.Redirect(301, prefixed("/swagger/index.html"))
    return nil
  })
  e.StaticFS(prefixed("/swagger"), fs)

  return &ServeSwaggerResult{SwaggerPath: prefixed("/swagger")}
}

// func SwaggerUiHandler() http.Handler {
// 	return http.FileServer(http.FS(swaggerUi))
// }
