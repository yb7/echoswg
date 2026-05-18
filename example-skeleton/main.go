package main

import (
    "os"

    _ "github.com/yb7/echoswg/example-skeleton/controller"
    "github.com/yb7/echoswg/example-skeleton/config"
    "github.com/yb7/echoswg/example-skeleton/db"
    "github.com/yb7/echoswg/example-skeleton/util"

    "net/http"

    "github.com/labstack/echo/v5/middleware"
    "github.com/yb7/alilog"
    "github.com/yb7/echoswg"
)

func main() {
    if err := os.Setenv("TZ", "Asia/Shanghai"); err != nil {
        alilog.Fatal(err)
    }

    config.EnsureInitSuccess()

    db.InitRueidisClient()
    defer db.CloseRedis()

    db.OpenDB()
    defer db.CloseDB()

    db.DbMigrate()

    e := util.EchoInstance

    echoswg.ServeSwagger(e, echoswg.SwaggerConfig{
        UrlPrefix:   "/api",
        Title:       "Example Skeleton API",
        Description: "A standalone echoswg skeleton that does not depend on the example directory.",
        CdnPrefix:   "https://statics.stock001.com/swagger-ui-5.32.5",
    })

    e.Use(middleware.RequestLogger())
    e.Use(middleware.Gzip())
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
    }))
    e.Use(util.EchoRecover)

    alilog.Infof("rest server started at port [%s]", config.C.Ports.Http)
    if err := e.Start(config.C.Ports.Http); err != nil {
        alilog.Fatal(err)
    }
}
