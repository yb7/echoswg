package security

import (
    "context"
    "time"
)

type AuthCtx interface {
    context.Context
    GetUserPrincipal() *UserPrincipal
    WithValue(key, val any)
}

var GlobalCtx AuthCtx

const KEY_USER_PRINCIPAL = "user_principal"

func init() {
    globalCtx := context.WithValue(context.Background(), KEY_USER_PRINCIPAL, &UserPrincipal{
        UserID:      0,
        Name:        "SYSTEM_ROOT",
        Roles:       make([]string, 0),
        PhoneNumber: 18900000000,
        AccessToken: "",
    })
    GlobalCtx = &authCtxImpl{Context: globalCtx}
}

type authCtxImpl struct {
    context.Context
}

func (ctx *authCtxImpl) GetUserPrincipal() *UserPrincipal {
    userPrincipal, ok := ctx.Value(KEY_USER_PRINCIPAL).(*UserPrincipal)
    if !ok {
        return nil
    }
    return userPrincipal
}

func (ctx *authCtxImpl) WithValue(key, val any) {
    ctx.Context = context.WithValue(ctx.Context, key, val)
}

func (ctx *authCtxImpl) Deadline() (deadline time.Time, ok bool) {
    return ctx.Context.Deadline()
}

func (ctx *authCtxImpl) Done() <-chan struct{} {
    return ctx.Context.Done()
}

func (ctx *authCtxImpl) Err() error {
    return ctx.Context.Err()
}

func (ctx *authCtxImpl) Value(key any) any {
    return ctx.Context.Value(key)
}
