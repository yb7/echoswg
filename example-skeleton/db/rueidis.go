package db

import (
    "github.com/redis/rueidis"
    "github.com/yb7/alilog"

    "github.com/yb7/echoswg/example-skeleton/config"
)

var RedisClient rueidis.Client

func InitRueidisClient() {
    if config.C == nil || config.C.Redis == nil || config.C.Redis.Host == "" {
        alilog.Warnf("skip redis init because redis config is empty")
        return
    }

    client, err := rueidis.NewClient(rueidis.ClientOption{
        InitAddress: []string{config.C.Redis.Host},
        Username:    config.C.Redis.Username,
        Password:    config.C.Redis.Password,
        SelectDB:    config.C.Redis.Database,
        ClientName:  config.C.Redis.ClientName,
    })
    if err != nil {
        alilog.Warnf("init redis client failed: %v", err)
        return
    }
    RedisClient = client
}

func CloseRedis() {
    if RedisClient != nil {
        RedisClient.Close()
    }
}
