package service

import "github.com/yb7/echoswg/example-skeleton/security"

var Demo = &demoService{}

type demoService struct{}

type CreateDemoVo struct {
    Name string `json:"name" validate:"required" desc:"演示名称"`
}

type DemoVo struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    CurrentUser string `json:"currentUser,omitempty"`
}

func (*demoService) Create(ctx security.AuthCtx, req *CreateDemoVo) (*DemoVo, error) {
    result := &DemoVo{
        ID:   1,
        Name: req.Name,
    }
    if user := ctx.GetUserPrincipal(); user != nil {
        result.CurrentUser = user.Name
    }
    return result, nil
}
