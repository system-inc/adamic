package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestClassWrongOutput103(t *testing.T) {
	for _, probe := range []struct{ name, output, path string }{
		{"init_super_number", "score NaN\n", "Base.describe -> Derived.score -> Derived.count"},
		{"init_super_getter", "score NaN\n", "Base.summary -> Derived.score -> Derived.count"},
		{"init_super", "caught TypeError\n", "Base.describe -> Derived.name -> Derived.label"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/classfeat_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node: %+v", observed)
			}
			_, err = lowered(t, path)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || refusal.What != "a field initializer reading an uninitialized field through "+probe.path || refusal.Fix != "declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set" {
				t.Fatalf("want path and repair, got %v", err)
			}
		})
	}
}

func TestClassWrongOutput107(t *testing.T) {
	for _, probe := range []struct{ name, output string }{
		{"iterators_override_this", "0\n100\nscaled:return\n0,100,200\n0,100,200\nscaled:return\n0\n"},
		{"iterators_hidden_return", "0\nclosing:return\nclosing:return\n0\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node: %+v", observed)
			}
			_, err = lowered(t, path)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || refusal.What != "an iterator factory returning this whose runtime next or return can differ from its declared iterator type (adamic/iterator-receiver-origin)" || refusal.Fix != "return a separate iterator object with next and return closures; do not erase subclass protocol methods behind the base iterator return type" {
				t.Fatalf("want pinned receiver-origin refusal, got %v", err)
			}
		})
	}
}

func TestClassWrongOutput108(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/iterators_sym_keys_view.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "value\n1,2\nvalue\nvalue\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node: %+v", observed)
	}
	_, err = lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || refusal.What != "a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view)" || refusal.Fix != "pass a new plain object containing only the desired string-keyed fields to Object.keys" {
		t.Fatalf("want pinned symbol-key-view refusal, got %v", err)
	}
}

func TestClassWrongOutput106(t *testing.T) {
	for _, probe := range []struct{ name, output, member string }{
		{"classfeat_static_private_instance", "s1\nt2\nhidden t2\n", "#secret"},
		{"classfeat_static_private_method", "first function\nhidden s1\n", "#hidden"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != probe.output || len(observed.stderr) != 0 {
				t.Fatalf("Node: %+v", observed)
			}
			_, err = lowered(t, path)
			var refusal *lower.Refused
			if !errors.As(err, &refusal) || refusal.What != "instance private storage reached from a static member: Box."+probe.member+" (adamic/private-instance-from-static)" || refusal.Fix != "move the private access into an instance method; call that instance method from the static member" {
				t.Fatalf("want pinned private-instance-from-static refusal, got %v", err)
			}
		})
	}
}
