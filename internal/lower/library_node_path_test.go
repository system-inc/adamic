package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodePathBasenameSignatures(t *testing.T) {
	for _, source := range []string{
		`import {basename} from 'node:path'; console.log(basename('a/b.txt'));`,
		`import {basename as base} from 'node:path'; console.log(base('a/b.txt','.txt'));`,
		`import {basename} from 'node:path'; console.log(basename('a/b.txt',undefined));`,
		`import {basename} from 'node:path'; function name(path:string,suffix:string|undefined):string {return basename(path,suffix);}`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	_, err := lowerSource(t, `import {basename} from 'node:path'; const names:[string]=['x']; console.log(basename(...names));`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "node:path.basename") {
		t.Fatalf("want named NotYet, got %v", err)
	}
}
