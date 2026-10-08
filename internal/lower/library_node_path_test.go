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

func TestNodePathHost29Signatures(t *testing.T) {
	for _, source := range []string{
		`import {normalize,extname,isAbsolute,sep} from 'node:path'; console.log(normalize('a/../'));console.log(extname('x.txt'));console.log(isAbsolute('/')?'yes':'no');console.log(sep);`,
		`import * as path from 'node:path'; console.log(path.posix.normalize('/a/../')); console.log(path.posix.sep);`,
		`import {posix as unix} from 'node:path'; console.log(unix.extname('x.c'));`,
		`import * as unix from 'node:path/posix'; console.log(unix.join('a','b'));`,
		`import type {ParsedPath} from 'node:path'; console.log('type only');`,
		`import type {win32} from 'node:path'; type Windows=typeof win32; console.log('type only');`,
		`import {normalize,sep} from 'path'; console.log(normalize('a/..')); console.log(sep);`,
		`import * as unix from 'path/posix'; console.log(unix.dirname('/x'));`,
		`const win32={normalize:(value:string):string=>value}; console.log(win32.normalize('user'));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("%s: %v", source, err)
		}
	}
}

func TestNodePathHost29Win32Refusals(t *testing.T) {
	for _, source := range []string{
		`import * as path from 'node:path'; console.log(path.win32.normalize('C:/x'));`,
		`import {win32 as windows} from 'node:path'; console.log(windows.basename('C:/x'));`,
		`import * as windows from 'node:path/win32'; console.log(windows.dirname('C:/x'));`,
		`import {basename} from 'node:path/win32'; console.log(basename('C:/x'));`,
		`import * as path from 'node:path'; const windows=path.win32; console.log(windows.sep);`,
		`import {win32 as windows} from 'path'; console.log(windows.extname('C:/x.c'));`,
		`import * as windows from 'path/win32'; console.log(windows.join('C:/','x'));`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "Windows path semantics are not implemented") {
			t.Fatalf("%s: want explicit win32 refusal, got %v", source, err)
		}
	}
}
