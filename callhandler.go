package echoswg

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"github.com/labstack/echo"
  "github.com/go-playground/locales/en"
  ut "github.com/go-playground/universal-translator"
  "gopkg.in/go-playground/validator.v9"
  en_translations "gopkg.in/go-playground/validator.v9/translations/en"
	"errors"
	"github.com/gorilla/schema"
	"time"
  "strings"
)

var (
  uni      *ut.UniversalTranslator
  validate *validator.Validate
  trans    ut.Translator
)

func init() {
  en := en.New()
  uni = ut.New(en, en)

  // this is usually know or extracted from http 'Accept-Language' header
  // also see uni.FindTranslator(...)
  trans, _ = uni.GetTranslator("en")

  validate = validator.New()
  en_translations.RegisterDefaultTranslations(validate, trans)
}

// BuildEchoHandler func
func BuildEchoHandler(fullRequestPath string, handlers []interface{}) echo.HandlerFunc {
	inTypes, _, _ := validateChain(handlers)

	return func(c echo.Context) error {
		// var requestObj reflect.Value
		StartAt := time.Now()

		var logError = func (err error) error {
			fmt.Printf("%6s | %3d [%.3fs] | %s\n", c.Request().Method(),
				c.Response().Status(), time.Now().Sub(StartAt).Seconds(),
				fullRequestPath)
			return err
		}
		var err error
		//var c = NewGonextContextFromEcho(echoContext)
		inParams := make(map[reflect.Type]reflect.Value)
		inParams[reflect.TypeOf((*echo.Context)(nil)).Elem()] = reflect.ValueOf(c)
		for _, inType := range inTypes {
			requestObj, err := newType(fullRequestPath, inType, c)
			if err != nil {
				return logError(err)
			}
			inParams[inType] = requestObj
		}

		var lastHandler interface{}
		var out []reflect.Value

		//fmt.Printf("call %s\n", fullRequestPath)
		//for inParamKey := range inParams {
		//	fmt.Printf("    in[%s]\n", inParamKey)
		//}
		for _, h := range handlers {
			lastHandler = h
			out, err = callHandler(h, inParams)
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

func callHandler(handler interface{}, inParams map[reflect.Type]reflect.Value) ([]reflect.Value, error) {
	handlerRef := reflect.ValueOf(handler)
	var params []reflect.Value
	for i := 0; i < handlerRef.Type().NumIn(); i++ {
		v, ok := inParams[handlerRef.Type().In(i)]
		if ok {
			params = append(params, v)
		} else {
			msg := fmt.Sprintf("cannot find inParam of [%v]", handlerRef.Type().In(i))
			fmt.Println(msg)
			return nil, errors.New(msg)
		}

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
func newType(fullRequestPath string, typ reflect.Type, c echo.Context) (reflect.Value, error) {
	requestType := typ
	if requestType.Kind() == reflect.Ptr {
		requestType = requestType.Elem()
	}
	requestObj := reflect.New(requestType)

	pathAndQueryParams := c.QueryParams()

	for _, name := range c.ParamNames() {
		pathAndQueryParams[name] = []string{c.Param(name)}
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

      if (field.Name == "Body") {
        err = c.Bind(value)
      } else {
        err = decoder.Decode(value, pathAndQueryParams)
      }
			if err != nil {
				return requestObj, err
			}

			if theType.Kind() == reflect.Ptr {
				requestObj.Elem().FieldByName(field.Name).Set(reflect.ValueOf(value))
			} else {
				requestObj.Elem().FieldByName(field.Name).Set(reflect.ValueOf(value).Elem())
			}
		}
	}
	if err = validate.Struct(requestObj.Interface()); err != nil {
    // translate all error at once
    errs := err.(validator.ValidationErrors)

    // returns a map with key = namespace & value = translated error
    return requestObj, TranslatedValidationErrors {
      Data: errs.Translate(trans),
    }
  }
	return requestObj, nil
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
