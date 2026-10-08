package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Full original tuple declarations still refuse before either backend. These
// Node controls are visible gaps, never runtime pair certification.
func TestCheckedViewOriginalEmitTupleFrontier(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, variant := range []string{"", "-empty", "-tuple", "-wrong"} {
		t.Run(variant, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/lane4/primitive-original/emit-tuple-frontier" + variant + ".a")
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(directory, "compiler/builder.d.ts"))), 1)
			file := filepath.Join(t.TempDir(), "tuple.a")
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			text := map[string]string{"": "string", "-empty": "object", "-tuple": "object", "-wrong": "boolean"}[variant] + "\n"
			if diff := disagreement(run{stdout: []byte(text)}, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			_, err = lowered(t, file)
			if err == nil || !strings.Contains(err.Error(), "can't lower a tuple element of type union of differently held members yet") {
				t.Fatalf("original tuple frontier: %v", err)
			}
			t.Log("original tuple member selection remains uncredited; helper refuses before unchecked read")
		})
	}
}
