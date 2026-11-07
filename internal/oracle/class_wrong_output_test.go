package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestClassWrongOutput103(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, output, path string }{
		{"init_super_number", "score NaN\n", "Base.describe -> Derived.score -> Derived.count"},
		{"init_super_getter", "score NaN\n", "Base.summary -> Derived.score -> Derived.count"},
		{"init_super", "caught TypeError\n", "Base.describe -> Derived.name -> Derived.label"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(filepath.Join(
				repository,
				"internal/oracle/testdata/class_wrong_output_refused/classfeat_"+probe.name+".a",
			))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			observed := onNode(test, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				test.Fatalf("Node: %+v", observed)
			}
			_, loweredError := lowered(test, path)
			var refusal *lower.Refused
			if !errors.As(loweredError, &refusal) ||
				refusal.What != "a field initializer reading an uninitialized field through "+probe.path ||
				refusal.Fix != "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set" {
				test.Fatalf("want path and repair, got %v", loweredError)
			}
		})
	}
}

func TestClassWrongOutput107(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, output string }{
		{"iterators_override_this", "0\n100\nscaled:return\n0,100,200\n0,100,200\nscaled:return\n0\n"},
		{"iterators_hidden_return", "0\nclosing:return\nclosing:return\n0\n"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/class_wrong_output_refused/"+probe.name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			observed := onNode(test, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				test.Fatalf("Node: %+v", observed)
			}
			_, loweredError := lowered(test, path)
			var refusal *lower.Refused
			if !errors.As(loweredError, &refusal) ||
				refusal.What != "an iterator factory returning this whose runtime next or return can differ from its declared iterator type (adamic/iterator-receiver-origin)" ||
				refusal.Fix != "return a separate iterator object with next and return closures; do not erase subclass protocol methods behind the base iterator return type" {
				test.Fatalf("want pinned receiver-origin refusal, got %v", loweredError)
			}
		})
	}
}

func TestClassWrongOutput108(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/class_wrong_output_refused/iterators_sym_keys_view.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	observed := onNode(test, path)
	if observed.exitCode != 0 || string(observed.stdout) != "value\n1,2\nvalue\nvalue\n" || len(observed.stderr) != 0 {
		test.Fatalf("Node: %+v", observed)
	}
	_, loweredError := lowered(test, path)
	var refusal *lower.Refused
	if !errors.As(loweredError, &refusal) ||
		refusal.What != "a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view)" ||
		refusal.Fix != "pass a new plain object containing only the desired string-keyed fields to Object.keys" {
		test.Fatalf("want pinned symbol-key-view refusal, got %v", loweredError)
	}
}

func TestClassWrongOutput106(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, output, member string }{
		{"classfeat_static_private_instance", "s1\nt2\nhidden t2\n", "#secret"},
		{"classfeat_static_private_method", "first function\nhidden s1\n", "#hidden"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/class_wrong_output_refused/"+probe.name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			observed := onNode(test, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				test.Fatalf("Node: %+v", observed)
			}
			_, loweredError := lowered(test, path)
			var refusal *lower.Refused
			if !errors.As(loweredError, &refusal) ||
				refusal.What != "instance private storage reached from a static member: Box."+
					probe.member+
					" (adamic/private-instance-from-static)" ||
				refusal.Fix != "move the private access into an instance method; call that instance method from the static member" {
				test.Fatalf("want pinned private-instance-from-static refusal, got %v", loweredError)
			}
		})
	}
}
