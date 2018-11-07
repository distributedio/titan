package toml

import (
	"go/ast"
	"reflect"
	"strings"
	"unicode"
)

type CfgTag struct {
	Name        string
	Value       string
	Check       string
	Description string
}

// toCamelCase returns a copy of the string s with all Unicode letters mapped to their camel case.
// It will convert to upper case previous letter of '_' and first letter, and remove letter of '_'.
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))
	upper := false
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			result = append(result, unicode.ToUpper(r))
			upper = false
			continue
		}
		result = append(result, r)
	}
	result[0] = unicode.ToUpper(result[0])
	return string(result)
}

const (
	fieldTagName = "cfg"
)

func findField(rv reflect.Value, name string) (field reflect.Value, fieldName string, found bool, tag *CfgTag) {
	switch rv.Kind() {
	case reflect.Struct:
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			ft := rt.Field(i)
			if !ast.IsExported(ft.Name) {
				continue
			}
			if tag := extractTag(ft.Tag.Get(fieldTagName)); tag != nil && tag.Name == name {
				return rv.Field(i), ft.Name, true, tag
			}
		}
		for _, name := range []string{
			strings.Title(name),
			toCamelCase(name),
			strings.ToUpper(name),
		} {
			if field := rv.FieldByName(name); field.IsValid() {
				if ft, ok := rt.FieldByName(name); ok {
					tag := extractTag(ft.Tag.Get(fieldTagName))
					return field, name, true, tag
				}
				return field, name, true, nil
			}
		}
	case reflect.Map:
		return reflect.New(rv.Type().Elem()).Elem(), name, true, nil
	}
	return field, "", false, nil
}

func extractTag(tag string) *CfgTag {
	tags := strings.SplitN(tag, ";", 4)
	cfg := &CfgTag{}
	switch c := len(tags); c {
	case 1:
		cfg.Name = strings.TrimSpace(tags[0])
	case 2:
		cfg.Name = strings.TrimSpace(tags[0])
		cfg.Value = strings.TrimSpace(tags[1])
	case 3:
		cfg.Name = strings.TrimSpace(tags[0])
		cfg.Value = strings.TrimSpace(tags[1])
		cfg.Check = strings.TrimSpace(tags[2])
	case 4:
		cfg.Name = strings.TrimSpace(tags[0])
		cfg.Value = strings.TrimSpace(tags[1])
		cfg.Check = strings.TrimSpace(tags[2])
		cfg.Description = strings.TrimSpace(tags[3])
	default:
		return nil
	}
	return cfg
}

func tableName(prefix, name string) string {
	if prefix != "" {
		return prefix + string(tableSeparator) + name
	}
	return name
}
