package generator

import (
	"strings"
)

func FieldGoName(name string) string {

	if len(name) == 0 {
		return name
	}

	return strings.ToUpper(
		name[:1],
	) + name[1:]
}
func FieldGoType(field Field) string {

	t := field.Type

	if field.Type == "float" {
		t = "float32"
	} else if field.Type == "double" {
		t = "float64"
	}

	if field.Repeated {

		t = "[]" + t

	}

	return t
}
func FieldJSONName(field Field) string {

	if field.JsonName != "" {

		return field.JsonName

	}

	return field.Name
}
