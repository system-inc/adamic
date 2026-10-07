package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeLibraryNamesUnimplementedMembers(t *testing.T) {
	for _, one := range []struct{ source, name string }{
		{`import {copyFileSync as copy} from 'node:fs'; copy('x','y');`, "node:fs.copyFileSync"},
		{`import * as fs from 'node:fs'; fs.readFile('x',()=>{});`, "node:fs.readFile"},
		{`import {toNamespacedPath} from 'node:path'; toNamespacedPath('x');`, "node:path.toNamespacedPath"},
		{`import {userInfo} from 'node:os'; userInfo();`, "node:os.userInfo"},
		{`import {randomUUID} from 'node:crypto'; randomUUID();`, "node:crypto.randomUUID"},
		{`import {Buffer} from 'node:buffer'; Buffer.byteLength('x');`, "node:buffer.byteLength"},
		{`import {sep} from 'node:path'; console.log(sep);`, "node:path.sep"},
		{`import {readFile} from 'node:fs'; const saved=readFile;`, "node:fs.readFile"},
	} {
		t.Run(one.name, func(t *testing.T) {
			_, err := lowerSource(t, one.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, one.name) {
				t.Fatalf("want named NotYet %s, got %v", one.name, err)
			}
		})
	}
}

func TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames(t *testing.T) {
	for _, source := range []string{
		`import type {Stats} from 'node:fs'; type Size=Pick<Stats,'size'>; console.log('okay');`,
		`import type {Stats} from 'node:fs'; function existsSync(path:string):boolean {return path==='x';} console.log(existsSync('x')?'yes':'no');`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
