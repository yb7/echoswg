package echoswg

import (
	"reflect"
)

type PetPutA struct {
	ID       int64 `desc:"the id"`
	Category *string
	Body     struct {
		Name string
		Age  int
	}
}
type PetPutB struct {
	Color string
	Body  *struct {
		Name  string
		Age   int
		Owner string
	}
}

func newRequestParam() *RequestParam {
	return BuildRequestParam("/pets/:ID", []reflect.Type{reflect.TypeOf(PetPutA{}), reflect.TypeOf(PetPutB{})})
}

// func TestBuildRequestParam(t *testing.T) {
// 	requestParam := newRequestParam()
// 	for _, pathParam := range requestParam.PathParams {
// 		switch pathParam.Name {
// 		case "ID":
// 			matched := pathParam.Type == reflect.TypeOf(int64(0)) && pathParam.Required == true
// 			if !matched {
// 				t.Fatalf("not matched path param for 'ID'")
// 			}
// 		default:
// 			t.Fatalf("bad path param: [%s]", pathParam.Name)
// 		}
// 	}
// 	for _, q := range requestParam.QueryParams {
// 		switch q.Name {
// 		case "Category":
// 			matched := q.Type == reflect.TypeOf("string") && q.Required == false
// 			if !matched {
// 				t.Fatalf("not matched path param for 'ID'")
// 			}
// 		case "Color":
// 			matched := q.Type == reflect.TypeOf("string") && q.Required == true
// 			if !matched {
// 				t.Fatalf("not matched path param for 'ID'")
// 			}
// 		default:
// 			t.Fatalf("bad param in query [%s]", q.Name)
// 		}
// 	}
// 	if requestParam.RequestBody != reflect.TypeOf(PetPutB{}.Body) {
// 		t.Fatalf("request body must be the last one: %s", PetPutB{}.Body)
// 	}
// }

// func TestRequestParam_ToSwaggerJSON(t *testing.T) {
//   requestParam := newRequestParam()
//   sjson := requestParam.ToSwaggerJSON()

//   matched := reflect.DeepEqual(sjson, []map[string]interface{} {
//     {
//       "name": "ID",
//       "in": "path",
//       "format": "int64",
//       "required": true,
//       "type": "integer",
//       "description": "the id",
//     },
//     {
//       "name": "Category",
//       "in": "query",
//       "format": "string",
//       "required": false,
//       "type": "string",
//       "description": "",
//     },
//     {
//       "name": "Color",
//       "in": "query",
//       "format": "string",
//       "required": true,
//       "type": "string",
//       "description": "",
//     },
//     {
//       "name": "body",
//       "in": "body",
//       "required": true,
//       "schema": map[string]interface{} {
//         "$ref": "#/definitions/anonymous00",
//       },
//     },
//   })
//   if !matched {
//     t.Fatal(sjson)
//   }
// }
