package sqlbuilder

import (
	"fmt"
	"reflect"
	"strings"
)

// StructToMap converts a struct to a map using db tags
// Supports omitempty and skipping zero values
func StructToMap(model interface{}, skipZero bool) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get db tag
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Parse tag options (e.g., "name,omitempty")
		tagParts := strings.Split(dbTag, ",")
		columnName := tagParts[0]

		// Check for omitempty
		omitEmpty := false
		for _, opt := range tagParts[1:] {
			if opt == "omitempty" {
				omitEmpty = true
			}
		}

		// Skip zero values if requested
		if skipZero && isZeroValue(value) {
			continue
		}

		// Skip if omitempty and value is zero
		if omitEmpty && isZeroValue(value) {
			continue
		}

		result[columnName] = value.Interface()
	}

	return result
}

// StructToMapExclude converts struct to map but excludes specified fields
func StructToMapExclude(model interface{}, excludeFields ...string) map[string]interface{} {
	result := StructToMap(model, false)

	for _, field := range excludeFields {
		delete(result, field)
	}

	return result
}

// StructToMapInclude converts struct to map but only includes specified fields
func StructToMapInclude(model interface{}, includeFields ...string) map[string]interface{} {
	allFields := StructToMap(model, false)
	result := make(map[string]interface{})

	for _, field := range includeFields {
		if val, ok := allFields[field]; ok {
			result[field] = val
		}
	}

	return result
}

// GetColumns returns all column names from a struct using db tags
func GetColumns(model interface{}) []string {
	var columns []string

	v := reflect.TypeOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return columns
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Get column name (before comma if exists)
		tagParts := strings.Split(dbTag, ",")
		columnName := tagParts[0]
		columns = append(columns, columnName)
	}

	return columns
}

// GetColumnsExclude returns column names excluding specified fields
func GetColumnsExclude(model interface{}, excludeFields ...string) []string {
	allColumns := GetColumns(model)
	excludeMap := make(map[string]bool)

	for _, field := range excludeFields {
		excludeMap[field] = true
	}

	result := []string{}
	for _, col := range allColumns {
		if !excludeMap[col] {
			result = append(result, col)
		}
	}

	return result
}

// GetColumnValue gets the value of a specific column from a struct
func GetColumnValue(model interface{}, columnName string) (interface{}, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag == "" {
			continue
		}

		tagParts := strings.Split(dbTag, ",")
		if tagParts[0] == columnName {
			return v.Field(i).Interface(), nil
		}
	}

	return nil, fmt.Errorf("column %s not found", columnName)
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Struct:
		// For time.Time and other structs, check if it's zero
		return v.IsZero()
	default:
		return false
	}
}

// BuildSelectColumns builds SELECT column list from struct
func BuildSelectColumns(model interface{}, tableAlias ...string) string {
	columns := GetColumns(model)

	if len(tableAlias) > 0 && tableAlias[0] != "" {
		alias := tableAlias[0]
		for i, col := range columns {
			columns[i] = fmt.Sprintf("%s.%s", alias, col)
		}
	}

	return strings.Join(columns, ", ")
}
