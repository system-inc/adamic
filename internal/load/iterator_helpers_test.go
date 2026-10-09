package load

import (
	"os"
	"strings"
	"testing"
)

func TestProjectIteratorHelpersFollowLib(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/library_iterator_to_array.a")
	if err != nil {
		t.Fatal(err)
	}
	// The oracle explicitly includes this lib. Projects below select it in config instead.
	text := strings.TrimPrefix(string(source), "/// <reference lib=\"es2025.iterator\" />\n")
	for _, lib := range []string{"es2024", "es2024\",\"es2025.iterator"} {
		t.Run(lib, func(t *testing.T) {
			t.Parallel()
			paths := projectProgram(t, `{"strict":true,"target":"es2024","lib":["`+lib+`"],"types":[]}`, text)
			_, err := Load(paths)
			if lib == "es2024" {
				if err == nil || !strings.Contains(err.Error(), "toArray") {
					t.Fatalf("want absent iterator helper diagnostics, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
