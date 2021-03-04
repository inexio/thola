package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func toHumanReadable(value reflect.Value, insertion int) string {
	kind := value.Kind()

	switch kind {
	case reflect.Struct:
		output := "\n"
		for i := 0; i < value.NumField(); i++ {
			fieldValue := toHumanReadable(value.Field(i), insertion+1)
			if strings.TrimSpace(fieldValue) == "" {
				continue
			}
			output += strings.Repeat("  ", insertion)
			output += value.Type().Field(i).Name + ": "
			output += fieldValue
			output += "\n"
		}
		return output
	case reflect.Slice:
		if value.IsNil() {
			return ""
		}
		output := "[" + strconv.Itoa(value.Len()) + "] "
		for i := 0; i < value.Len(); i++ {
			output += toHumanReadable(value.Index(i), insertion+1)
			output += " "
		}
		output += "\n"
		return output
	case reflect.Map:
		output := "(" + strconv.Itoa(value.Len()) + ") \n"
		for _, key := range value.MapKeys() {
			output += strings.Repeat("  ", insertion)
			output += key.String() + ": "
			output += toHumanReadable(value.MapIndex(key), insertion+1)
			output += "\n"
		}
		return output
	case reflect.String:
		return value.String()
	case reflect.Int:
		return strconv.Itoa(int(value.Int()))
	case reflect.Uint:
		return strconv.Itoa(int(value.Uint()))
	case reflect.Uint64:
		return strconv.Itoa(int(value.Uint()))
	case reflect.Float64:
		return fmt.Sprintf("%f", value.Float())
	case reflect.Ptr:
		if value.IsNil() {
			return ""
		}
		return toHumanReadable(reflect.Indirect(value), insertion)
	case reflect.Interface:
		return toHumanReadable(reflect.ValueOf(value.Interface()), insertion)
	default:
		if !value.IsValid() {
			return ""
		} else {
			return fmt.Sprint(value.Interface())
		}
	}
}
