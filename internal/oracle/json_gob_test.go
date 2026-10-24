package oracle

import (
	"bytes"
	"context"
	"encoding/gob"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Checked IR is a checkpoint format used by estree and lint. Exercise the whole
// JSON corpus, including class receivers with no boundary schema. Refusal cases
// have no checked IR; their existing refusal contracts must still hold.
func TestJSONCheckedIRGobRoundTrip(t *testing.T) {
	t.Parallel()
	registered := map[reflect.Type]bool{}
	var register func(reflect.Value)
	register = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return
			}
			concrete := v.Elem()
			if !registered[concrete.Type()] {
				gob.Register(concrete.Interface())
				registered[concrete.Type()] = true
			}
			register(concrete)
		case reflect.Pointer:
			if !v.IsNil() {
				register(v.Elem())
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					register(v.Field(i))
				}
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				register(v.Index(i))
			}
		case reflect.Map:
			entries := v.MapRange()
			for entries.Next() {
				register(entries.Key())
				register(entries.Value())
			}
		}
	}
	var paths []string
	for _, pattern := range []string{"json_decode_*.a", "json_encode_*.a", "json_schema_names.a", "json_encode_refused/*.a"} {
		matches, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata", pattern))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	if len(paths) != 29 {
		t.Fatalf("want all 29 JSON fixtures, got %d", len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), source)
			if jsonDecodeEnumerationRefused(filepath.Base(path)) {
				assertJSONDecodeEnumerationRefusal(t, err)
				return
			}
			if strings.Contains(filepath.ToSlash(path), "/json_encode_refused/") {
				expected := "encodeJson of a class instance is not yet supported"
				if filepath.Base(path) == "getter.a" {
					expected = "Object enumeration with getter or accessor descriptors"
				}
				if err == nil || !strings.Contains(err.Error(), expected) {
					t.Fatalf("want %s, got %v", expected, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			register(reflect.ValueOf(program))
			var data bytes.Buffer
			if err := gob.NewEncoder(&data).Encode(program); err != nil {
				t.Fatalf("encode checked IR: %v", err)
			}
			var decoded ir.Program
			if err := gob.NewDecoder(&data).Decode(&decoded); err != nil {
				t.Fatalf("decode checked IR: %v", err)
			}
			if len(program.Functions) != len(decoded.Functions) {
				t.Fatal("function count changed")
			}
			for i, function := range program.Functions {
				if !reflect.DeepEqual(function.Boundary, decoded.Functions[i].Boundary) {
					t.Fatalf("boundary metadata changed for %s", function.Name)
				}
			}
			if !reflect.DeepEqual(program.BoundaryStrings.Values, decoded.BoundaryStrings.Values) {
				t.Fatal("boundary literals changed")
			}
			if native.C(program) != native.C(&decoded) {
				t.Fatal("native emission changed after gob round trip")
			}
			if javascript.JavaScript(program) != javascript.JavaScript(&decoded) {
				t.Fatal("JS emission changed after gob round trip")
			}
		})
	}
}
