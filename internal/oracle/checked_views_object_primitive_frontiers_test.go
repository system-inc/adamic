package oracle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// These probes expose remaining dependencies; they confer no certificate.
func TestCheckedViewObjectPrimitiveRemainingFrontiers(t *testing.T) {
	declarations := os.Getenv("ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS to prepare.cjs output")
	}
	expected := map[string]string{
		"emit-helper-frontier":               "a field of type string | ((node: EmitHelperUniqueNameCallback) => string)",
		"incremental-tuple-element-frontier": "a cast the runtime can't check",
		"build-options-key-frontier":         "a cast the runtime can't check",
		"incremental-bundle-each-frontier":   "an array of IncrementalBundleEmitBuildInfoFileInfo",
		"incremental-multi-each-frontier":    "an array of IncrementalMultiFileEmitBuildInfoFileInfo",
		"incremental-union-each-frontier":    "an array of string | FileInfo | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo",
	}
	for _, probe := range []struct{ name, source string }{
		{"emit-helper-frontier", "string\n"},
		{"incremental-tuple-element-frontier", "string\n"},
		{"jsdoc-parent-union-frontier", "80\n"},
		{"jsdoc-parent-optional-frontier", "80\n"},
		{"build-options-key-frontier", "boolean\n"},
		{"incremental-root-frontier", "number\n"},
		{"incremental-signature-frontier", "number\n"},
		{"incremental-out-signature-frontier", "string\n"},
		{"incremental-bundle-each-frontier", "string\ndone\n"},
		{"incremental-multi-each-frontier", "string\ndone\n"},
		{"incremental-signature-each-frontier", "number\ndone\n"},
		{"incremental-union-each-frontier", "string\ndone\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/original/" + probe.name + ".a"))
			if err != nil {
				t.Fatal(err)
			}
			bound := string(input)
			for _, module := range []string{"types", "builder", "tsbuildPublic"} {
				bound = strings.ReplaceAll(bound, "'original-tsc-"+module+"'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler", module+".d.ts"))))
			}
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			if diff := disagreement(run{stdout: []byte(probe.source)}, onNode(t, path)); diff != "" {
				t.Fatal("source Node: " + diff)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if probe.name == "incremental-root-frontier" || probe.name == "incremental-signature-frontier" || probe.name == "incremental-out-signature-frontier" || probe.name == "incremental-signature-each-frontier" {
				if err != nil {
					t.Fatal(err)
				}
				objectPrimitiveOriginalCount(t, program, probe.name+".a")
				actual, binary := nativelyUncached(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(run{stdout: []byte(probe.source)}, got); diff != "" {
						t.Fatal("supported original tuple consumer: " + diff)
					}
				}
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			if strings.HasPrefix(probe.name, "jsdoc-parent-") {
				if err != nil {
					t.Fatal(err)
				}
				// The complete intersection is now supported. This reduced producer
				// still lacks required Node fields and must stop at the parent read.
				missing := "flags is not initialized; expected NodeFlags"
				if probe.name == "jsdoc-parent-optional-frontier" {
					missing = "escapedText is not initialized; expected __String"
				}
				want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: node.parent." + missing + ", found missing\n")}
				actual, _ := nativelyUncached(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("parent obligation: %s; got %#v", diff, got)
					}
				}
				if count := changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == "node.parent" }, func(read ir.Property) ir.Property { read.View = ""; return read }); count != 1 {
					t.Fatalf("want one parent-read mutant, got %d", count)
				}
				mutated, _ := nativelyUncached(t, program)
				for _, got := range []run{mutated, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(run{stdout: []byte(probe.source)}, got); diff != "" {
						t.Fatalf("parent mutant must execute original Node value: %s; got %#v", diff, got)
					}
					t.Log("parent-read omission caught by exact missing-field refusal after valid execution")
				}
				return
			}
			switch failure := err.(type) {
			case *lower.Refused:
				if failure.What != expected[probe.name] {
					t.Fatalf("frontier changed: %v", err)
				}
				t.Log("blocked original contract: " + failure.What)
			case *lower.NotYet:
				if failure.What != expected[probe.name] {
					t.Fatalf("frontier changed: %v", err)
				}
				t.Log("blocked original consumer: " + failure.What)
			default:
				t.Fatalf("frontier changed; certify before claiming it: %v", err)
			}
		})
	}
}
