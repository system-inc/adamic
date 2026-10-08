package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewLazyUnread(t *testing.T) {
	for _, name := range []string{"unread", "unread-untagged", "ordinary-disjoint", "optional-unread"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lazy/"+name)
			truth := onNode(t, path)
			expected := "okok\n"
			if name == "optional-unread" {
				expected = "identifier\n"
			}
			if name == "ordinary-disjoint" {
				expected = "okok\n3\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != expected {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestCheckedViewLazyHelperMutant(t *testing.T) {
	for _, name := range []string{"helper", "generic", "callback", "field", "call"} {
		t.Run(name, func(t *testing.T) {
			path, pathError := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lazy/"+name+"-mutant.a"))
			if pathError != nil {
				t.Fatal(pathError)
			}
			truth := onNode(t, path)
			expected := "must not run\n"
			if name == "call" {
				expected = "3\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != expected || len(truth.stderr) != 0 {
				t.Fatalf("source Node pin: %#v", truth)
			}
			_, err := lowered(t, path)
			if err == nil {
				t.Fatal("unsupported helper read ran on")
			}
			message := err.Error()
			guard := strings.Contains(message, "field opaque") && strings.Contains(message, "callable")
			if name == "call" {
				guard = strings.Contains(message, "a class method through a view that erases its prototype origin")
			}
			if !strings.Contains(message, name+"-mutant.a:9:") || !guard {
				t.Fatalf("want callable refusal at helper read, got %v", err)
			}
			t.Logf("caught helper mutant: %v", err)
		})
	}
}
