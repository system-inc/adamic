package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestDefiniteAssignmentIsRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, location string }{
		{"definite-field.a", "class Box { n!: number; }\nconsole.log(`${new Box().n + 1}`);\n", "main.a:1:14:"},
		{"definite-local.a", "let n!: number;\nconsole.log(`${n + 1}`);\n", "main.a:1:6:"},
		{"definite-field-assigned.a", "class Box { n!: number; constructor() { this.n = 2; } }\nconsole.log(`${new Box().n + 1}`);\n", "main.a:1:14:"},
		{"definite-local-assigned.a", "let n!: number;\nn = 2;\nconsole.log(`${n + 1}`);\n", "main.a:1:6:"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("got %v, want definite assignment refusal", err)
			}
			for _, text := range []string{probe.location, "definite assignment assertion", "initialize", "constructor", "T | undefined"} {
				if !strings.Contains(refused.Error(), text) {
					t.Errorf("got %v, want %q", err, text)
				}
			}
		})
	}
}

func TestDefiniteAssignmentSoundNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class Box { n: number = 2; }\nconsole.log(`${new Box().n + 1}`);\n",
		"let n: number;\nn = 2;\nconsole.log(`${n + 1}`);\n",
		"class Box { n: number; constructor() { this.n = 2; } }\nconsole.log(`${new Box().n + 1}`);\n",
		"// This mentions ts-ignore in prose.\nconsole.log('@ts-ignore');\n",
	} {
		_, err := lowerSource(t, source)
		if err != nil {
			t.Errorf("sound neighbor: %v", err)
		}
	}
}
