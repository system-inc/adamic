package oracle

import (
	"fmt"
	"testing"
)

func TestCheckedViewOptionalHostMethods(t *testing.T) {
	for _, directory := range []string{"realpath-binding", "directory-condition"} {
		variants := []string{"good", "alias", "absent", "undefined", "wrong-type", "wrong-signature", "wrong-signature-read"}
		if directory == "realpath-binding" {
			variants = append(variants, "identity")
		}
		for _, variant := range variants {
			t.Run(directory+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/later-ranked-callables/"+directory+"/"+variant)
				truth := onNode(t, path)
				actual, binary := nativelyUncached(t, program)
				runs := []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)}
				if variant == "good" || variant == "alias" || variant == "absent" || variant == "undefined" || variant == "identity" {
					if truth.exitCode != 0 {
						t.Fatalf("Node: %#v", truth)
					}
					for _, got := range runs {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatalf("%s; exit=%d stdout=%q stderr=%q", difference, got.exitCode, got.stdout, got.stderr)
						}
					}
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
					return
				}
				t.Logf("Node negative control: exit=%d stdout=%q stderr=%q", truth.exitCode, truth.stdout, truth.stderr)
				member, expected := "realpath", "(path: string) => string"
				if directory == "directory-condition" {
					member, expected = "directoryExists", "(path: string) => boolean"
				}
				message := fmt.Sprintf("adamic: panic: field read failed: host.%s expected %s, found function with incompatible result representation\n", member, expected)
				if variant == "wrong-type" {
					union := "(" + expected + ") | undefined"
					message = fmt.Sprintf("adamic: panic: field read failed: host.%s matches no member of %s; expected %s, found number\n", member, union, union)
				}
				for _, got := range runs {
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
						t.Fatalf("callable pin: exit=%d stdout=%q stderr=%q; want %q", got.exitCode, got.stdout, got.stderr, message)
					}
				}
			})
		}
	}
}
