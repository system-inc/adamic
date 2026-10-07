package lower

import (
	"errors"
	"strings"
	"testing"
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
