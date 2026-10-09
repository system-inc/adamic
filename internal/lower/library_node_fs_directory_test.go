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
