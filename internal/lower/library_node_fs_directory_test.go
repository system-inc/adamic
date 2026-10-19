package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeFSDirectorySymlinkSignatures(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import {symlinkSync} from 'node:fs'; symlinkSync('target','link');`,
		`import {symlinkSync} from 'node:fs'; symlinkSync('target','link','dir');`,
		`import {symlinkSync} from 'node:fs'; function link():void {return symlinkSync('target','link',null);}`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{
		`import {symlinkSync} from 'node:fs'; function link(type:'file'|'dir'|'junction'):void {symlinkSync('target','link',type);}`,
		`import {symlinkSync} from 'node:fs'; import {Buffer} from 'node:buffer'; symlinkSync(Buffer.from('target'),'link');`,
		`import {symlinkSync} from 'node:fs'; const value=symlinkSync('target','link');`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "node:fs.symlinkSync") {
			t.Fatalf("want named NotYet, got %v", err)
		}
	}
}

func TestNodeFSDirectoryRealpathClosureBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; console.log(resolve('.'));`,
		`import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; function invoke(){return resolve('.');} console.log(invoke());`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	// Main's closed-use walk rejects aliases before inspecting their calls.
	for _, one := range []struct{ source, reason string }{
		{`import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; const alias=resolve; console.log(alias('.'));`, "stored in another binding"},
		{`import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; console.log(resolve('.',{encoding:'buffer'}).toString());`, "zero or multiple arguments"},
		{`import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; const alias=resolve; console.log(alias('.','utf8'));`, "stored in another binding"},
		{`import {realpathSync} from 'node:fs'; const resolve=realpathSync.native; function invoke(){return resolve('.',{encoding:'buffer'});} console.log(invoke().toString());`, "zero or multiple arguments"},
	} {
		_, err := lowerSource(t, one.source)
		var missing *NotYet
		if !errors.As(err, &missing) || !strings.Contains(missing.What, "node:fs.realpathSync.native as a value") || !strings.Contains(missing.What, one.reason) {
			t.Fatalf("want named closed-use refusal %s, got %v", one.reason, err)
		}
	}
}
