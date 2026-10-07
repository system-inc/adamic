package tailwind

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"reflect"
	"sort"
	"strings"
)

// The oracle exports the real prerequisite answers and observes the actual helper.
// No settings-key, regexp compiler, compiled cache or file cache is rewritten.
func AdamicReaderForCorpus() []byte {
	settings := []ClassLiteralSettings{DefaultClassLiteralSettings(), {},
		{AttributeNames: []string{"tw"}}, {AttributeNames: []string{"ab"}},
		{AttributeNames: []string{"a", "b"}}, {AttributeNames: []string{"a"}, CalleeNames: []string{"b"}},
		{AttributeNames: []string{"className", "className"}, CalleeNames: []string{"cn"}, VariablePatterns: []string{"[", "^classes$", "^classes$"}},
		{AttributeNames: []string{"a\x00b"}, VariablePatterns: []string{".*", "(?i)^x$"}},
	}
	type Definition struct {
		Key                           string
		Attributes, Callees, Patterns []string
	}
	definitions := []Definition{}
	var want strings.Builder
	for index, s := range settings {
		key := s.key()
		compiled := compiledClassLiteralReader(key, s)
		d := Definition{Key: key, Attributes: []string{}, Callees: []string{}, Patterns: []string{}}
		for name := range compiled.attributeNames {
			d.Attributes = append(d.Attributes, name)
		}
		sort.Strings(d.Attributes)
		for name := range compiled.calleeNames {
			d.Callees = append(d.Callees, name)
		}
		sort.Strings(d.Callees)
		for _, p := range compiled.variablePatterns {
			d.Patterns = append(d.Patterns, p.String())
		}
		definitions = append(definitions, d)
		cache := rule.NewFileCache()
		first := ClassLiteralReaderFor(cache, s)
		node := ast.NewNodeFactory(ast.NodeFactoryHooks{}).NewIdentifier("sentinel")
		first.values[node] = classValues{}
		second := ClassLiteralReaderFor(cache, s)
		other := ClassLiteralReaderFor(rule.NewFileCache(), s)
		nilA, nilB := ClassLiteralReaderFor(nil, s), ClassLiteralReaderFor(nil, s)
		zero := &rule.FileCache{}
		zeroA, zeroB := ClassLiteralReaderFor(zero, s), ClassLiteralReaderFor(zero, s)
		// The heterogeneous FileCache declines a cached value of a different type.
		collision := rule.NewFileCache()
		rule.Cached(collision, "tailwind.classValues:"+key, func() int { return 7 })
		collisionA, collisionB := ClassLiteralReaderFor(collision, s), ClassLiteralReaderFor(collision, s)
		fmt.Fprintf(&want, "%d:%t:%d:%t:%d:%t:%t:%t:%t:%t:%d:%d:%d\n", index,
			first == second, len(second.values), first == other, len(other.values), nilA == nilB, zeroA == zeroB, collisionA == collisionB,
			reflect.ValueOf(first.attributeNames).Pointer() == reflect.ValueOf(other.attributeNames).Pointer(),
			reflect.ValueOf(first.calleeNames).Pointer() == reflect.ValueOf(other.calleeNames).Pointer(),
			len(first.attributeNames), len(first.calleeNames), len(first.variablePatterns))
		fmt.Fprintf(&want, "fills:%d:%d:%t\n", cache.Fills()["tailwind.classValues:"+key], len(cache.Fills()), reflect.DeepEqual(first.variablePatterns, other.variablePatterns))
	}
	// Different settings in one file must stay separate, but their shared keys
	// (including a real NUL collision) follow settings.key rather than deep equality.
	cache := rule.NewFileCache()
	readers := []*ClassLiteralReader{}
	for _, s := range settings {
		readers = append(readers, ClassLiteralReaderFor(cache, s))
	}
	for i, a := range readers {
		for j, b := range readers {
			fmt.Fprintf(&want, "identity:%d:%d:%t\n", i, j, a == b)
		}
	}
	data, err := json.Marshal(struct {
		Definitions []Definition
		Want        string
	}{definitions, want.String()})
	if err != nil {
		panic(err)
	}
	return data
}
