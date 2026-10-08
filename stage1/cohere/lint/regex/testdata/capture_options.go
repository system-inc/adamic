package rule_testing

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
)

// Keep encoding/json's tags, custom marshalers, omissions and numeric precision.
// Only a compiled matcher's JSON leaf changes: its source replaces private state.
func marshalCapturedOptions(options any) ([]byte, error) {
	original, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(original))
	decoder.UseNumber()
	var encoded any
	if err := decoder.Decode(&encoded); err != nil {
		return nil, err
	}
	return json.Marshal(capturePatternSources(reflect.ValueOf(options), encoded))
}

func capturePatternSources(value reflect.Value, encoded any) any {
	if !value.IsValid() {
		return encoded
	}
	if (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) && value.IsNil() {
		return encoded
	}
	if value.CanInterface() {
		if pattern, ok := value.Interface().(interface{ Source() string }); ok {
			return pattern.Source()
		}
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		return capturePatternSources(value.Elem(), encoded)
	case reflect.Slice, reflect.Array:
		if elements, ok := encoded.([]any); ok {
			for i := range elements {
				elements[i] = capturePatternSources(value.Index(i), elements[i])
			}
		}
	case reflect.Map:
		if fields, ok := encoded.(map[string]any); ok && value.Type().Key().Kind() == reflect.String {
			for _, key := range value.MapKeys() {
				if item, exists := fields[key.String()]; exists {
					fields[key.String()] = capturePatternSources(value.MapIndex(key), item)
				}
			}
		}
	case reflect.Struct:
		if fields, ok := encoded.(map[string]any); ok {
			for i := 0; i < value.NumField(); i++ {
				field := value.Type().Field(i)
				if field.PkgPath != "" {
					continue
				}
				name := strings.Split(field.Tag.Get("json"), ",")[0]
				if name == "-" {
					continue
				}
				if name == "" {
					if field.Anonymous {
						capturePatternSources(value.Field(i), fields)
						continue
					}
					name = field.Name
				}
				if item, exists := fields[name]; exists {
					fields[name] = capturePatternSources(value.Field(i), item)
				}
			}
		}
	}
	return encoded
}
