package lower

import (
	"errors"
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

func TestNodeFSFileRefusesPinnedUnsupportedOverloads(t *testing.T) {
	for _, one := range []struct{ source, member string }{
		{`import {readFileSync} from 'node:fs'; readFileSync('x','latin1');`, "readFileSync"},
		{`import {openSync} from 'node:fs'; openSync('x',0);`, "openSync"},
		{`import {statSync} from 'node:fs'; statSync('x',{bigint:true});`, "statSync"},
		{`import {readSync} from 'node:fs'; readSync(0,new Uint8Array(1),0,1,null);`, "readSync"},
		{`import {writeFileSync} from 'node:fs'; writeFileSync('x',new Uint8Array(1));`, "writeFileSync"},
	} {
		t.Run(one.member, func(t *testing.T) {
			_, err := lowerSource(t, one.source)
			var missing *NotYet
			if !errors.As(err, &missing) || !strings.Contains(missing.What, one.member) {
				t.Fatalf("want named NotYet %s, got %v", one.member, err)
			}
		})
	}
}

func TestNodeFSFileNamespaceImport(t *testing.T) {
	if _, err := lowerSource(t, `import * as fs from 'node:fs'; fs.existsSync('x');`); err != nil {
		t.Fatal(err)
	}
}

func TestNodeFSFileQualifiedErrorType(t *testing.T) {
	_, err := lowerSource(t, `import type {Stats} from 'node:fs'; const error=new Error('plain'); const code=(error as NodeJS.ErrnoException).code; console.log(code??'missing');`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/no-optional-widening") {
		t.Fatalf("want optional host fields proven before the qualified cast, got %v", err)
	}
}

func TestNodeFSFileKeepsDetachedMethodRefusal(t *testing.T) {
	_, err := lowerSource(t, `import {statSync} from 'node:fs'; const stat=statSync('x'); const method=stat.isFile; console.log(method()?'file':'other');`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "isFile") {
		t.Fatalf("want existing unbound-method refusal naming isFile, got %v", err)
	}
}

func TestNodeFSFileBufferBorrow(t *testing.T) {
	if _, err := lowerSource(t, `import {readSync,writeFileSync} from 'node:fs'; import {Buffer} from 'node:buffer'; const bytes=Buffer.from('x'); readSync(0,bytes,0,1,null); writeFileSync('x',bytes);`); err != nil {
		t.Fatal(err)
	}
}

func TestNodeFSFileScratchOptionsBorrow(t *testing.T) {
	for _, source := range []string{
		`import {rmSync} from 'node:fs'; const options={recursive:true,force:true}; rmSync('missing',options);`,
		`import {mkdtempSync} from 'node:fs'; mkdtempSync('prefix',{encoding:'utf8'});`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNodeFSFileScratchOverloadsAreNamed(t *testing.T) {
	for _, one := range []struct{ source, member string }{
		{`import {mkdtempSync} from 'node:fs'; mkdtempSync('prefix','buffer');`, "mkdtempSync"},
		{`import {rmSync} from 'node:fs'; rmSync('missing',{maxRetries:2});`, "rmSync"},
		{`import {rmSync} from 'node:fs'; function flag():boolean {console.log('effect');return true;} rmSync('missing',{force:flag()});`, "rmSync"},
	} {
		_, err := lowerSource(t, one.source)
		var missing *NotYet
		if !errors.As(err, &missing) || !strings.Contains(err.Error(), one.member) {
			t.Fatalf("%s: want named NotYet, got %v", one.member, err)
		}
	}
}
