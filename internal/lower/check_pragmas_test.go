package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestCheckPragmasAreRefused(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"ts-nocheck", "ts-check"} {
		for _, prefix := range []string{
			"// @" + name + "\n",
			"/// @" + name + " reason\n",
			"// license\n\n\t// @" + name + "\n",
			"#!/usr/bin/env node\n// @" + name + "\n",
			"// @" + name + "\n// @ts-check\n// @ts-nocheck\n",
		} {
			test.Run(name+"/"+prefix, func(test *testing.T) {
				test.Parallel()
				_, lowerSourceError := lowerSource(test, prefix+"console.log('ok');\n")
				var refused *Refused
				if !errors.As(lowerSourceError, &refused) {
					test.Fatalf("got %v, want pragma refusal", lowerSourceError)
				}
				location := "main.a:1:1:"
				if strings.HasPrefix(prefix, "// license") {
					location = "main.a:3:2:"
				}
				if strings.HasPrefix(prefix, "#!") {
					location = "main.a:2:1:"
				}
				for _, text := range []string{location, "@" + name, "pragma", "remove it", "fix"} {
					if !strings.Contains(refused.Error(), text) {
						test.Errorf("got %v, want %q", lowerSourceError, text)
					}
				}
			})
		}
	}
}

func TestCheckPragmaNeighborsCompile(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"// Explain ts-nocheck and @ts-nocheck in prose.\nconsole.log('@ts-nocheck');\n",
		"// Explain @ts-check in prose.\nconsole.log('@ts-check');\n",
		"/* @ts-nocheck */\nconsole.log('ok');\n",
		"/**\n * @ts-check\n */\nconsole.log('ok');\n",
		"console.log('ok');\n// @ts-nocheck\n",
	} {
		if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
			test.Errorf("neighbor: %v", lowerSourceError)
		}
	}
}

// The loader presents .a files to tsgo as TypeScript. Check the checker itself, before lowering.
func TestTsgoHonorsNoCheckInAdamicFiles(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	const wrong = "const b: boolean = 2;\nconsole.log(`${b}`);\n"
	if writeError := os.WriteFile(path, []byte(wrong), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if _, loadError := load.Load([]string{path}); loadError == nil {
		test.Fatal("checker accepted boolean = 2 without pragma")
	}
	if writeError := os.WriteFile(path, []byte("// @ts-nocheck\n"+wrong), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if _, loadError := load.Load([]string{path}); loadError != nil {
		test.Fatalf("tsgo did not honor @ts-nocheck in .a: %v", loadError)
	}
	_, lowerSourceError := lowerSource(test, "// @ts-nocheck\n"+wrong)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) {
		test.Fatalf("got %v, want pragma refusal", lowerSourceError)
	}
}
