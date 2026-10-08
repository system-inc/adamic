package oracle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		"emit-helper-frontier":                "a field of type string | ((node: EmitHelperUniqueNameCallback) => string)",
		"incremental-tuple-element-frontier":  "a cast the runtime can't check",
		"jsdoc-parent-union-frontier":         "checked view read of field kind with unsupported union intersection contract",
		"jsdoc-parent-optional-frontier":      "checked view read of field kind with unsupported union intersection contract",
		"build-options-key-frontier":          "a cast the runtime can't check",
		"incremental-root-frontier":           "an array of IncrementalBuildInfoRoot",
		"incremental-signature-frontier":      "an array of IncrementalBuildInfoEmitSignature",
		"incremental-out-signature-frontier":  "checked view read of field outSignature with unsupported tuple union member contract",
		"incremental-bundle-each-frontier":    "an array of IncrementalBundleEmitBuildInfoFileInfo",
		"incremental-multi-each-frontier":     "an array of IncrementalMultiFileEmitBuildInfoFileInfo",
		"incremental-signature-each-frontier": "an array of IncrementalBuildInfoEmitSignature",
		"incremental-union-each-frontier":     "an array of string | FileInfo | IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo",
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
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane4b/original/" + probe.name + ".a")
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
			_, err = lower.Lower(context.Background(), loaded)
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
