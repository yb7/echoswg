package bizerrors

import (
    "net/http"
    "strings"
)

type BizError struct {
    HttpStatus   int    `json:"-"`
    Success      bool   `json:"success"`
    ErrorCode    string `json:"errorCode"`
    ErrorMessage string `json:"errorMessage"`
    ShowType     int    `json:"showType"`
}

func (e *BizError) Error() string {
    return e.ErrorMessage
}

func (e *BizError) Code() string {
    return e.ErrorCode
}

func New(httpStatus int, errorCode string, message string) *BizError {
    return &BizError{
        HttpStatus:   httpStatus,
        Success:      false,
        ErrorCode:    errorCode,
        ErrorMessage: message,
        ShowType:     2,
    }
}

func BadRequest(msg string) *BizError {
    return New(http.StatusBadRequest, "BAD_REQUEST", msg)
}

var Unauthorized = New(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
var MissingAccessToken = New(http.StatusUnauthorized, "MISSING_ACCESS_TOKEN", "missing access token")

func MissingPermissions(roles ...string) *BizError {
    msg := "missing permissions"
    if len(roles) > 0 {
        msg += ": " + strings.Join(roles, ",")
    }
    return New(http.StatusForbidden, "MISSING_PERMISSIONS", msg)
}
