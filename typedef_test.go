package echoswg

import (
	"reflect"
	"testing"
)

type Tag struct {
	ID   int64
	Name string
}
type Pet struct {
	ID        *int64
	Name      string
	PhotoUrls []string
	Tags      []Tag
}

type A struct {
	B struct {
		Name string
	}
}
type C struct {
	B struct {
		Name string
	}
}
type D struct {
	B struct {
		Name string
		age  int
	}
}
type PagedData[T any] struct {
	Data []*T `json:"data"`
}
type NewsVo struct {
}

func TestUniqueTypeName(t *testing.T) {
	builder := NewTypeDefBuilder()

	testCases := [][]any{
		{(&A{}).B, "anonymous"},
		{(&A{}).B, "anonymous"},
		{(&C{}).B, "anonymous"},
		{(&D{}).B, "anonymous001"},
		{Pet{}, "Pet"},
		{PagedData[*NewsVo]{}, "PagedDataNewsVo"},
	}

	for _, testCase := range testCases {
		obj := testCase[0]
		expected := testCase[1]
		uniqueName := builder.uniqueStructName(reflect.TypeOf(obj))
		if uniqueName != expected {
			t.Fatalf("bad unique defination name: %s, expect is %s", uniqueName, expected)
		}
	}
}

// func TestToSwaggerType(t *testing.T) {
// 	builder := NewTypeDefBuilder()
// 	var i32 = int32(0)
// 	swaggerType := builder.ToSwaggerType(reflect.TypeOf(i32))
// 	if swaggerType.String() != "optional: false, type: integer, format: int32" {
// 		t.Fatalf("actual is %s", swaggerType.String())
// 	}
// 	swaggerType = builder.ToSwaggerType(reflect.TypeOf(&i32))
// 	if swaggerType.String() != "optional: true, type: integer, format: int32" {
// 		t.Fatalf("actual is %s", swaggerType.String())
// 	}
// 	swaggerType = builder.ToSwaggerType(reflect.TypeOf([]string{}))
// 	if swaggerType.String() != "type: array, items: [optional: false, type: string, format: string]" {
// 		t.Fatalf("actual is %s", swaggerType.String())
// 	}
// 	swaggerType = builder.ToSwaggerType(reflect.TypeOf(Pet{}))
// 	if swaggerType.String() != "$ref: #/definitions/Pet" {
// 		t.Fatalf("actual is %s", swaggerType.String())
// 	}
// }

// func TestStructDefinitions(t *testing.T) {
//   type B struct {
//     API string `json:"api"`
//   }
//   type A struct {
//     Name string
//     B B
//   }
//   type C struct {
//     Age int
//   }

//   GlobalTypeDefBuilder.Build(reflect.TypeOf(A{}))
//   GlobalTypeDefBuilder.Build(reflect.TypeOf(C{}))
//   s, _ := json.Marshal(GlobalTypeDefBuilder.StructDefinitions)
//   fmt.Println(string(s))
//   matched := reflect.DeepEqual(GlobalTypeDefBuilder.StructDefinitions, map[string]map[string]interface{} {
//       "A": {
//         "properties": map[string]interface{} {
//           "B": map[string]interface{} {
//             "$ref": "#/definitions/B",
//           },
//           "Name": map[string]interface{} {
//             "format": "string",
//             "type": "string",
//           },
//         },
//         "required": []string{"Name", "B"},
//         "type": "object",
//       },
//       "B": {
//         "properties": map[string]interface{} {
//           "api": map[string]interface{} {
//             "format": "string",
//             "type": "string",
//           },
//         },
//         "required": []string{"api"},
//         "type": "object",
//       },
//       "C": {
//         "properties": map[string]interface{} {
//           "Age": map[string]interface{} {
//             "format": "int32",
//             "type": "integer",
//           },
//         },
//         "required": []string{"Age"},
//         "type": "object",
//       },
//   })
//   if !matched {
//     t.Failed()
//   }
// }
