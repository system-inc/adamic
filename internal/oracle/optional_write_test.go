package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOptionalCheckedWrites(t *testing.T) {
	for _, probe := range []struct{ name, node, failure string }{
		{"uninitialized-slot", "3\n", ""},
		{"dictionary-own-slot", "3\n", ""},
		{"second-optional", "after\n", ""},
		{"boolean-write", "true\n", ""},
		{"fresh-undefined", "done\n", ""},
		{"fresh-boolean", "done\n", ""},
		{"undefined-write", "undefined\n", ""},
		{"object", "after\n", ""},
		{"nominal-slot", "after\n", "record { x: number; y: Value; }"},
		{"nominal-subclass-slot", "before\n", ""},
		{"undefined-sentinel", "undefined\n", "Real"},
		{"object-invariant", "3\n", "record { x: number; y: { label: number; }; }"},
		{"object-sentinel", "after\n", "Sentinel"},
		{"optional-literal", "3\n", "record { x: number; y: undefined; }"},
		{"tag", "3\n", ""},
		{"reduced-never", "done\n", ""},
		{"unreachable", "done\n", ""},
		{"fresh", "done\n", ""},
		{"class", "3\n", ""},
		{"record", "3\n", ""},
		{"sentinel", "3\n", "Sentinel"},
		{"stored", "3\n", "record { x: number; }"},
		{"optional", "3\n", ""},
		{"literal", "3\n", "record { x: number; y: 2; }"},
		{"string", "after\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/lower/testdata/optional_widening/checked_writes", probe.name+".a")
			observation, err := os.ReadFile(filepath.Join(repository, "internal/lower/testdata/optional_widening/checked_writes", probe.name+".expected"))
			if err != nil {
				t.Fatal(err)
			}
			if string(observation) != probe.node {
				t.Fatal("Node record differs")
			}
			if diff := disagreement(run{stdout: observation}, onNode(t, path)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(probe.node)}
			if probe.failure != "" {
				diagnostic, err := os.ReadFile(filepath.Join(repository, "internal/lower/testdata/optional_widening/checked_writes", probe.name+".stderr"))
				if err != nil {
					t.Fatal(err)
				}
				want = run{exitCode: 70, stderr: diagnostic}
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s; got %#v", diff, got)
				}
			}
		})
	}
}

func TestOptionalNeverWrite(t *testing.T) {
	source := "let value: {x:number;y:number} | boolean = false;\nfunction change():void{value={x:1,y:2};}\nchange();\nif(typeof value !== 'boolean'){\n    (value as {x:number;y?:number}).y=3;\n    console.log('continued');\n}\n"
	path := filepath.Join(t.TempDir(), "never-write.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(run{stdout: []byte("continued\n")}, onNode(t, path)); diff != "" {
		t.Fatal("Node: " + diff)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: unreachable expression value at never-write.a:5:6\n")}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s; got %#v", diff, got)
		}
	}
}

func init() {
	for _, name := range []string{"uninitialized-slot", "dictionary-own-slot", "second-optional", "boolean-write", "fresh-undefined", "fresh-boolean", "undefined-write", "object", "nominal-subclass-slot", "tag", "reduced-never", "unreachable", "fresh", "class", "record", "optional", "string"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/lower/testdata/optional_widening/checked_writes/" + name + ".a", lowers: true})
	}
}
