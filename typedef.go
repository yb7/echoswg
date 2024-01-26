package echoswg

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var GlobalTypeDefBuilder = NewTypeDefBuilder()

type TypeDefBuilder struct {
	cachedTypes       []reflect.Type
	position          int
	StructDefinitions map[string]map[string]interface{}
	typeNames         map[reflect.Type]string
	//anonymousTypes map[reflect.Type]string
}

func NewTypeDefBuilder() *TypeDefBuilder {
	return &TypeDefBuilder{
		cachedTypes:       make([]reflect.Type, 0),
		position:          0,
		StructDefinitions: make(map[string]map[string]interface{}),
		typeNames:         make(map[reflect.Type]string),
	}
}

func (b *TypeDefBuilder) Build(typ reflect.Type, tag reflect.StructTag) *SwaggerType {
	swaggerType := b.ToSwaggerType(typ, tag)

	for b.position < len(b.cachedTypes) {
		pendingType := b.cachedTypes[b.position]
		typeName := b.uniqueStructName(pendingType)
		if _, ok := b.StructDefinitions[typeName]; !ok {
			b.StructDefinitions[typeName] = propertiesOfEntity(pendingType)
		}
		b.position += 1
	}
	return swaggerType
}

func nameInJsonTag(tag reflect.StructTag) string {
	return strings.Split(tag.Get("json"), ",")[0]
}
func propertiesOfEntity(bodyType reflect.Type) map[string]interface{} {
	//isT := bodyType.String() == "util.PagedData[*flashnews.cn/systemctl/service.NewsSourceVo]"
	//fmt.Printf("propertiesOfEntity: %s\n", bodyType)
	properties := make(map[string]interface{})
	requiredFields := make([]string, 0)
	inspectStructType(bodyType, properties, &requiredFields)
	//for i := 0; i < bodyType.NumField(); i++ {
	//	field := bodyType.Field(i)
	//	if field.Anonymous {
	//		inspectAnonymousField(field, properties, requiredFields)
	//		continue
	//	}
	//	propertyName := field.Name
	//	propertyJsonName := nameInJsonTag(field.Tag) //strings.Split(field.Tag.Get("json"), ",")[0]
	//	if len(propertyJsonName) > 0 {
	//		propertyName = propertyJsonName
	//	}
	//	if isT {
	//		fmt.Printf("field [%s]  Anonymous = %v\n", field.Name, field.Anonymous)
	//	}
	//
	//	swaggerType := GlobalTypeDefBuilder.ToSwaggerType(field.Type, field.Tag)
	//
	//	if !swaggerType.Optional {
	//		requiredFields = append(requiredFields, propertyName)
	//	}
	//
	//	propertyJson := swaggerType.ToSwaggerJSON()
	//
	//	description := field.Tag.Get("desc")
	//	description = strings.TrimSpace(description)
	//	if len(description) > 0 {
	//		propertyJson["description"] = description
	//	}
	//
	//	properties[propertyName] = propertyJson
	//}
	return map[string]interface{}{
		"type":       "object",
		"required":   requiredFields,
		"properties": properties,
	}
}
func inspectStructType(inputType reflect.Type, properties map[string]interface{}, requiredFields *[]string) {
	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		if field.Anonymous {
			inspectStructType(field.Type, properties, requiredFields)
			continue
		}
		propertyName := field.Name
		propertyJsonName := nameInJsonTag(field.Tag) //strings.Split(field.Tag.Get("json"), ",")[0]
		if len(propertyJsonName) > 0 {
			propertyName = propertyJsonName
		}
		swaggerType := GlobalTypeDefBuilder.ToSwaggerType(field.Type, field.Tag)

		if !swaggerType.Optional {
			*requiredFields = append(*requiredFields, propertyName)
		}

		propertyJson := swaggerType.ToSwaggerJSON()

		description := field.Tag.Get("desc")
		description = strings.TrimSpace(description)
		if len(description) > 0 {
			propertyJson["description"] = description
		}

		properties[propertyName] = propertyJson
	}
}

