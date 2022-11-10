package echoswg

import (
	"bytes"
	"fmt"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/labstack/echo/v4"
)

var (
	uni *ut.UniversalTranslator
	//validate *validator.Validate
	trans ut.Translator
)

func init() {
	en := en.New()
	zh := zh.New()
	uni = ut.New(en, zh)

}

type HandlerConfig struct {
	DisableLog bool
}

// BuildEchoHandler func
func BuildEchoHandler(fullRequestPath string, config HandlerConfig, handlers []interface{}) echo.HandlerFunc {
	//inTypes, _, _ := validateChain(handlers)

	return func(c echo.Context) error {
		// var requestObj reflect.Value
		StartAt := time.Now()

		var logError = func(err error) error {
			if !config.DisableLog {
				fmt.Printf("%6s | %3d [%.3fs] | %s\n", c.Request().Method,
					c.Response().Status, time.Now().Sub(StartAt).Seconds(),
					fullRequestPath)
			}
			return err
		}
		var err error
		//var c = NewGonextContextFromEcho(echoContext)
		inParams := make(map[reflect.Type]reflect.Value)
		inParams[reflect.TypeOf((*echo.Context)(nil)).Elem()] = reflect.ValueOf(c)
		//for _, inType := range inTypes {
		//	requestObj, err := newType(fullRequestPath, inType, c)
		//	if err != nil {
		//		return logError(err)
		//	}
		//	inParams[inType] = requestObj
		//}

		var lastHandler interface{}
		var out []reflect.Value

		//fmt.Printf("call %s\n", fullRequestPath)
		//for inParamKey := range inParams {
		//	fmt.Printf("    in[%s]\n", inParamKey)
		//}
		for _, h := range handlers {
			lastHandler = h
			out, err = callHandler(h, inParams, c)
			if err != nil {
				return logError(err)
			}
		}
		if len(out) > 1 {
			return logError(fmt.Errorf("return more then one data value is not supported: %s", runtime.FuncForPC(reflect.ValueOf(lastHandler).Pointer()).Name()))
		} else if len(out) == 0 {
			return logError(c.NoContent(http.StatusOK))
		} else {
			return logError(c.JSON(http.StatusOK, out[0].Interface()))
		}
	}
}

func callHandler(handler interface{}, inParams map[reflect.Type]reflect.Value, c echo.Context) ([]reflect.Value, error) {
	handlerRef := reflect.ValueOf(handler)
	var params []reflect.Value
	for i := 0; i < handlerRef.Type().NumIn(); i++ {
		paramType := handlerRef.Type().In(i)
		v, ok := inParams[paramType]
		var err error
		if !ok {
			v, err = newType(paramType, c)
			if err != nil {
				//msg := fmt.Sprintf("error in build input param of [%v], %s", paramType, err.Error())
				//fmt.Println(msg)
				return nil, err
			}
			inParams[paramType] = v
		}
		params = append(params, v)
	}
	values := handlerRef.Call(params)

	var notErrors []reflect.Value
	var err error
	for _, value := range values {
		if value.Interface() != nil {
			if !isErrorType(value) {
				inParams[value.Type()] = value
				notErrors = append(notErrors, value)
			} else {
				err = value.Interface().(error)
			}
		}
	}
	return notErrors, err
}

func isErrorType(v reflect.Value) bool {
	return v.MethodByName("Error").IsValid()
}
func newType(typ reflect.Type, c echo.Context) (reflect.Value, error) {
	requestType := typ
	if requestType.Kind() == reflect.Ptr {
		requestType = requestType.Elem()
	}
	requestObj := reflect.New(requestType)

	pathAndQueryParams := c.QueryParams()

	for _, name := range c.ParamNames() {
		value := c.Param(name)
		for _, maybeName := range strings.Split(name, ",") {
			pathAndQueryParams[maybeName] = []string{value}
		}
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(requestObj.Interface(), pathAndQueryParams)
	if err != nil {
		return requestObj, err
	}
	for i := 0; i < requestType.NumField(); i++ {
		field := requestType.Field(i)

		if field.Name == "Body" || field.Anonymous {
			theType := field.Type
			var value interface{}
			if theType.Kind() == reflect.Ptr {
				value = reflect.New(field.Type.Elem()).Interface()
			} else {
				value = reflect.New(field.Type).Interface()
			}

			if field.Name == "Body" {
				buf, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return requestObj, err
				}
				c.Request().Body = io.NopCloser(bytes.NewBuffer(buf))
				if err = c.Bind(value); err != nil {
					return requestObj, err
				}
				c.Request().Body = io.NopCloser(bytes.NewBuffer(buf)) // for next handler
			} else {
				err = decoder.Decode(value, pathAndQueryParams)
			}
			if err != nil {
				return requestObj, err
			}

			targetField := requestObj.Elem().FieldByName(field.Name)
			if targetField.CanSet() {
				if theType.Kind() == reflect.Ptr {
					targetField.Set(reflect.ValueOf(value))
				} else {
					targetField.Set(reflect.ValueOf(value).Elem())
				}
			}
		}
	}

	validate := getValidator(c)

	//if len(acceptLanguage) == 0 {
	//  acceptLanguage :=  c.Request().URL.t("Accept-Language")
	//}
	if err = validate.Struct(requestObj.Interface()); err != nil {
		// translate all error at once
		errs := err.(validator.ValidationErrors)

		// returns a map with key = namespace & value = translated error
		return requestObj, TranslatedValidationErrors{
			Data: errs.Translate(trans),
		}
	}
	return requestObj, nil
}

func newValidate() *validator.Validate {
	validate := validator.New()

	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return validate
}

func getValidator(c echo.Context) *validator.Validate {
	for _, local := range acceptLanguage(c) {
		if strings.HasPrefix(local, "zh") {
			trans, _ = uni.GetTranslator("zh")
			validate := newValidate()

			zh_translations.RegisterDefaultTranslations(validate, trans)
			return validate
		}
	}
	trans, _ = uni.GetTranslator("en")

	validate := newValidate()
	en_translations.RegisterDefaultTranslations(validate, trans)
	return validate
}

// this is usually know or extracted from http 'Accept-Language' header
// also see uni.FindTranslator(...)
func acceptLanguage(c echo.Context) []string {
	var defaultLanguage = []string{"zh"}
	acceptLanguage := c.Request().Header.Get("Accept-Language")
	if len(acceptLanguage) == 0 {
		return defaultLanguage
	}
	prefix := strings.Split(acceptLanguage, ";")[0]
	return strings.Split(prefix, ",")
}

type TranslatedValidationErrors struct {
	Data validator.ValidationErrorsTranslations
}

func (e TranslatedValidationErrors) Error() string {
	var msg []string
	for k, v := range e.Data {
		msg = append(msg, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(msg, "\n")
}
