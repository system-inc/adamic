package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeFSDirectoryStatConstOptions(t *testing.T) {
	for _, source := range []string{
		`import * as fs from 'node:fs'; const statSyncOptions={throwIfNoEntry:false} as const; function stat(path:string):import('fs').Stats|undefined {return fs.statSync(path,statSyncOptions);}`,
		`import {statSync} from 'node:fs'; const options=({throwIfNoEntry:true}) as const; statSync('missing',options);`,
		`import {statSync} from 'node:fs'; statSync('missing',{throwIfNoEntry:false});`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{
		`import {statSync} from 'node:fs'; const options={bigint:true,throwIfNoEntry:false} as const; statSync('missing',options);`,
		`import {statSync} from 'node:fs'; function status(options:{throwIfNoEntry?:boolean}):void {statSync('missing',options);}`,
		`import {statSync} from 'node:fs'; console.log(String(statSync('missing',{throwIfNoEntry:false})?.isFile() ?? false));`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !(strings.Contains(notYet.What, "statSync") || strings.Contains(notYet.What, "node:fs.StatsBase.isFile")) {
			t.Fatalf("want named NotYet, got %v", err)
		}
	}
}
