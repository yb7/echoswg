package config

import (
    "fmt"
    "reflect"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/spf13/viper"
)

var C = &Configuration{
    Redis: &RedisConfig{
        Database: 0,
    },
}

type Configuration struct {
    Ports *PortsConfig `validate:"required"`
    Db    *DbConfig    `validate:"required"`
    Redis *RedisConfig `validate:"required"`
}

type PortsConfig struct {
    Http string `validate:"required"`
}

type DbConfig struct {
    Url string `validate:"required"`
}

type RedisConfig struct {
    Host       string `validate:"required"`
    Password   string
    Username   string
    ClientName string
    Database   int `validate:"gte=0,lte=16"`
}

var configInitError error

func EnsureInitSuccess() {
    if configInitError != nil || C == nil {
        panic(fmt.Sprintf("config init failed: %v", configInitError))
    }
}

func init() {
    viper.SetConfigName("config")
    viper.SetConfigType("toml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("/etc/app/")
    viper.AddConfigPath("$HOME/.app")

    bindAllKeys(reflect.TypeOf(*C))

    if err := viper.ReadInConfig(); err != nil {
        configInitError = err
        return
    }
    if err := viper.Unmarshal(C); err != nil {
        configInitError = err
        C = nil
        return
    }

    validate := validator.New()
    if err := validate.Struct(C); err != nil {
        configInitError = err
        C = nil
        return
    }
}

func bindAllKeys(t reflect.Type) {
    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        walkThroughFields(envKeyName(f), f.Type)
    }
}

func envKeyName(f reflect.StructField) string {
    envKey, ok := f.Tag.Lookup("env")
    if ok {
        return envKey
    }
    return strings.ToLower(f.Name[0:1]) + f.Name[1:]
}

func walkThroughFields(fieldKey string, t reflect.Type) {
    if t.Kind() == reflect.Ptr {
        walkThroughFields(fieldKey, t.Elem())
        return
    }
    if t.Kind() == reflect.Struct {
        for i := 0; i < t.NumField(); i++ {
            f := t.Field(i)
            walkThroughFields(fieldKey+"."+envKeyName(f), f.Type)
        }
        return
    }
    alterKey := strings.ReplaceAll(fieldKey, ".", "_")
    upperAlterKey := strings.ToUpper(alterKey)
    _ = viper.BindEnv(fieldKey, fieldKey, alterKey, upperAlterKey)
}
