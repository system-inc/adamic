package lower

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// objectAssignCallableRefusal is what objectAssignCallableSourceRefusal says, up to the source's type.
const objectAssignCallableRefusal = "Adamic 0.1 refuses Object.assign with the callable source "

// Correctness's assign-function fixtures (typescript/record-handle-fixtures 420caaea, the logging-only
// adaptation, with the ambient enums of kinds 3, 9 and 10 standing as their value types) each copy a
// function, a constructor or an optional function into a record and read the result as the function: tsc
// passes it, and Node throws TypeError at the call. Each stops at the new refusal on its Object.assign line
// (#k8aak04, Adamic's ruling on #ny0c3gb). Mutant (testdata/object_assign_callable/mutants.md): without the
// rule they stop at the intersection value's not-yet, or kind 4 construct's ClassExpression, and go red here.
func TestObjectAssignCallableSourceIsRefused(t *testing.T) {
	t.Parallel()
	directory := "testdata/object_assign_callable"
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	fixtures := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), "--assign-function.ts") {
			continue
		}
		fixtures++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			line := 0
			for index, text := range strings.Split(string(source), "\n") {
				if strings.Contains(text, "Object.assign(") {
					line = index + 1
				}
			}
			_, err = lowerModules(t, map[string]string{"main.ts": string(source)})
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), fmt.Sprintf("main.ts:%d:", line)) || !strings.Contains(err.Error(), objectAssignCallableRefusal) {
				t.Errorf("got %v, want main.ts:%d: %s...", err, line, objectAssignCallableRefusal)
			}
		})
	}
	if fixtures != 30 {
		t.Fatalf("%d assign-function fixtures in %s, want Correctness's 30", fixtures, directory)
	}
}

// Language's class copies: Object.assign copies a class instance's own fields, never its prototype's methods,
// so Node throws "copy.bark is not a function" and "copy.read is not a function". Neither source is callable,
// so they stop where they stopped before this rule, as a widened source.
func TestObjectAssignClassCopiesStayRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"class-dog.ts", "class-box.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/object_assign_callable", name))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerModules(t, map[string]string{"main.ts": string(source)})
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "Adamic 0.1 refuses Object.assign; each source must have its complete shape proven") {
				t.Errorf("got %v, want the widened-source refusal", err)
			}
		})
	}
}

// The call is judged by the lib declaration it resolves to and the source's type, never by its spelling: an
// alias, an element access, a spread, a union holding a function, lib's Function, a class's constructor and a
// type parameter whose constraint is callable, alone or in a union, all stop at the rule. A source that isn't
// callable is left to the lowering's own rules.
func TestObjectAssignCallableSourceRoads(t *testing.T) {
	t.Parallel()
	const prelude = "type Receiver = (path: string) => string;\nconst h: { value: number } = { value: 1 };\nconst fn: Receiver = path => path;\n"
	for _, probe := range []struct{ name, source string }{
		{"an alias", "const assign = Object.assign;\nconst r = assign(h, fn);\n"},
		{"an element access", "const r = Object[\"assign\"](h, fn);\n"},
		{"a spread", "const r = Object.assign(h, ...[fn]);\n"},
		{"a union holding a function", "function pick(flag: boolean): Receiver | { value: number } { return flag ? fn : { value: 2 }; }\nconst r = Object.assign(h, pick(true));\n"},
		{"lib's Function", "const any: Function = fn;\nconst r = Object.assign(h, any);\n"},
		{"a class's constructor", "class Dog { name = 'rex'; }\nconst r = Object.assign(h, Dog);\n"},
		{"globalThis.Object", "const r = globalThis.Object.assign(h, fn);\n"},
		{"a type parameter constrained to a function", "function copy<F extends () => void>(target: { value: number }, f: F): void {\n\tObject.assign(target, f);\n}\ncopy(h, (): void => {});\n"},
		{"a type parameter constrained to a union holding a function", "function copy<F extends Receiver | { value: number }>(target: { value: number }, f: F): void {\n\tObject.assign(target, f);\n}\ncopy(h, fn);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, map[string]string{"main.ts": prelude + probe.source})
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), objectAssignCallableRefusal) {
				t.Errorf("got %v, want %s...", err, objectAssignCallableRefusal)
			}
		})
	}
	t.Run("a spread whose element types aren't known", func(t *testing.T) {
		t.Parallel()
		// A class extending Array<Receiver> holds functions, and its own type has no type arguments to read.
		_, err := lowerModules(t, map[string]string{"main.ts": prelude + "class Receivers extends Array<Receiver> {}\nconst many = new Receivers();\nconst r = Object.assign(h, ...many);\n"})
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "Adamic 0.1 refuses Object.assign with a spread source whose element types aren't known") {
			t.Errorf("got %v, want the unread spread's refusal", err)
		}
	})
	t.Run("a plain source", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerModules(t, map[string]string{"main.ts": "const h = { value: 1 };\nconst r = Object.assign(h, { value: 2 });\nconsole.log(String(r.value));\n"}); err != nil {
			t.Errorf("want it to lower, got %v", err)
		}
	})
}