func (b *TypeDefBuilder) uniqueStructName(typ reflect.Type) string {
	if existed, ok := b.typeNames[typ]; ok {
		return existed
	}

	typeName := typ.Name()
	if strings.ContainsAny(typeName, "[]*./") {
		//fmt.Printf("type name = %s\n", typeName)

		// remove packages
		// PagedData[*flashnews.cn/systemctl/service.NewsSourceVo] => PagedData[*service.NewsSourceVo]
		m1 := regexp.MustCompile(`([a-zA-Z0-9\-\.]+/)+`)
		typeName = m1.ReplaceAllString(typeName, "")
		//fmt.Printf("    1 = %s\n", typeName)
		// remove packages again
		// PagedData[*service.NewsSourceVo] => PagedData[*NewsSourceVo]
		m2 := regexp.MustCompile(`([a-zA-Z0-9\-]+\.)+`)
		typeName = m2.ReplaceAllString(typeName, "")
		//fmt.Printf("    2 = %s\n", typeName)
		typeName = strings.ReplaceAll(typeName, "[", "")
		typeName = strings.ReplaceAll(typeName, "]", "")
		typeName = strings.ReplaceAll(typeName, "*", "")
		//fmt.Printf("    3 = %s\n", typeName)
	}

	if len(typeName) == 0 {
		typeName = "anonymous"
	}

	typeName = nextAvailableName(b.typeNames, typeName)

	b.typeNames[typ] = typeName
	return typeName
}

func nextAvailableName(exists map[reflect.Type]string, requestName string) string {
	relatedNames := make(map[string]bool)
	for _, name := range exists {
		if strings.HasPrefix(name, requestName) {
			relatedNames[name] = true
		}
	}
	for i := 0; i < 1000; i++ {
		nextName := requestName
		if i > 0 {
			nextName = fmt.Sprintf("%s%03d", requestName, i)
		}
		_, ok := relatedNames[nextName]
		if !ok {
			return nextName
		}
	}
	return strconv.Itoa(int(time.Now().UnixMicro()))
}

type SwaggerType struct {
	Optional bool
	Type     string
	Format   string
	Items    *SwaggerType
	Ext      map[string]any
}

func (t *SwaggerType) String() string {
	if t == nil {
		return ""
	}
	switch t.Type {
	case "array":
		return fmt.Sprintf("type: array, items: [%s]", t.Items.String())
	case "object":
		return fmt.Sprintf("$ref: %s", t.Format)
	default:
		return fmt.Sprintf("optional: %t, type: %s, format: %s", t.Optional, t.Type, t.Format)
	}
}

func (t *SwaggerType) ToSwaggerJSON() map[string]interface{} {
	switch t.Type {
	case "array":
		return mergeMap(t.Ext, map[string]interface{}{
			"type":  "array",
			"items": t.Items.ToSwaggerJSON(),
		})
	case "object":
		return map[string]interface{}{
			"$ref": t.Format,
		}
	default:
		return mergeMap(t.Ext, map[string]interface{}{
			"type":   t.Type,
			"format": t.Format,
		})
	}
}
func (b *TypeDefBuilder) ToSwaggerType(typ reflect.Type, tag reflect.StructTag) *SwaggerType {
	v := &SwaggerType{}
	b._toSwaggerType(typ, tag, v)
	return v
}

var TimeType = reflect.TypeOf((*time.Time)(nil)).Elem()

func (b *TypeDefBuilder) _toSwaggerType(typ reflect.Type, tag reflect.StructTag, dest *SwaggerType) {

	if typ == TimeType {
		dest.Type = "string"
		dest.Format = "string"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	}

	validateTag := tag.Get("validate")
	tags := validateTagToMap(validateTag)
	hasRequiredTag := false
	for name, _ := range tags {
		if name == "required" {
			hasRequiredTag = true
			break
		}
	}

	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		dest.Type = "integer"
		dest.Format = "int32"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.Int64, reflect.Uint64:
		dest.Type = "integer"
		dest.Format = "int64"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.String:
		dest.Type = "string"
		dest.Format = "string"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.Float32:
		dest.Type = "number"
		dest.Format = "float"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.Float64:
		dest.Type = "number"
		dest.Format = "double"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.Bool:
		dest.Type = "boolean"
		dest.Format = "boolean"
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.Array, reflect.Slice:
		dest.Type = "array"
		itemType := &SwaggerType{}
		b._toSwaggerType(typ.Elem(), "", itemType)
		dest.Items = itemType
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		if !hasRequiredTag {
			dest.Optional = true
		}
		return
	case reflect.Ptr:
		dest.Optional = true
		b._toSwaggerType(typ.Elem(), "", dest)
		dest.Ext = parseValidateTag(dest.Type, tag.Get("validate"))
		return
	case reflect.Struct:
		dest.Type = "object"
		dest.Format = "#/components/schemas/" + b.uniqueStructName(typ)
		b.cachedTypes = append(b.cachedTypes, typ)
		//fmt.Printf("add type to cache: %s", typ)
		return
	default:
		return
	}
}
