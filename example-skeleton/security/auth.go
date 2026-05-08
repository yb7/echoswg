package security

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/labstack/echo/v5"
    "github.com/redis/rueidis"
    "github.com/yb7/alilog"

    "github.com/yb7/echoswg/example-skeleton/bizerrors"
    "github.com/yb7/echoswg/example-skeleton/db"
)

const COOKIE_ACCESSTOKEN = "example_skeleton_access_token"

type UserPrincipal struct {
    UserID      int      `json:"userId"`
    PhoneNumber int64    `json:"phoneNumber"`
    Name        string   `json:"name"`
    Roles       []string `json:"roles"`
    AccessToken string   `json:"accessToken"`
}

func (user UserPrincipal) FindRole(role Role) (Role, bool) {
    r, ok := user.FindRoleWithPrefix(string(role))
    return Role(r), ok
}

func (user UserPrincipal) FindRoleWithPrefix(rolePrefix string) (string, bool) {
    for _, role := range user.Roles {
        if strings.HasPrefix(role, rolePrefix) {
            return role, true
        }
    }
    return "", false
}

func getAccessTokenInRequest(ctx *echo.Context) string {
    authorization := ctx.Request().Header.Get("Authorization")
    if authorization != "" {
        if !strings.HasPrefix(authorization, "Bearer ") {
            return ""
        }
        return strings.TrimPrefix(authorization, "Bearer ")
    }
    cookie, err := ctx.Cookie(COOKIE_ACCESSTOKEN)
    if err == nil && cookie != nil {
        return cookie.Value
    }
    return ""
}

func RequireAuth(requiredRoles ...Role) func(ctx *echo.Context) (AuthCtx, error) {
    return func(ctx *echo.Context) (AuthCtx, error) {
        if len(requiredRoles) == 1 && requiredRoles[0] == RoleAnonymous {
            return &authCtxImpl{Context: ctx.Request().Context()}, nil
        }

        accessToken := getAccessTokenInRequest(ctx)
        if accessToken == "" {
            return nil, bizerrors.MissingAccessToken
        }

        userPrincipal, err := getUserPrincipalByAccessToken(GlobalCtx, accessToken)
        if err != nil {
            return nil, err
        }

        missingRoles := make([]string, 0)
        for _, requiredRole := range requiredRoles {
            _, ok := userPrincipal.FindRoleWithPrefix(string(requiredRole))
            if !ok {
                missingRoles = append(missingRoles, string(requiredRole))
            }
        }
        if len(missingRoles) > 0 {
            return nil, bizerrors.MissingPermissions(missingRoles...)
        }

        log := alilog.WithTraceId(
            fmt.Sprintf("requestUserId=%d", userPrincipal.UserID),
            fmt.Sprintf("requestUserPhone=%d", userPrincipal.PhoneNumber),
        )
        authCtx := log.CreateLogContext(context.Background())
        authCtx = context.WithValue(authCtx, KEY_USER_PRINCIPAL, userPrincipal)
        return &authCtxImpl{Context: authCtx}, nil
    }
}

func AccessTokenKeyInRedis(token string) (string, error) {
    parts := strings.Split(token, "@")
    if len(parts) < 2 {
        return "", alilog.Errorf("bad access token format: %s", token)
    }
    return fmt.Sprintf("example-skeleton:access-token:%s", parts[0]), nil
}

func getUserPrincipalByAccessToken(ctx context.Context, accessToken string) (*UserPrincipal, error) {
    if db.RedisClient == nil {
        return nil, bizerrors.Unauthorized
    }

    tokenKey, err := AccessTokenKeyInRedis(accessToken)
    if err != nil {
        return nil, err
    }

    cmd := db.RedisClient.B().Get().Key(tokenKey).Build()
    dat, err := db.RedisClient.Do(ctx, cmd).AsBytes()
    if err != nil {
        if rueidis.IsRedisNil(err) {
            return nil, bizerrors.Unauthorized
        }
        return nil, bizerrors.Unauthorized
    }
    if len(dat) == 0 {
        return nil, bizerrors.Unauthorized
    }

    userPrincipal := &UserPrincipal{}
    if err := json.Unmarshal(dat, userPrincipal); err != nil {
        return nil, alilog.Error(err)
    }
    if userPrincipal.AccessToken != accessToken {
        return nil, bizerrors.Unauthorized
    }
    return userPrincipal, nil
}
