package echoswg

import (
	"reflect"
	"testing"
)

func TestDescriptionFromTag(t *testing.T) {
	type sample struct {
		A int `desc:"legacy"`
		B int `jsonschema_description:"primary"`
		C int `desc:"legacy" jsonschema_description:"primary"`
		D int `jsonschema_description:"   spaced   "`
		E int
	}

	cases := []struct {
		field string
		want  string
	}{
		{"A", "legacy"},
		{"B", "primary"},
		{"C", "primary"},
		{"D", "spaced"},
		{"E", ""},
	}

	typ := reflect.TypeOf(sample{})
	for _, c := range cases {
		f, _ := typ.FieldByName(c.field)
		if got := descriptionFromTag(f.Tag); got != c.want {
			t.Fatalf("field %s: got %q, want %q", c.field, got, c.want)
		}
	}
}
