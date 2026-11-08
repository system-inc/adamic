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
func TestJsonCheckedIrGobRoundTrip(test *testing.T) {
	test.Parallel()
	ir.RegisterGob()
	var paths []string
	for _, pattern := range []string{
		"json_decode_*.a",
		"json_encode_*.a",
		"json_schema_names.a",
		"json_encode_refused/*.a",
	} {
		matches, globError := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata", pattern))
		if globError != nil {
			test.Fatal(globError)
		}
		paths = append(paths, matches...)
	}
	if len(paths) != 30 {
		test.Fatalf("want all 30 JSON fixtures, got %d", len(paths))
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			source, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			program, lowerError := lower.Lower(context.Background(), source)
			if jsonDecodeEnumerationRefused(filepath.Base(path)) {
				assertJsonDecodeEnumerationRefusal(test, lowerError)
				return
			}
			if strings.Contains(filepath.ToSlash(path), "/json_encode_refused/") {
				expected := "encodeJson of a class instance is not yet supported"
				if filepath.Base(path) == "getter.a" {
					expected = "Object enumeration with getter or accessor descriptors"
				}
				if lowerError == nil || !strings.Contains(lowerError.Error(), expected) {
					test.Fatalf("want %s, got %v", expected, lowerError)
				}
				return
			}
			if lowerError != nil {
				test.Fatal(lowerError)
			}
			var data bytes.Buffer
			if encodeError := gob.NewEncoder(&data).Encode(program); encodeError != nil {
				test.Fatalf("encode checked IR: %v", encodeError)
			}
			var decoded ir.Program
			if decodeError := gob.NewDecoder(&data).Decode(&decoded); decodeError != nil {
				test.Fatalf("decode checked IR: %v", decodeError)
			}
			if len(program.Functions) != len(decoded.Functions) {
				test.Fatal("function count changed")
			}
			for index, function := range program.Functions {
				if !reflect.DeepEqual(function.Boundary, decoded.Functions[index].Boundary) {
					test.Fatalf("boundary metadata changed for %s", function.Name)
				}
			}
			if !reflect.DeepEqual(program.BoundaryStrings.Values, decoded.BoundaryStrings.Values) {
				test.Fatal("boundary literals changed")
			}
			if native.C(program) != native.C(&decoded) {
				test.Fatal("native emission changed after gob round trip")
			}
			if javascript.JavaScript(program) != javascript.JavaScript(&decoded) {
				test.Fatal("JS emission changed after gob round trip")
			}
		})
	}
}
