package server_utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

import (
	msconfig "mockserver/config"
)

// ValidateJSONSchema performs a recursive validation of data against a JSON schema.
// It supports structural validation (Objects, Arrays) and constraints (Min/Max, Regex, Enums).
func ValidateJSONSchema(schema *msconfig.JSONSchema, data interface{}, path string) error {
	if schema == nil {
		return nil
	}

	if path == "" {
		path = "root"
	}

	if err := validateType(schema.Type, data, path); err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	// Constraint Validation based on Type
	switch schema.Type {
	case "object":
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s: expected object", path)
		}
		return validateObject(schema, dataMap, path)

	case "array":
		dataArr, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("%s: expected array", path)
		}
		return validateArray(schema, dataArr, path)

	case "string":
		val := data.(string)
		if schema.MinLength != nil && len(val) < *schema.MinLength {
			return fmt.Errorf("%s: length must be >= %d", path, *schema.MinLength)
		}
		if schema.MaxLength != nil && len(val) > *schema.MaxLength {
			return fmt.Errorf("%s: length must be <= %d", path, *schema.MaxLength)
		}
		if schema.Pattern != "" {
			matched, _ := regexp.MatchString(schema.Pattern, val)
			if !matched {
				return fmt.Errorf("%s: value does not match pattern '%s'", path, schema.Pattern)
			}
		}
		if len(schema.Enum) > 0 && !contains(schema.Enum, val) {
			return fmt.Errorf("%s: invalid value '%s'. allowed: %v", path, val, schema.Enum)
		}

	case "integer", "number":
		val, _ := data.(float64)
		if schema.Minimum != nil && val < *schema.Minimum {
			return fmt.Errorf("%s: must be >= %f", path, *schema.Minimum)
		}
		if schema.Maximum != nil && val > *schema.Maximum {
			return fmt.Errorf("%s: must be <= %f", path, *schema.Maximum)
		}
	}

	return nil
}

// validateObject enforces 'required' fields and recursively validates nested properties.
func validateObject(schema *msconfig.JSONSchema, data map[string]interface{}, parentPath string) error {
	// Required Fields Check
	for _, reqField := range schema.Required {
		if _, exists := data[reqField]; !exists {
			return fmt.Errorf("%s: missing required field '%s'", parentPath, reqField)
		}
	}

	//Property Validation (Recursive)
	for key, val := range data {
		propSchema, defined := schema.Properties[key]
		if !defined {
			if !schema.AdditionalProperties {
			}
			continue
		}

		if err := ValidateJSONSchema(propSchema, val, parentPath+"."+key); err != nil {
			return err
		}
	}
	return nil
}

// validateArray iterates over the slice and validates each item against the 'items' schema.
func validateArray(schema *msconfig.JSONSchema, data []interface{}, parentPath string) error {
	if schema.Items == nil {
		return nil
	}
	for i, item := range data {
		if err := ValidateJSONSchema(schema.Items, item, fmt.Sprintf("%s[%d]", parentPath, i)); err != nil {
			return err
		}
	}
	return nil
}

func validateType(expectedType string, data interface{}, path string) error {
	if expectedType == "" {
		return nil
	}

	gotType := ""
	switch val := data.(type) {

	case string:
		if expectedType == "integer" || expectedType == "number" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				if expectedType == "integer" && f != float64(int64(f)) {
					return fmt.Errorf("%s: expected integer (whole number), got float string '%s'", path, val)
				}
				return nil
			}
		}

		if expectedType == "boolean" {
			if val == "true" || val == "false" {
				return nil
			}
		}

		gotType = "string"

	case float64, int, int64:
		if expectedType == "integer" {
			var f float64
			switch v := data.(type) {
			case float64:
				f = v
			case int:
				f = float64(v)
			case int64:
				f = float64(v)
			}

			if f != float64(int64(f)) {
				return fmt.Errorf("%s: expected integer, got float", path)
			}
			return nil
		}
		gotType = expectedType // accept "number" or "integer"

	case bool:
		gotType = "boolean"
	case map[string]interface{}:
		gotType = "object"
	case []interface{}:
		gotType = "array"
	case nil:
		gotType = "null"
	default:
		gotType = fmt.Sprintf("%T", data)
	}

	// Loose check: 'number' type accepts both integer and float
	if expectedType == "number" && (gotType == "integer" || fmt.Sprintf("%T", data) == "float64") {
		return nil
	}

	if gotType != expectedType {
		return fmt.Errorf("%s: expected type '%s', got '%s'", path, expectedType, gotType)
	}
	return nil
}

func contains(slice []interface{}, val interface{}) bool {
	for _, item := range slice {
		if reflect.DeepEqual(item, val) {
			return true
		}
	}
	return false
}
