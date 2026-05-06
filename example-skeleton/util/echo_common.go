package util

import (
    "net/http"

    "github.com/labstack/echo/v4"

    "github.com/yb7/echoswg/example-skeleton/bizerrors"
)

func init() {
    EchoInstance.HTTPErrorHandler = func(err error, c echo.Context) {
        if c.Response().Committed {
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
