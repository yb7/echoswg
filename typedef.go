package echoswg

import (
  "reflect"
  "strings"
  "fmt"
)

type TypeDefBuilder struct {
  RootPath string
  cachedTypes []reflect.Type
  position int
  StructDefinitions map[reflect.Type]*SwaggerType
  anonymousTypes map[reflect.Type]string
}

func NewTypeDefBuilder(rootPath string) *TypeDefBuilder {
  return &TypeDefBuilder {
    RootPath: rootPath,
    cachedTypes: make([]reflect.Type, 0),
    position: 0,
    StructDefinitions: make(map[reflect.Type]*SwaggerType),
    anonymousTypes: make(map[reflect.Type]string),
  }
}
func (b *TypeDefBuilder) Build(typ reflect.Type) *SwaggerType {
  swaggerType := b.ToSwaggerType(typ)

  for b.position < len(b.cachedTypes) {
    pendingType := b.cachedTypes[b.position]
    if _, ok := b.StructDefinitions[pendingType]; !ok {
      b.StructDefinitions[pendingType] = b.ToSwaggerType(pendingType)
    }
    b.position += 1
  }
  return swaggerType
}


func (b *TypeDefBuilder) uniqueStructName(typ reflect.Type) string {
  typeName := typ.Name()
  if len(typeName) == 0 {
    typeName = fmt.Sprintf("anonymous%02d", len(b.anonymousTypes))
  }
  pkgPath := strings.Replace(strings.TrimPrefix(typ.PkgPath(), b.RootPath), "/", "_", -1)
  pkgPath = strings.Replace(pkgPath, ".", "_", -1)


  uniqueName := pkgPath
  if len(uniqueName) > 0 {
    uniqueName += "_"
  }
  uniqueName += typeName
  return uniqueName
}

type SwaggerType struct {
  Optional bool
  Type string
  Format string
  Items *SwaggerType
}

func (t *SwaggerType) String() string {
  if t == nil {
    return ""
  }
  switch t.Type {
  case "array": return fmt.Sprintf("type: array, items: [%s]", t.Items.String())
  case "object": return fmt.Sprintf("$ref: %s", t.Format)
  default:
    return fmt.Sprintf("optional: %t, type: %s, format: %s", t.Optional, t.Type, t.Format)
  }
}
func (b *TypeDefBuilder) ToSwaggerType(typ reflect.Type) *SwaggerType {
  v := &SwaggerType{}
  b._toSwaggerType(typ, v)
  return v
}
func (b *TypeDefBuilder) _toSwaggerType(typ reflect.Type, dest *SwaggerType) {
  if typ == TimeType {
    dest.Type = "string"
    dest.Format = "string"
    return
  }
  switch typ.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
    reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
    dest.Type = "integer"
    dest.Format = "int32"
    return
  case reflect.Int64, reflect.Uint64:
    dest.Type = "integer"
    dest.Format = "int64"
    return
  case reflect.String:
    dest.Type = "string"
    dest.Format = "string"
    return
  case reflect.Float32:
    dest.Type = "number"
    dest.Format = "float"
    return
  case reflect.Float64:
    dest.Type = "number"
    dest.Format = "double"
    return
  case reflect.Bool:
    dest.Type = "boolean"
    dest.Format = "boolean"
    return
  case reflect.Array, reflect.Slice:
    dest.Type = "array"
    itemType := &SwaggerType{}
    b._toSwaggerType(typ.Elem(), itemType)
    dest.Items = itemType
    return
  case reflect.Ptr:
    dest.Optional = true
    b._toSwaggerType(typ.Elem(), dest)
    return
  case reflect.Struct:
    dest.Type = "object"
    dest.Format = "#/definitions/" + b.uniqueStructName(typ)
    b.cachedTypes = append(b.cachedTypes, typ)
    return
  default:
    return
  }
}
