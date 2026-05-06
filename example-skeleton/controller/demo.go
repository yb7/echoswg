package controller

import (
    "github.com/yb7/echoswg"

    "github.com/yb7/echoswg/example-skeleton/security"
    "github.com/yb7/echoswg/example-skeleton/service"
    "github.com/yb7/echoswg/example-skeleton/util"
)

type DemoController struct{}

func init() {
    c := new(DemoController)
    g := echoswg.NewApiGroup(util.EchoInstance, "Demo", "/api/demos")
    g.POST("", security.RequireAuth(security.RoleAnonymous), c.Create, echoswg.WithOperationId("createDemo"), echoswg.WithDescription("Create a demo resource without requiring a real login token in the skeleton project."))
}

func (*DemoController) Create(ctx security.AuthCtx, req *struct {
    Body *service.CreateDemoVo `desc:"创建Demo"`
}) (*service.DemoVo, error) {
    return service.Demo.Create(ctx, req.Body)
}
