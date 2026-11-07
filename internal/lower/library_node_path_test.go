package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodePathBasenameSignatures(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import {basename} from 'node:path'; console.log(basename('a/b.txt'));`,
		`import {basename as base} from 'node:path'; console.log(base('a/b.txt','.txt'));`,
		`import {basename} from 'node:path'; console.log(basename('a/b.txt',undefined));`,
		`import {basename} from 'node:path'; function name(path:string,suffix:string|undefined):string {return basename(path,suffix);}`,
	} {
		if _, err := lowerSource(test, source); err != nil {
			test.Fatal(err)
		}
	}
	_, err := lowerSource(
		test,
		`import {basename} from 'node:path'; const names:[string]=['x']; console.log(basename(...names));`,
	)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "node:path.basename") {
		test.Fatalf("want named NotYet, got %v", err)
	}
}
