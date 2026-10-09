package main

import (
	"strings"
	"testing"
)

func TestAdaptVarToLet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		src   string
		want  string
		count int
	}{
		{
			name:  "plain",
			src:   "var value = 1;\nvalue;\n",
			want:  "let value = 1;\nvalue;\n",
			count: 1,
		},
		{
			name:  "two names",
			src:   "var left = 1, right = 2;\nleft + right;\n",
			want:  "let left = 1, right = 2;\nleft + right;\n",
			count: 2,
		},
		{
			name: "redeclaration",
			src:  "var value = 1;\nvar value = 2;\n",
			want: "var value = 1;\nvar value = 2;\n",
		},
		{
			name: "read before",
			src:  "value;\nvar value = 1;\n",
			want: "value;\nvar value = 1;\n",
		},
		{
			name: "initializer reads itself",
			src:  "var value = value;\n",
			want: "var value = value;\n",
		},
		{
			name: "closure above the declarator",
			src:  "function read() { return value; }\nvar value = 1;\nread();\n",
			want: "function read() { return value; }\nvar value = 1;\nread();\n",
		},
		{
			name:  "loop variable captured",
			src:   "var fns = [];\nfor (var i = 0; i < 3; i++) { fns.push(function () { return i; }); }\n",
			want:  "let fns = [];\nfor (var i = 0; i < 3; i++) { fns.push(function () { return i; }); }\n",
			count: 1,
		},
		{
			name:  "loop variable not captured",
			src:   "var sum = 0;\nfor (var i = 0; i < 3; i++) { sum = sum + i; }\n",
			want:  "let sum = 0;\nfor (let i = 0; i < 3; i++) { sum = sum + i; }\n",
			count: 2,
		},
		{
			name:  "var in loop body captured",
			src:   "var fns = [];\nfor (var i = 0; i < 3; i++) { var x = i; fns.push(function () { return x; }); }\n",
			want:  "let fns = [];\nfor (let i = 0; i < 3; i++) { var x = i; fns.push(function () { return x; }); }\n",
			count: 2,
		},
		{
			name:  "var in loop body not captured",
			src:   "for (var i = 0; i < 3; i++) { var x = i; x; }\n",
			want:  "for (let i = 0; i < 3; i++) { let x = i; x; }\n",
			count: 2,
		},
		{
			name: "used after the loop",
			src:  "for (var i = 0; i < 3; i++) {}\ni;\n",
			want: "for (var i = 0; i < 3; i++) {}\ni;\n",
		},
		{
			name: "block leaks",
			src:  "if (true) { var value = 1; }\nvalue;\n",
			want: "if (true) { var value = 1; }\nvalue;\n",
		},
		{
			name:  "block does not leak",
			src:   "if (true) { var value = 1; value; }\n",
			want:  "if (true) { let value = 1; value; }\n",
			count: 1,
		},
		{
			name: "switch case",
			src:  "switch (1) { case 1: var value = 1; break; }\n",
			want: "switch (1) { case 1: var value = 1; break; }\n",
		},
		{
			name:  "var inside a string",
			src:   "var value = 1;\nvar text = \"var value = 2\";\n",
			want:  "let value = 1;\nlet text = \"var value = 2\";\n",
			count: 2,
		},
		{
			name: "parameter shadow",
			src:  "function f(value) { var value = 1; }\n",
			want: "function f(value) { var value = 1; }\n",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := adaptSource(test.src)
			if got.Source != test.want {
				t.Fatalf("source\n got %q\nwant %q", got.Source, test.want)
			}
			if got.Counts[adaptVarToLet] != test.count {
				t.Fatalf("count %d, want %d", got.Counts[adaptVarToLet], test.count)
			}
		})
	}
}

