package lower

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

var cycleComparisonInputs = sync.OnceValues(func() ([]string, error) {
	var paths []string
	err := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel("../..", path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if entry.Name() == ".git" || relative == "cohere" {
				return filepath.SkipDir
			}
			return nil
		}
		fixture := strings.HasPrefix(relative, "internal/oracle/testdata/") || strings.HasPrefix(relative, "internal/lower/testdata/") || strings.HasPrefix(relative, "review/") || strings.HasPrefix(relative, "stage3/fixtures/cycles/")
		extension := filepath.Ext(path)
		if extension == ".a" || fixture && (extension == ".ts" || extension == ".tsx") {
			paths = append(paths, relative)
		}
		return nil
	})
	slices.Sort(paths)
	return paths, err
})

func cycleComparisonShard(path string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(path))
	return h.Sum64() % 128
}

// Compare the entire IR as well as refusal identity, rather than just admission.
// This includes every graph/Program flag, membership-by-type map, frame, cell,
// weak field and ownership-query consequence carried by the lowered program.
func cycleCompareProgram(t *testing.T, path string) {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join("../..", path))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{absolute})
	if err != nil {
		t.Logf("%s: common loader refusal: %v", path, err)
		return
	}
	for _, region := range []bool{false, true} {
		// Use fresh, independently initialized checkers. Reusing one checker changes
		// fresh literal type IDs on the second lowering, even for scanner vs scanner.
		if region {
			loaded, err = load.Load([]string{absolute})
			if err != nil {
				t.Fatal(err)
			}
		}
		indexedInput, err := load.Load([]string{absolute})
		if err != nil {
			t.Fatal(err)
		}
		options := Options{ProgramRegion: region, OwnershipQuery: true}
		before, scanErr := scanLowerWithOptions(context.Background(), loaded, options)
		after, indexErr := LowerWithOptions(context.Background(), indexedInput, options)
		text := func(err error) string {
			if err == nil {
				return ""
			}
			return fmt.Sprintf("%T: %v", err, err)
		}
		var scanRefused, indexRefused *Refused
		errors.As(scanErr, &scanRefused)
		errors.As(indexErr, &indexRefused)
		cycle := func(r *Refused) bool { return r != nil && r.cycle }
		if text(scanErr) != text(indexErr) || cycle(scanRefused) != cycle(indexRefused) {
			t.Fatalf("cycle equality difference: %s ProgramRegion=%t\nscanning: %s (cycle=%t)\nindexed: %s (cycle=%t)", path, region, text(scanErr), cycle(scanRefused), text(indexErr), cycle(indexRefused))
		}
		if difference := cycleFirstDifference(reflect.ValueOf(before), reflect.ValueOf(after), "program"); difference != "" {
			t.Fatalf("cycle equality difference: %s ProgramRegion=%t: %s", path, region, difference)
		}
	}
	t.Logf("%s: equal scanning/indexed outcomes in ordinary and Program modes", path)
}

func cycleFirstDifference(a, b reflect.Value, path string) string {
	if !a.IsValid() || !b.IsValid() {
		return fmt.Sprintf("%s: validity %t/%t", path, a.IsValid(), b.IsValid())
	}
	if a.Type() != b.Type() {
		return fmt.Sprintf("%s: types %v/%v", path, a.Type(), b.Type())
	}
	if a.CanInterface() && b.CanInterface() && reflect.DeepEqual(a.Interface(), b.Interface()) {
		return ""
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			return fmt.Sprintf("%s: nil scanning=%t indexed=%t", path, a.IsNil(), b.IsNil())
		}
		return cycleFirstDifference(a.Elem(), b.Elem(), path)
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !a.Field(i).CanInterface() {
				continue
			}
			if d := cycleFirstDifference(a.Field(i), b.Field(i), path+"."+a.Type().Field(i).Name); d != "" {
				return d
			}
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: scanning length=%d indexed length=%d", path, a.Len(), b.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if d := cycleFirstDifference(a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i)); d != "" {
				return d
			}
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: scanning=%v indexed=%v", path, a.Interface(), b.Interface())
		}
		for _, key := range a.MapKeys() {
			if d := cycleFirstDifference(a.MapIndex(key), b.MapIndex(key), fmt.Sprintf("%s[%v]", path, key)); d != "" {
				return d
			}
		}
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(a.Float()) && math.IsNaN(b.Float()) {
			return ""
		}
		return fmt.Sprintf("%s: scanning=%v indexed=%v", path, a.Float(), b.Float())
	default:
		if a.CanInterface() && b.CanInterface() {
			return fmt.Sprintf("%s: scanning=%#v indexed=%#v", path, a.Interface(), b.Interface())
		}
	}
	return "" // Unexported IR analysis caches do not carry ownership answers.
}

func runCycleFinderEqualityShard(t *testing.T, shard uint64) {
	t.Helper()
	paths, err := cycleComparisonInputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 3500 {
		t.Fatalf("cycle comparison inventory shrank: %d", len(paths))
	}
	for _, path := range paths {
		if cycleComparisonShard(path) == shard {
			t.Run(path, func(t *testing.T) { cycleCompareProgram(t, path) })
		}
	}
}

func TestCycleIndexCandidateEdgeEquality(t *testing.T) {
	t.Parallel()
	indexed, types := cycleIndexTypes(t, `declare const view:{visible:number};declare const actual:{visible:number;hidden:{target:true}};declare const target:{target:true};`)
	indexed.shapes = []*checker.Type{types["actual"]}
	scanning := &scanCycleFinder{l: indexed.l, where: indexed.where, shapes: indexed.shapes}
	target := cycleNode{proven: types["target"]}
	before, after := scanning.reaches(types["view"], target), indexed.reaches(types["view"], target)
	if !before || before != after {
		t.Fatalf("candidate-edge equality: scanning=%t indexed=%t", before, after)
	}
}
