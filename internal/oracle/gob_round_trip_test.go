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

// The round trip is five families, one for each way native emits a program (release, Debug, the two sanitized builds
// and wasm32-wasi), each sharded as the oracle's others are (families_test.go): one over the whole corpus was 187 s of
// one core per variant (Golem, Oct 31), most of it checking the source.
var gobRoundTripVariants = map[string]native.Options{
	"TestGobRoundTripEmitsTheSameRelease":         {},
	"TestGobRoundTripEmitsTheSameDebug":           {Debug: true},
	"TestGobRoundTripEmitsTheSameSanitized":       {Sanitize: true},
	"TestGobRoundTripEmitsTheSameThreadSanitized": {ThreadSanitize: true},
	"TestGobRoundTripEmitsTheSameWASI":            {Target: "wasm32-wasi"},
}

// gobRoundTripWitness is a fixture whose lowering holds a -0 constant, each family's planted case. Its probe reads the
// decode back with every -0 made +0, gob's own reading of a float written by value.
const gobRoundTripWitness = "internal/oracle/testdata/WITNESS"

func runGobRoundTripReleaseShard(t *testing.T, shard int) {
	gobRoundTripShard(t, "TestGobRoundTripEmitsTheSameRelease", shard)
}

func runGobRoundTripDebugShard(t *testing.T, shard int) {
	gobRoundTripShard(t, "TestGobRoundTripEmitsTheSameDebug", shard)
}

func runGobRoundTripSanitizedShard(t *testing.T, shard int) {
	gobRoundTripShard(t, "TestGobRoundTripEmitsTheSameSanitized", shard)
}

func runGobRoundTripThreadSanitizedShard(t *testing.T, shard int) {
	gobRoundTripShard(t, "TestGobRoundTripEmitsTheSameThreadSanitized", shard)
}

func runGobRoundTripWASIShard(t *testing.T, shard int) {
	gobRoundTripShard(t, "TestGobRoundTripEmitsTheSameWASI", shard)
}

// gobRoundTripShard lowers each fixture its shard owns, writes the program as gob twice before anything emits it, and
// holds the bytes equal; then emits the program with the family's options, emits a fresh decode of the bytes with the
// same options, and holds the two C byte for byte equal, and every -0 constant the program holds a -0 in the decode.
// Emit writes into the IR (element_borrow.go, async_ordinary.go), so each side is emitted once, from a program nothing
// emitted before.
func gobRoundTripShard(t *testing.T, name string, shard int) {
	t.Helper()
	ir.RegisterGob()
	options := gobRoundTripVariants[name]
	planted, probing := familyProbe(name)
	keys := gobRoundTripPaths()
	for _, index := range familyCases(t, name, shard, keys) {
		path := keys[index]
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			program := gobRoundTripLowered(t, path)
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
			if probing && path == planted {
				clearNegativeZeros(reflect.ValueOf(&decoded))
			}
			if read := negativeZeros(reflect.ValueOf(&decoded)); read != zeros {
				t.Fatalf("the program holds %d -0 constants, its decode %d", zeros, read)
			}
			if zeros > 0 {
				t.Logf("holds %d -0 constants, as its decode does", zeros)
			}
			if got := native.CWithOptions(&decoded, options); got != want {
				line := firstDifferingLine([]byte(want), []byte(got))
				t.Fatalf("the C emitted from the decode differs at line %d:\nlowered: %s\ndecoded: %s", line,
					cLine(want, line), cLine(got, line))
			}
		})
	}
}

// gobRoundTripLowered is a fixture's lowered program, from the repository root.
func gobRoundTripLowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	source, err := load.Load([]string{filepath.Join(repository, path)})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

// The corpus the round trip reads holds a -0 constant, so its -0 check reads one back in every variant: the witness is
// one of its cases and lowers to one. Mutant: a witness that holds none.
func TestGobRoundTripReadsANegativeZero(t *testing.T) {
	t.Parallel()
	if !slices.Contains(gobRoundTripPaths(), gobRoundTripWitness) {
		t.Fatalf("%s isn't one of the round trip's fixtures", gobRoundTripWitness)
	}
	if zeros := negativeZeros(reflect.ValueOf(gobRoundTripLowered(t, gobRoundTripWitness))); zeros == 0 {
		t.Fatalf("%s holds no -0 constant", gobRoundTripWitness)
	}
}

// clearNegativeZeros makes every -0 constant a value holds +0, in place.
func clearNegativeZeros(value reflect.Value) {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return
		}
		if constant, ok := value.Interface().(ir.NumberConstant); ok {
			if constant.Value == 0 {
				value.Set(reflect.ValueOf(ir.NumberConstant{}))
			}
			return
		}
		// What an interface holds can't be set through it: set a copy, then the copy back.
		held := reflect.New(value.Elem().Type()).Elem()
		held.Set(value.Elem())
		clearNegativeZeros(held)
		value.Set(held)
	case reflect.Pointer:
		if !value.IsNil() {
			clearNegativeZeros(value.Elem())
		}
	case reflect.Struct:
		for index := range value.NumField() {
			if value.Type().Field(index).IsExported() {
				clearNegativeZeros(value.Field(index))
			}
		}
	case reflect.Slice, reflect.Array:
		for index := range value.Len() {
			clearNegativeZeros(value.Index(index))
		}
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
