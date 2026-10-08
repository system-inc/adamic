package lint

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// The registry owns rule membership; the captured corpus and Go parser own JSX
// membership and case counts. New descriptors need no shared inventory edit.
func discoverJsxInventory(descriptors []registry.Descriptor, rows []string, hasJsx func(string) bool) ([]string, map[string]int, error) {
	registered := make(map[string]bool, len(descriptors))
	for _, descriptor := range descriptors {
		registered[descriptor.Name] = true
	}
	var paths []string
	counts := map[string]int{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 || fields[0] == "" {
			return nil, nil, fmt.Errorf("malformed captured JSX inventory row %q", row)
		}
		if !registered[fields[1]] {
			return nil, nil, fmt.Errorf("captured JSX inventory names unregistered rule %q", fields[1])
		}
		if hasJsx(fields[0]) {
			paths = append(paths, fields[0])
			counts[fields[1]]++
		}
	}
	// Keep a disappearance guard without pinning rule names or case counts:
	// each discovered JSX subscription must have actual captured JSX coverage.
	for _, descriptor := range descriptors {
		for _, kind := range descriptor.Kinds {
			if strings.HasPrefix(kind, "Jsx") && counts[descriptor.Name] == 0 {
				return nil, nil, fmt.Errorf("registered JSX listener %q has no captured JSX cases", descriptor.Name)
			}
		}
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("captured upstream corpus contains no JSX")
	}
	return paths, counts, nil
}

func TestJsxInventoryDiscovery(t *testing.T) {
	descriptors := []registry.Descriptor{
		{Name: "first", Kinds: []string{"JsxText"}},
		{Name: "generic", Kinds: []string{"CallExpression"}},
	}
	rows := []string{"first.tsx\tfirst", "generic.tsx\tgeneric", "plain.ts\tgeneric"}
	hasJsx := func(path string) bool { return strings.HasSuffix(path, ".tsx") }
	t.Run("new rule and cases require no inventory edit", func(t *testing.T) {
		expanded := append(append([]registry.Descriptor{}, descriptors...), registry.Descriptor{Name: "later", Kinds: []string{"JsxExpression"}})
		cases := append(append([]string{}, rows...), "later-one.tsx\tlater", "later-two.tsx\tlater")
		paths, counts, err := discoverJsxInventory(expanded, cases, hasJsx)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(paths, []string{"first.tsx", "generic.tsx", "later-one.tsx", "later-two.tsx"}) || !reflect.DeepEqual(counts, map[string]int{"first": 1, "generic": 1, "later": 2}) {
			t.Fatalf("discovered paths %v and counts %v", paths, counts)
		}
	})
	t.Run("missing JSX coverage is refused", func(t *testing.T) {
		_, _, err := discoverJsxInventory(descriptors, rows[1:], hasJsx)
		if err == nil || !strings.Contains(err.Error(), `registered JSX listener "first" has no captured JSX cases`) {
			t.Fatalf("missing coverage: %v", err)
		}
	})
	t.Run("unregistered captured rule is refused", func(t *testing.T) {
		_, _, err := discoverJsxInventory(descriptors, append(rows, "unknown.tsx\tunknown"), hasJsx)
		if err == nil || !strings.Contains(err.Error(), `unregistered rule "unknown"`) {
			t.Fatalf("unknown rule: %v", err)
		}
	})
	t.Run("malformed row is refused", func(t *testing.T) {
		_, _, err := discoverJsxInventory(descriptors, []string{"missing-rule.tsx"}, hasJsx)
		if err == nil || !strings.Contains(err.Error(), "malformed captured JSX inventory row") {
			t.Fatalf("malformed row: %v", err)
		}
	})
	t.Run("empty JSX inventory is refused", func(t *testing.T) {
		_, _, err := discoverJsxInventory(descriptors[1:], rows[2:], hasJsx)
		if err == nil || !strings.Contains(err.Error(), "contains no JSX") {
			t.Fatalf("empty inventory: %v", err)
		}
	})
}
