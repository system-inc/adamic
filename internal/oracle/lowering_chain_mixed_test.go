package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoweringChainMixedFileAgreement(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for file, source := range map[string]string{
		"main.ts":       `import { value } from './value.a'; console.log(value.toString());`,
		"value.a":       `export const value: number = 42;`,
		"tsconfig.json": `{"compilerOptions":{"strict":false,"module":"NodeNext","noEmit":true},"files":["main.ts"],"references":[]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(directory, "main.ts")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "42\n" {
		t.Fatalf("Node control: %+v", truth)
	}
	sanitized, binary := nativelyUncached(t, program)
	for mode, result := range map[string]run{"JavaScript": onJavaScriptBackend(t, program), "sanitized": sanitized, "release": releasedUncached(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s", mode, difference)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
