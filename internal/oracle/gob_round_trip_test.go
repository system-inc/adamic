package oracle

import (
	"bytes"
	"context"
	"encoding/gob"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// gobRoundTripPaths is every fixture stage 0 lowers, once each, from the repository root.
func gobRoundTripPaths() []string {
	var paths []string
	for _, fixture := range fixtures {
		if fixture.lowers && !refusedAdamicNonNullFixture(fixture.path) && !slices.Contains(paths, fixture.path) {
			paths = append(paths, fixture.path)
		}
	}
	return paths
}

// negativeZeros counts the -0 constants a value holds.
func negativeZeros(value reflect.Value) int {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return 0
		}
		if constant, ok := value.Interface().(ir.NumberConstant); ok {
			if constant.Value == 0 && math.Signbit(constant.Value) {
				return 1
			}
			return 0
		}
		return negativeZeros(value.Elem())
	case reflect.Pointer:
		if value.IsNil() {
			return 0
		}
		return negativeZeros(value.Elem())
	case reflect.Struct:
		count := 0
		for index := range value.NumField() {
			if value.Type().Field(index).IsExported() {
				count += negativeZeros(value.Field(index))
			}
		}
		return count
	case reflect.Slice, reflect.Array:
		count := 0
		for index := range value.Len() {
			count += negativeZeros(value.Index(index))
		}
		return count
	case reflect.Map:
		count := 0
		for entries := value.MapRange(); entries.Next(); {
			count += negativeZeros(entries.Key()) + negativeZeros(entries.Value())
		}
		return count
	}
	return 0
}

// gobRoundTrip lowers every fixture, writes the program as gob twice before anything emits it, and holds the bytes
// equal; then emits the program with options, emits a fresh decode of the bytes with the same options, and holds the
// two C byte for byte equal, and every -0 constant the program holds a -0 in the decode. Emit writes into the IR
// (element_borrow.go, async_ordinary.go), so each side is emitted once, from a program nothing emitted before.
func gobRoundTrip(t *testing.T, options native.Options) {
	ir.RegisterGob()
	var holding atomic.Int64
	// The corpus isn't parallel to its test, so its cases have all run when it returns.
	t.Run("corpus", func(t *testing.T) {
		for _, path := range gobRoundTripPaths() {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				source, err := load.Load([]string{filepath.Join(repository, path)})
				if err != nil {
					t.Fatal(err)
				}
				program, err := lower.Lower(context.Background(), source)
				if err != nil {
					t.Fatal(err)
				}
				var first, second bytes.Buffer
				if err := gob.NewEncoder(&first).Encode(program); err != nil {
					t.Fatal(err)
				}
				if err := gob.NewEncoder(&second).Encode(program); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(first.Bytes(), second.Bytes()) {
					t.Fatalf("one program, two gob encodings (%d and %d bytes)", first.Len(), second.Len())
				}
				zeros := negativeZeros(reflect.ValueOf(program))
				want := native.CWithOptions(program, options)
				var decoded ir.Program
				if err := gob.NewDecoder(&first).Decode(&decoded); err != nil {
					t.Fatal(err)
				}
				if read := negativeZeros(reflect.ValueOf(&decoded)); read != zeros {
					t.Fatalf("the program holds %d -0 constants, its decode %d", zeros, read)
				}
				if zeros > 0 {
					holding.Add(1)
				}
				if got := native.CWithOptions(&decoded, options); got != want {
					line := firstDifferingLine([]byte(want), []byte(got))
					t.Fatalf("the C emitted from the decode differs at line %d:\nlowered: %s\ndecoded: %s", line,
						cLine(want, line), cLine(got, line))
				}
			})
		}
	})
	if !t.Failed() && holding.Load() == 0 {
		t.Fatal("no fixture holds a -0 constant, so no -0 was read back")
	}
}

// cLine is the 1-based line of emitted C, or nothing past its end.
func cLine(text string, line int) string {
	lines := strings.Split(text, "\n")
	if line < 1 || line > len(lines) {
		return ""
	}
	return lines[line-1]
}

func TestGobRoundTripEmitsTheSameRelease(t *testing.T) {
	t.Parallel()
	gobRoundTrip(t, native.Options{})
}

func TestGobRoundTripEmitsTheSameDebug(t *testing.T) {
	t.Parallel()
	gobRoundTrip(t, native.Options{Debug: true})
}

func TestGobRoundTripEmitsTheSameSanitized(t *testing.T) {
	t.Parallel()
	gobRoundTrip(t, native.Options{Sanitize: true})
}

func TestGobRoundTripEmitsTheSameThreadSanitized(t *testing.T) {
	t.Parallel()
	gobRoundTrip(t, native.Options{ThreadSanitize: true})
}

func TestGobRoundTripEmitsTheSameWASI(t *testing.T) {
	t.Parallel()
	gobRoundTrip(t, native.Options{Target: "wasm32-wasi"})
}
