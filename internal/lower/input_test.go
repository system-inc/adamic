package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A tuple spread into one of the prelude's doors is a count the checker accepts, so it must come
// back NotYet naming the door, never as an internal error blaming the checker.
func TestInputSpreadArgumentsAreNotYet(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"readTextFile":     "import { readTextFile } from 'adamic';\nconst path: [string] = ['in.txt'];\nreadTextFile(...path);\n",
		"writeTextFile":    "import { writeTextFile } from 'adamic';\nconst args: [string, string] = ['out.txt', 'text'];\nwriteTextFile(...args);\n",
		"utf8Length":       "import { utf8Length } from 'adamic';\nconst text: [string] = ['abc'];\nutf8Length(...text);\n",
		"utf8At":           "import { utf8At } from 'adamic';\nconst args: [string, number] = ['abc', 1];\nutf8At(...args);\n",
		"readDirectory":    "import { readDirectory } from 'adamic';\nconst path: [string] = ['.'];\nreadDirectory(...path);\n",
		"fileStatus":       "import { fileStatus } from 'adamic';\nconst path: [string] = ['.'];\nfileStatus(...path);\n",
		"programArguments": "import { programArguments } from 'adamic';\nconst none: [] = [];\nprogramArguments(...none);\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want NotYet", err)
			}
			if !strings.Contains(err.Error(), "a spread argument to "+name) {
				t.Fatalf("got %q, want it to name the spread into %s", err.Error(), name)
			}
		})
	}
}

// Keep the rejected programs as .a files so each call can also be tried with the CLI.
func TestInputSpreadCoverage(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("testdata/input-spread/*.a")
	if err != nil || len(paths) != 47 {
		t.Fatalf("fixtures: %d, %v", len(paths), err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if strings.HasSuffix(path, "_array.a") {
				var check *load.CheckError
				if !errors.As(err, &check) || !strings.Contains(err.Error(), "A spread argument must either have a tuple type or be passed to a rest parameter") {
					t.Fatalf("got %v, want the checker's array spread diagnostic", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			name := strings.Split(filepath.Base(path), "_")[0]
			var notYet *NotYet
			if strings.HasSuffix(path, "_local.a") {
				if !errors.As(err, &notYet) || strings.Contains(err.Error(), "a spread argument to ") {
					t.Fatalf("got %v, want a general NotYet without blaming a prelude function", err)
				}
				return
			}
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a spread argument to "+name) {
				t.Fatalf("%s: got %v, want NotYet naming %s", source, err, name)
			}
			line := strings.Count(string(source[:strings.LastIndex(string(source), "\n")]), "\n") + 1
			if !strings.Contains(err.Error(), fmt.Sprintf("%s:%d:", filepath.Base(path), line)) {
				t.Fatalf("missing call location: %v", err)
			}
		})
	}
}
