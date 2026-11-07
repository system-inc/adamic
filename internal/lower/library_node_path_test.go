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
		if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
	}
	_, lowerSourceError := lowerSource(
		test,
		`import {basename} from 'node:path'; const names:[string]=['x']; console.log(basename(...names));`,
	)
	var notYet *NotYet
	if !errors.As(lowerSourceError, &notYet) || !strings.Contains(notYet.What, "node:path.basename") {
		test.Fatalf("want named NotYet, got %v", lowerSourceError)
	}
}
