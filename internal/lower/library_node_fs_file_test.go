package lower

import (
	"strings"
	"testing"
)

func TestNodeFSFileOptionsBorrow(t *testing.T) {
	for _, source := range []string{
		`import {statSync as status} from 'node:fs'; const options={throwIfNoEntry:false}; const result=status('missing',options);`,
		`import {unlinkSync} from 'node:fs'; function remove(path:string):void {return unlinkSync(path);} remove('missing');`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNodeFSFileDoesNotAuthorizeMutableWidening(t *testing.T) {
	_, err := lowerSource(t, `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false}; const wider:{throwIfNoEntry?:boolean | undefined}=options; wider.throwIfNoEntry=undefined; statSync('missing',options);`)
	if err == nil || !strings.Contains(err.Error(), "invariant-mutable") {
		t.Fatalf("want mutable widening refusal, got %v", err)
	}
}

func TestNodeFSFileRefusesOptionEffects(t *testing.T) {
	_, err := lowerSource(t, `import {statSync} from 'node:fs'; function option():boolean {console.log('effect');return false;} const result=statSync('missing',{throwIfNoEntry:option()});`)
	if err == nil || !strings.Contains(err.Error(), "evaluated expressions") {
		t.Fatalf("want effects refusal, got %v", err)
	}
}

func TestNodeFSFileRefusesVoidValues(t *testing.T) {
	_, err := lowerSource(t, `import {closeSync} from 'node:fs'; const result=closeSync(0)===undefined;`)
	if err == nil || !strings.Contains(err.Error(), "fs void calls used as values") {
		t.Fatalf("want void value refusal, got %v", err)
	}
}
