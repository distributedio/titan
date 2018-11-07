package configo

import (
	"github.com/shafreeck/configo/rule"
	"github.com/shafreeck/toml"
	"github.com/shafreeck/toml/ast"

	"fmt"
	goast "go/ast"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	fieldTagName = "cfg"
)

func init() {
	toml.SetValue = fieldValidate
}

func fieldValidate(field string, rv reflect.Value, av ast.Value, tag *toml.CfgTag) error {
	if tag == nil {
		return nil
	}
	val, ok := av.(*ast.String)
	if tag.Check != "" && ok {
		return validate(field, val.Value, tag.Check)
	}
	return nil
}
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func extractTag(tag string) *toml.CfgTag {
	tags := strings.SplitN(tag, ";", 4)
	cfg := &toml.CfgTag{}
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

func validate(key, value string, check string) error {
	r := rule.Rule(check)
	vlds, err := r.Parse()
	if err != nil {
		return err
	}
	for _, vld := range vlds {
		if err := vld.Validate(value); err != nil {
			return fmt.Errorf("validate %s failed, %s does not match rule %q, reason: %v", key, value, check, err)
		}
	}
	return nil
}

//parse a toml array
func unmarshalArray(key, value string, v interface{}) error {
	//construct a valid toml array
	data := key + " = " + value
	if err := toml.Unmarshal([]byte(data), v); err != nil {
		return err
	}
	return nil
}

func applyDefaultValue(fv reflect.Value, ft reflect.StructField, rv reflect.Value, ignoreRequired bool) (err error) {
	tag := extractTag(ft.Tag.Get(fieldTagName))

	//Default value is not supported
	if tag.Value == "required" {
		if ignoreRequired {
			return nil
		}
		return fmt.Errorf("value of %q is required in %v", ft.Name, rv.Type())
	}

	//No default value supplied
	if tag.Value == "" {
		return nil
	}

	//Validate the default value
	//reflect.Slice will be validated by unmarshalArray
	if tag.Check != "" && fv.Kind() != reflect.Slice {
		if err := validate(ft.Name, tag.Value, tag.Check); err != nil {
			return err
		}
	}

	//Set the default value
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		var v int64
		if v, err = strconv.ParseInt(tag.Value, 10, 64); err != nil {
			if fv.Kind() == reflect.Int64 {
				//try to parse a time.Duration
				if d, err := time.ParseDuration(tag.Value); err == nil {
					fv.SetInt(int64(d))
					return nil
				}
			}
			return err
		}
		fv.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		var v uint64
		if v, err = strconv.ParseUint(tag.Value, 10, 64); err != nil {
			return err
		}
		fv.SetUint(v)
	case reflect.Float32, reflect.Float64:
		var v float64
		if v, err = strconv.ParseFloat(tag.Value, 64); err != nil {
			return err
		}
		fv.SetFloat(v)
	case reflect.Bool:
		var v bool
		if v, err = strconv.ParseBool(tag.Value); err != nil {
			return err
		}
		fv.SetBool(v)
	case reflect.String:
		fv.SetString(tag.Value)
	case reflect.Slice:
		v := rv.Addr().Interface()
		if err := unmarshalArray(ft.Name, tag.Value, v); err != nil {
			return err
		}
	default:
		return fmt.Errorf("set default value of type %s is not supported yet", ft.Type)
	}
	return nil
}

//Notice toCamelCase is copied from github.com/naoina/toml
// toCamelCase returns a copy of the string s with all Unicode letters mapped to their camel case.
// It will convert to upper case previous letter of '_' and first letter, and remove letter of '_'.
func toUnderscore(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))

	result = append(result, unicode.ToLower(rune(s[0])))
	for _, r := range s[1:] {
		if unicode.ToUpper(r) == r {
			result = append(result, '_', unicode.ToLower(r))
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

func findField(t *ast.Table, field reflect.StructField) (interface{}, bool) {
	if t == nil {
		return nil, false
	}
	tag := extractTag(field.Tag.Get(fieldTagName))
	if tag != nil && tag.Name != "" {
		if f, found := t.Fields[tag.Name]; found {
			return f, found
		}
		return nil, false
	}

	name := field.Name
	for _, n := range []string{name, strings.ToLower(name), toUnderscore(name)} {
		if f, found := t.Fields[n]; found {
			return f, found
		}
	}
	return nil, false
}

func applyDefault(t *ast.Table, rv reflect.Value, ignoreRequired bool) error {
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type()

	if kind := rt.Kind(); kind == reflect.Struct {
		for i := 0; i < rt.NumField(); i++ {
			ft := rt.Field(i)
			fv := rv.Field(i)
			if !goast.IsExported(ft.Name) {
				continue
			}
			for fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
			}
			if fv.Kind() == reflect.Struct {
				var subt *ast.Table
				var ok bool
				if f, found := findField(t, ft); found {
					subt, ok = f.(*ast.Table)
					//Assgin t back to subt
					//This is becuase the reflect.Struct is emmbed
					// type D struct {
					//    time.Duration
					// }
					// D is a struct , but there is no sub table in conf
					if !ok {
						subt = t
					}
				}

				if err := applyDefault(subt, fv, ignoreRequired); err != nil {
					return err
				}
				continue
			}

			//Maybe array of table
			if fv.IsValid() && !isEmptyValue(fv) && fv.Kind() == reflect.Slice {
				if arrtable, found := findField(t, ft); found {
					arrtable, ok := arrtable.([]*ast.Table)
					if ok {
						for j := 0; j < fv.Len(); j++ {
							ev := fv.Index(j)
							st := arrtable[j]
							if err := applyDefault(st, ev, ignoreRequired); err != nil {
								return err
							}
						}
					}
				}
			}

			if fv.IsValid() && isEmptyValue(fv) {
				if _, found := findField(t, ft); !found {
					if err := applyDefaultValue(fv, ft, rv, ignoreRequired); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

//Unmarshal data into struct v, v shoud be a pointer to struct
func Unmarshal(data []byte, v interface{}) error {
	table, err := toml.Parse(data)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, v); err != nil {
		return err
	}

	if err := applyDefault(table, reflect.ValueOf(v), false); err != nil {
		return err
	}
	return nil
}

//Marshal v to configuration in toml format
func Marshal(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	pv := reflect.New(rv.Type())
	pv.Elem().Set(rv)

	if err := applyDefault(nil, pv, true); err != nil {
		return nil, err
	}
	return toml.Marshal(pv.Interface())
}

//Patch the base using the value from v, the new bytes returned
//combines the base's value and v's default value
func Patch(base []byte, v interface{}) ([]byte, error) {
	//Clone struct v, v shoud not be modified
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	pv := reflect.New(rv.Type())
	pv.Elem().Set(rv)

	nv := pv.Interface()

	//unmarshal base
	table, err := toml.Parse(base)
	if err != nil {
		return nil, err
	}

	if err := toml.UnmarshalTable(table, nv); err != nil {
		return nil, err
	}

	return Marshal(nv)
}
