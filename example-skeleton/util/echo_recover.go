package util

import (
    "fmt"
    "net/http/httputil"
    "runtime"

    "github.com/labstack/echo/v5"
    "github.com/yb7/alilog"

    "github.com/yb7/echoswg/example-skeleton/bizerrors"
)

var StackSize = 4 << 10

func EchoRecover(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) (err error) {
        defer func() {
            if r := recover(); r != nil {
                if bizError, ok := r.(*bizerrors.BizError); ok {
                    err = c.JSON(bizError.HttpStatus, bizError)
                    return
                }

                recovered, ok := r.(error)
                if !ok {
                    recovered = fmt.Errorf("%v", r)
                }

                stack := make([]byte, StackSize)
                length := runtime.Stack(stack, true)
                reqDump, _ := httputil.DumpRequest(c.Request(), true)

                alilog.Errorf("[PANIC RECOVER] Request\n%s", string(reqDump))
                alilog.Errorf("[PANIC RECOVER] %v\n%s", recovered, string(stack[:length]))

                err = recovered
            }
        }()
        return next(c)
    }
}
