package validate

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Struct validates exported fields of a struct according to `validate` struct tags.
// Supported tags (comma-separated):
//
//	required   – field must not be the zero value
//	min=N      – for strings/slices/maps: length >= N; for numbers: value >= N
//	max=N      – for strings/slices/maps: length <= N; for numbers: value <= N
//	gt=N       – for numbers: value > N
//	email      – string must contain exactly one '@'
func Struct(v interface{}) error {
	if v == nil {
		return fmt.Errorf("validate: value is nil")
	}

	val := reflect.ValueOf(v)
	// unwrap pointer
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("validate: pointer is nil")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validate: expected struct, got %s", val.Kind())
	}

	return validateStruct(val)
}

func validateStruct(val reflect.Value) error {
	typ := val.Type()

	for i := range typ.NumField() {
		field := typ.Field(i)
		fval := val.Field(i)

		// skip unexported
		if !field.IsExported() {
			continue
		}

		tag, ok := field.Tag.Lookup("validate")
		if !ok || tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			if err := applyRule(field.Name, fval, rule); err != nil {
				return err
			}
		}

		// recurse into nested structs
		concrete := fval
		for concrete.Kind() == reflect.Ptr {
			if concrete.IsNil() {
				break
			}
			concrete = concrete.Elem()
		}
		if concrete.Kind() == reflect.Struct {
			if err := validateStruct(concrete); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyRule(name string, fval reflect.Value, rule string) error {
	switch {
	case rule == "required":
		return checkRequired(name, fval)

	case strings.HasPrefix(rule, "min="):
		n, err := strconv.ParseFloat(strings.TrimPrefix(rule, "min="), 64)
		if err != nil {
			return fmt.Errorf("validate: field %s invalid min tag: %w", name, err)
		}
		return checkMin(name, fval, n)

	case strings.HasPrefix(rule, "max="):
		n, err := strconv.ParseFloat(strings.TrimPrefix(rule, "max="), 64)
		if err != nil {
			return fmt.Errorf("validate: field %s invalid max tag: %w", name, err)
		}
		return checkMax(name, fval, n)

	case strings.HasPrefix(rule, "gt="):
		n, err := strconv.ParseFloat(strings.TrimPrefix(rule, "gt="), 64)
		if err != nil {
			return fmt.Errorf("validate: field %s invalid gt tag: %w", name, err)
		}
		return checkGt(name, fval, n)

	case rule == "email":
		return checkEmail(name, fval)
	}

	// unknown rules are silently ignored to allow forward-compatibility
	return nil
}

func checkRequired(name string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return fmt.Errorf("validate: field %s is required", name)
		}
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return fmt.Errorf("validate: field %s is required (nil)", name)
		}
	case reflect.Slice, reflect.Map:
		if v.IsNil() || v.Len() == 0 {
			return fmt.Errorf("validate: field %s is required (empty)", name)
		}
	default:
		if v.IsZero() {
			return fmt.Errorf("validate: field %s is required", name)
		}
	}
	return nil
}

func checkMin(name string, v reflect.Value, min float64) error {
	switch v.Kind() {
	case reflect.String:
		if float64(len([]rune(v.String()))) < min {
			return fmt.Errorf("validate: field %s must have length >= %.0f", name, min)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if float64(v.Len()) < min {
			return fmt.Errorf("validate: field %s must have length >= %.0f", name, min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) < min {
			return fmt.Errorf("validate: field %s must be >= %.0f", name, min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) < min {
			return fmt.Errorf("validate: field %s must be >= %.0f", name, min)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < min {
			return fmt.Errorf("validate: field %s must be >= %g", name, min)
		}
	}
	return nil
}

func checkMax(name string, v reflect.Value, max float64) error {
	switch v.Kind() {
	case reflect.String:
		if float64(len([]rune(v.String()))) > max {
			return fmt.Errorf("validate: field %s must have length <= %.0f", name, max)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if float64(v.Len()) > max {
			return fmt.Errorf("validate: field %s must have length <= %.0f", name, max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) > max {
			return fmt.Errorf("validate: field %s must be <= %.0f", name, max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) > max {
			return fmt.Errorf("validate: field %s must be <= %.0f", name, max)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > max {
			return fmt.Errorf("validate: field %s must be <= %g", name, max)
		}
	}
	return nil
}

func checkGt(name string, v reflect.Value, threshold float64) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) <= threshold {
			return fmt.Errorf("validate: field %s must be > %.0f", name, threshold)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) <= threshold {
			return fmt.Errorf("validate: field %s must be > %.0f", name, threshold)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() <= threshold {
			return fmt.Errorf("validate: field %s must be > %g", name, threshold)
		}
	}
	return nil
}

func checkEmail(name string, v reflect.Value) error {
	if v.Kind() != reflect.String {
		return nil
	}
	s := v.String()
	at := strings.Count(s, "@")
	if at != 1 {
		return fmt.Errorf("validate: field %s must be a valid email", name)
	}
	parts := strings.Split(s, "@")
	if parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return fmt.Errorf("validate: field %s must be a valid email", name)
	}
	return nil
}
