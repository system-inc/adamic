package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassRefusalsSyntaxCoverage(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, rule string
		notYet     bool
	}{
		{"super_parentheses", "uninitialized field", false},
		{"super_optional", "uninitialized field", false},
		{"super_arrow_call", "uninitialized field", false},
		{"super_getter_override", "uninitialized field", false},
		{"super_nested", "uninitialized field", false},
		{"super_generic", "uninitialized field", false},
		{"super_interface", "uninitialized field", false},
		{"super_union", "uninitialized field", false},
		{"iterator_for_of", "iterator-receiver-origin", false},
		{"iterator_spread", "iterator-receiver-origin", false},
		{"iterator_from", "iterator-receiver-origin", false},
		{"iterator_destructure", "iterator-receiver-origin", false},
		{"iterator_alias", "iterator-receiver-origin", false},
		{"iterator_arrow", "iterator-receiver-origin", false},
		{"iterator_getter", "iterator-receiver-origin", false},
		{"iterator_generic", "iterator-receiver-origin", false},
		{"iterator_interface", "an iterator method whose concrete origin is erased by a structural signature", true},
		{"iterator_union", "iterator-receiver-origin", false},
		{"iterator_optional", "iterator-receiver-origin", false},
		{"iterator_return_alias", "iterator-receiver-origin", false},
		{"keys_interface", "symbol-key-view", false},
		{"keys_arrow", "symbol-key-view", false},
		{"keys_generic", "symbol-key-view", false},
		{"keys_getter", "symbol-key-view", false},
		{"keys_union", "symbol-key-view", false},
		{"keys_optional", "symbol-key-view", false},
		{"keys_nested", "a function inside a function (a closure)", true},
		{"private_arrow", "private-instance-from-static", false},
		{"private_getter", "private-instance-from-static", false},
		{"private_generic", "private-instance-from-static", false},
		{"private_union", "private-instance-from-static", false},
		{"private_nested", "private-instance-from-static", false},
		{"private_write", "private-instance-from-static", false},
		{"keys_optional_call", "a call through ?. (an optional call)", true},
		{"keys_nested_arrow", "symbol-key-view", false},
		{"private_optional_holder", "private-instance-from-static", false},
		{"private_interface", "private-instance-from-static", false},
		{"iterator_optional_chain", "iterator-receiver-origin", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("../oracle/testdata/class_refusals_refused", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var refusal *Refused
			var notYet *NotYet
			expected := errors.As(err, &refusal)
			if probe.notYet {
				expected = errors.As(err, &notYet)
			}
			if !expected || !strings.Contains(err.Error(), probe.rule) {
				t.Fatalf("want %q (NotYet=%t), got %v", probe.rule, probe.notYet, err)
			}
		})
	}
}

func TestClassRefusalsSafeNeighbors(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("../oracle/testdata/class_refusals_safe_*.a")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no safe neighbors")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lowerSource(t, string(source)); err != nil {
				t.Fatalf("safe neighbor refused: %v", err)
			}
		})
	}
}
