package util

import (
    "net/http"

    "github.com/labstack/echo/v5"

    "github.com/yb7/echoswg/example-skeleton/bizerrors"
)

func init() {
    EchoInstance.HTTPErrorHandler = func(c *echo.Context, err error) {
        if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil && resp.Committed {
            return
        }

        var bizErr *bizerrors.BizError
        switch err := err.(type) {
        case *echo.HTTPError:
            bizErr = &bizerrors.BizError{
                HttpStatus:   err.Code,
                Success:      false,
                ErrorCode:    "HTTP_ERROR",
                ErrorMessage: err.Error(),
                ShowType:     2,
            }
        case *bizerrors.BizError:
            bizErr = err
        default:
            bizErr = &bizerrors.BizError{
                HttpStatus:   http.StatusInternalServerError,
                Success:      false,
                ErrorCode:    "SYS_ERROR",
                ErrorMessage: err.Error(),
                ShowType:     2,
            }
        }
        _ = c.JSON(bizErr.HttpStatus, bizErr)
    }
}
