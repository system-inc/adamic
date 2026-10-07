package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewLazyUnread(t *testing.T) {
	for _, name := range []string{"unread", "unread-untagged", "unread-mixed"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lazy/"+name)
			truth := onNode(t, path)
			expected := "okok\n"
			if name == "unread-mixed" {
				expected = "true\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != expected {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestCheckedViewLazyHelperMutant(t *testing.T) {
	for _, name := range []string{"helper", "generic", "callback", "field"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "stage3/interface-downcasts/lazy/"+name+"-mutant.a")
			_, err := lowered(t, path)
			if err == nil {
				t.Fatal("unsupported helper read ran on")
			}
			message := err.Error()
			if !strings.Contains(message, name+"-mutant.a:9:") || !strings.Contains(message, "field opaque") || !strings.Contains(message, "callable") {
				t.Fatalf("want callable refusal at helper read, got %v", err)
			}
			t.Logf("caught helper mutant: %v", err)
		})
	}
}

func TestCheckedViewLazyMissingNameMutant(t *testing.T) {
	program, path := interfaceFixture(t, "lazy/missing-name-mutant")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "undefined\n" {
		t.Fatalf("Node source mutant: %#v", truth)
	}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		message := string(got.stderr)
		if got.exitCode != 70 || !strings.Contains(message, "name") || !strings.Contains(message, "string") {
			t.Fatalf("supported helper read ran on: %#v", got)
		}
		t.Logf("caught missing-name mutant: %s", message)
	}
}
