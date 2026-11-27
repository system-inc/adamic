package lower

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// relationSizeTrace is a probe, never for main (#nxb3xh1): each relation cache's size, read by name as
// load.RelationOverflowMarked reads the caches, printed on stderr as one RELSIZE line.
func relationSizeTrace(typeChecker *checker.Checker, label string) {
	value := reflect.ValueOf(typeChecker).Elem()
	var parts []string
	total := 0
	for _, name := range []string{"subtypeRelation", "strictSubtypeRelation", "assignableRelation", "comparableRelation", "identityRelation"} {
		relation := value.FieldByName(name)
		size := -1
		if relation.IsValid() && relation.Kind() == reflect.Pointer && !relation.IsNil() {
			if results := relation.Elem().FieldByName("results"); results.IsValid() && results.Kind() == reflect.Map {
				size = results.Len()
				total += size
			}
		}
		parts = append(parts, fmt.Sprintf("%s=%d", strings.TrimSuffix(name, "Relation"), size))
	}
	if enum := value.FieldByName("enumRelation"); enum.IsValid() && enum.Kind() == reflect.Map {
		parts = append(parts, fmt.Sprintf("enum=%d", enum.Len()))
	}
	fmt.Fprintf(os.Stderr, "RELSIZE %s %s total=%d %s\n", time.Now().UTC().Format("15:04:05"), label, total, strings.Join(parts, " "))
}

// armPairs and armCalls are the probe's count, per arm of graphLinks, of assignable pairs added and calls made,
// printed and reset at each findCycles end (#nxb3xh1, Compiler's (b)).
var (
	armPairs = map[string]int{}
	armCalls = map[string]int{}
)

func assignableSize(typeChecker *checker.Checker) int {
	relation := reflect.ValueOf(typeChecker).Elem().FieldByName("assignableRelation")
	if !relation.IsValid() || relation.Kind() != reflect.Pointer || relation.IsNil() {
		return -1
	}
	results := relation.Elem().FieldByName("results")
	if !results.IsValid() || results.Kind() != reflect.Map {
		return -1
	}
	return results.Len()
}

func armTrace(label string) {
	keys := make([]string, 0, len(armPairs))
	for key := range armCalls {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(os.Stderr, "ARM %s %s pairs=%d calls=%d\n", label, key, armPairs[key], armCalls[key])
	}
	armPairs = map[string]int{}
	armCalls = map[string]int{}
}
