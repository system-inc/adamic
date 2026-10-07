package rule_testing

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"reflect"
	"sync"
	"unicode/utf16"
)

var waveFixtureLock sync.Mutex

func waveOption(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && v.IsNil() {
		return nil
	}
	if v.CanInterface() {
		if p, ok := v.Interface().(interface{ Source() string }); ok {
			return p.Source()
		}
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		return waveOption(v.Elem())
	case reflect.Struct:
		out := map[string]any{}
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanInterface() {
				out[v.Type().Field(i).Name] = waveOption(v.Field(i))
			}
		}
		return out
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil
		}
		out := []any{}
		for i := 0; i < v.Len(); i++ {
			out = append(out, waveOption(v.Index(i)))
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		out := map[string]any{}
		it := v.MapRange()
		for it.Next() {
			out[it.Key().String()] = waveOption(it.Value())
		}
		return out
	}
	if v.CanInterface() {
		return v.Interface()
	}
	return nil
}
func waveFixtureCapture(subject rule.Rule, file, source string, options any, findings []rule.Diagnostic) {
	path := os.Getenv("ADAMIC_REGEX_FIXTURES")
	if path == "" || (subject.Name != "no-warning-comments" && subject.Name != "id-length" && subject.Name != "no-inline-comments") {
		return
	}
	out := []map[string]any{}
	for _, d := range findings {
		out = append(out, map[string]any{"start": len(utf16.Encode([]rune(source[:d.Range.Pos()]))), "end": len(utf16.Encode([]rune(source[:d.Range.End()]))), "id": d.Message.Id, "message": d.Message.Description})
	}
	row := map[string]any{"rule": subject.Name, "source": source, "file": file, "options": waveOption(reflect.ValueOf(options)), "findings": out}
	waveFixtureLock.Lock()
	defer waveFixtureLock.Unlock()
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e := json.NewEncoder(f).Encode(row); e != nil {
		panic(e)
	}
}
