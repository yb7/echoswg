package echoswg

import (
  "testing"
  "reflect"
)

type Tag struct {
  ID int64
  Name string
}
type Pet struct {
  ID *int64
  Name string
  PhotoUrls []string
  Tags []Tag
}

type A struct {
  B struct {
    Name string
  }
}

func TestUniqueTypeName(t *testing.T) {
  var pet = Pet{}
  builder := NewTypeDefBuilder("github.com/yb7/echoswg")

  var a = A{}
  uniqueName := builder.uniqueStructName(reflect.TypeOf(&a.B))
  if uniqueName != "anonymous00" {
    t.Fatalf("bad unique defination name: %s", uniqueName)
  }
  var a2 = A{}
  uniqueName = builder.uniqueStructName(reflect.TypeOf(&a2.B))
  if uniqueName != "anonymous00" {
    t.Fatalf("bad unique defination name: %s", uniqueName)
  }

  uniqueName = builder.uniqueStructName(reflect.TypeOf(pet))

  if uniqueName != "Pet" {
    t.Fatalf("bad unique defination name: %s", uniqueName)
  }
}

func TestIsScalarType(t *testing.T) {
  builder := NewTypeDefBuilder("github.com/yb7/echoswg")
  var i32 = int32(0)
  scalarType := builder.ToSwaggerType(reflect.TypeOf(i32))
  if scalarType.String() != "optional: false, type: integer, format: int32" {
    t.Fatalf("actual is %s", scalarType.String())
  }
  scalarType = builder.ToSwaggerType(reflect.TypeOf(&i32))
  if scalarType.String() != "optional: true, type: integer, format: int32" {
    t.Fatalf("actual is %s", scalarType.String())
  }
  scalarType = builder.ToSwaggerType(reflect.TypeOf([]string{}))
  if scalarType.String() != "type: array, items: [optional: false, type: string, format: string]" {
    t.Fatalf("actual is %s", scalarType.String())
  }
  scalarType = builder.ToSwaggerType(reflect.TypeOf(Pet{}))
  if scalarType.String() != "$ref: #/definitions/Pet" {
    t.Fatalf("actual is %s", scalarType.String())
  }
}
