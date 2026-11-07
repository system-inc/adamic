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
				_, err := lowerSource(test, prefix+"console.log('ok');\n")
				var refused *Refused
				if !errors.As(err, &refused) {
					test.Fatalf("got %v, want pragma refusal", err)
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
						test.Errorf("got %v, want %q", err, text)
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
		if _, err := lowerSource(test, source); err != nil {
			test.Errorf("neighbor: %v", err)
		}
	}
}

// The loader presents .a files to tsgo as TypeScript. Check the checker itself, before lowering.
func TestTsgoHonorsNoCheckInAdamicFiles(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	const wrong = "const b: boolean = 2;\nconsole.log(`${b}`);\n"
	if err := os.WriteFile(path, []byte(wrong), 0o644); err != nil {
		test.Fatal(err)
	}
	if _, err := load.Load([]string{path}); err == nil {
		test.Fatal("checker accepted boolean = 2 without pragma")
	}
	if err := os.WriteFile(path, []byte("// @ts-nocheck\n"+wrong), 0o644); err != nil {
		test.Fatal(err)
	}
	if _, err := load.Load([]string{path}); err != nil {
		test.Fatalf("tsgo did not honor @ts-nocheck in .a: %v", err)
	}
	_, err := lowerSource(test, "// @ts-nocheck\n"+wrong)
	var refused *Refused
	if !errors.As(err, &refused) {
		test.Fatalf("got %v, want pragma refusal", err)
	}
}