func TestAdaptCallbackParam(t *testing.T) {
	t.Parallel()
	src := "function callbackfn(val, idx, obj) {\n  return val;\n}\nvar arr = [1, 2, 3];\narr.forEach(callbackfn);\n"
	got := adaptSource(src)
	if !strings.Contains(got.Source, "function callbackfn(val: number, idx: number, obj: number[])") {
		t.Fatalf("annotation missing:\n%s", got.Source)
	}
	if got.Counts[adaptCallbackParam] != 3 {
		t.Fatalf("count %d", got.Counts[adaptCallbackParam])
	}
	if strings.Contains(got.Source, "\nvar ") {
		t.Fatalf("var should have become let:\n%s", got.Source)
	}

	conflict := "function callbackfn(val) { return val; }\n[1].forEach(callbackfn);\n[\"a\"].forEach(callbackfn);\n"
	got = adaptSource(conflict)
	if strings.Contains(got.Source, "val:") {
		t.Fatalf("conflicting calls were annotated:\n%s", got.Source)
	}

	unknown := "function callbackfn(val) { return val; }\nvar obj = { 0: 1, length: 1 };\nArray.prototype.forEach.call(obj, callbackfn);\n"
	got = adaptSource(unknown)
	if strings.Contains(got.Source, "val:") {
		t.Fatalf("untyped receiver was annotated:\n%s", got.Source)
	}

	inline := "var arr = [1, 2];\narr.forEach(function (val) { return val; });\n"
	got = adaptSource(inline)
	if strings.Contains(got.Source, "val:") {
		t.Fatalf("inline function was annotated; tsc already types it:\n%s", got.Source)
	}
}

func TestAdaptStrictEq(t *testing.T) {
	t.Parallel()
	got := adaptSource("var value = 1;\nif (value == 1) {}\nif (typeof value == \"number\") {}\nif (value != 2) {}\n")
	if !strings.Contains(got.Source, "value === 1") || !strings.Contains(got.Source, "typeof value === \"number\"") || !strings.Contains(got.Source, "value !== 2") {
		t.Fatalf("missing strict ops:\n%s", got.Source)
	}
	if got.Counts[adaptStrictEq] != 3 {
		t.Fatalf("count %d\n%s", got.Counts[adaptStrictEq], got.Source)
	}
	coerced := adaptSource("var value = 1;\nif (value == \"1\") {}\n")
	if strings.Contains(coerced.Source, "===") {
		t.Fatalf("coercing equality was rewritten:\n%s", coerced.Source)
	}
	chain := adaptSource("var value = 1;\nif (value == 1 == 1) {}\n")
	// The inner comparison is number === number. The outer is boolean == number, which coerces, so it stays.
	if !strings.Contains(chain.Source, "value === 1 == 1") {
		t.Fatalf("chain:\n%s", chain.Source)
	}
}

func TestAdaptThrowError(t *testing.T) {
	t.Parallel()
	got := adaptSource("if (false) { throw new Test262Error(\"nope\"); }\n")
	if !strings.Contains(got.Source, "throw new Error(\"nope\")") {
		t.Fatalf("throw not rewritten:\n%s", got.Source)
	}
	if got.Counts[adaptThrowError] != 1 {
		t.Fatalf("count %d", got.Counts[adaptThrowError])
	}
	kept := adaptSource("try { throw new Test262Error(\"nope\"); } catch (e) { if (e instanceof Test262Error) { throw e; } }\n")
	if !strings.Contains(kept.Source, "throw new Test262Error") {
		t.Fatalf("instanceof test was rewritten:\n%s", kept.Source)
	}
}

func TestAdaptLeavesCheckoutText(t *testing.T) {
	t.Parallel()
	// The rewrite is a pure function of the source. The runner never writes it back; this locks
	// that adaptSource is the whole mechanism and a confused file is returned untouched.
	src := "class Box { value: number; }\n"
	got := adaptSource(src)
	if got.Source != src || len(got.Counts) != 0 {
		t.Fatalf("class file changed: %#v", got)
	}
}

func TestClassifyAdapt(t *testing.T) {
	t.Parallel()
	source := "/*---\ndescription: x\n---*/\nvar value = 1;\nassert.sameValue(value == 1, true);\n"
	got := classify("t.js", source, true)
	if got.Skip != "" {
		t.Fatal(got.Skip)
	}
	if got.Adaptations[adaptVarToLet] != 1 || got.Adaptations[adaptStrictEq] != 1 {
		t.Fatalf("counts %v", got.Adaptations)
	}
	if strings.Contains(got.Program, "var value") || !strings.Contains(got.Program, "value === 1") {
		t.Fatalf("program was not adapted:\n%s", got.Program)
	}
	plain := classify("t.js", source, false)
	if strings.Contains(plain.Program, "let value") || plain.Adaptations[adaptVarToLet] != 0 {
		t.Fatal("adapt off still rewrote var")
	}
}

func TestCrashPath(t *testing.T) {
	t.Parallel()
	got := withCrashPath("built-ins/Array/prototype/sort/x.js", "clang: no newline at end of file")
	want := "built-ins/Array/prototype/sort/x.js: clang: no newline at end of file"
	if got != want {
		t.Fatalf("got %q", got)
	}
}
