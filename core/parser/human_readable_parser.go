package parser

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func toHumanReadable(value reflect.Value, insertion int) []byte {
	var outputString []byte

	kind := value.Kind()

	switch kind {
	case reflect.Struct:
		valueField := ""
		if insertion > 0 {
			valueField += "\n"
		}
		for index := 0; index < value.NumField(); index++ {
			valueField += strings.Repeat("  ", insertion)
			valueField += value.Type().Field(index).Name + ": "
			arg := toHumanReadable(value.Field(index), insertion+1)
			if arg == nil || string(arg) == "null" {
				valueField = ""
				continue
			}
			outputString = append(outputString, valueField...)
			outputString = append(outputString, arg...)
			fKind := value.Field(index).Type().Kind()
			if fKind == reflect.String || fKind == reflect.Int || fKind == reflect.Float64 || fKind == reflect.Ptr {
				outputString = append(outputString, "\n"...)
			}
			valueField = ""
		}
	case reflect.Slice:
		outputString = append(outputString, "["+strconv.Itoa(value.Len())+"] "...)
		for index := 0; index < value.Len(); index++ {
			arg := toHumanReadable(value.Index(index), insertion+1)
			outputString = append(outputString, arg...)
			outputString = append(outputString, " "...)
		}
		outputString = append(outputString, "\n"...)
	case reflect.Map:
		outputString = append(outputString, "("+strconv.Itoa(value.Len())+") \n"...)
		for _, key := range value.MapKeys() {
			outputString = append(outputString, strings.Repeat("  ", insertion)...)
			outputString = append(outputString, key.String()+": "...)
			arg := toHumanReadable(value.MapIndex(key), insertion+1)
			outputString = append(outputString, arg...)
			outputString = append(outputString, "\n"...)
		}
	case reflect.String:
		fieldValue := value.String()
		outputString = append(outputString, fieldValue...)
	case reflect.Int:
		fieldValue := strconv.Itoa(int(value.Int()))
		outputString = append(outputString, fieldValue...)
	case reflect.Uint:
		fieldValue := strconv.Itoa(int(value.Uint()))
		outputString = append(outputString, fieldValue...)
	case reflect.Uint64:
		fieldValue := strconv.Itoa(int(value.Uint()))
		outputString = append(outputString, fieldValue...)
	case reflect.Float64:
		fieldValue := strconv.FormatFloat(value.Float(), 'f', -1, 64)
		outputString = append(outputString, fieldValue...)
	case reflect.Ptr:
		if value.IsNil() {
			fieldValue := "null"
			outputString = append(outputString, fieldValue...)
			return outputString
		}
		arg := toHumanReadable(reflect.Indirect(value), insertion)
		outputString = append(outputString, arg...)
	case reflect.Interface:
		ifValue := reflect.ValueOf(value.Interface())
		outputString = append(outputString, toHumanReadable(ifValue, insertion)...)
	default:
		if !value.IsValid() {
			outputString = append(outputString, " "...)

		} else {
			outputString = append(outputString, fmt.Sprint(value.Interface())...)
		}
	}

	return bytes.TrimSpace(outputString)
}
