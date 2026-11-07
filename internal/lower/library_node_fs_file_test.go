package lower

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNodeFSFileOptionsBorrow(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import {statSync as status} from 'node:fs'; const options={throwIfNoEntry:false}; const result=status('missing',options); console.log(result===undefined?'absent':'present');`,
		`import {unlinkSync} from 'node:fs'; function remove(path:string):void {return unlinkSync(path);} console.log('before remove'); remove('missing');`,
	} {
		exit := 0
		if strings.Contains(source, "before remove") {
			exit = 70
		}
		lowersAndAgreesWithNodeExit(test, fsAgreementSource(test, source), exit)
	}
}

func TestNodeFSFileDoesNotAuthorizeMutableWidening(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`import {statSync} from 'node:fs'; const options={throwIfNoEntry:false}; const wider:{throwIfNoEntry?:boolean | undefined}=options; wider.throwIfNoEntry=undefined; statSync('missing',options);`,
	)
	if err == nil || !strings.Contains(err.Error(), "invariant-mutable") {
		test.Fatalf("want mutable widening refusal, got %v", err)
	}
}

func TestNodeFSFileRefusesOptionEffects(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`import {statSync} from 'node:fs'; function option():boolean {console.log('effect');return false;} const result=statSync('missing',{throwIfNoEntry:option()});`,
	)
	if err == nil || !strings.Contains(err.Error(), "evaluated expressions") {
		test.Fatalf("want effects refusal, got %v", err)
	}
}

func TestNodeFSFileRefusesVoidValues(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `import {closeSync} from 'node:fs'; const result=closeSync(0)===undefined;`)
	// H2's refusal of a void read (#x2gt39p) meets it first, before the fs-specific one.
	if err == nil || !strings.Contains(err.Error(), "a read of a void value") {
		test.Fatalf("want void value refusal, got %v", err)
	}
}

func TestNodeFSFileRefusesPinnedUnsupportedOverloads(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ source, member string }{
		{`import {readFileSync} from 'node:fs'; readFileSync('x','latin1');`, "readFileSync"},
		{`import {openSync} from 'node:fs'; openSync('x',0);`, "openSync"},
		{`import {statSync} from 'node:fs'; statSync('x',{bigint:true});`, "statSync"},
		{`import {readSync} from 'node:fs'; readSync(0,new Uint8Array(1),0,1,null);`, "readSync"},
		{`import {writeFileSync} from 'node:fs'; writeFileSync('x',new Uint8Array(1));`, "writeFileSync"},
	} {
		test.Run(one.member, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, one.source)
			var missing *NotYet
			if !errors.As(err, &missing) || !strings.Contains(missing.What, one.member) {
				test.Fatalf("want named NotYet %s, got %v", one.member, err)
			}
		})
	}
}

func TestNodeFSFileNamespaceImport(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		fsAgreementSource(test, `import * as fs from 'node:fs'; console.log(fs.existsSync('x')?'present':'absent');`),
	)
}

func TestNodeFSFileQualifiedErrorType(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`import type {Stats} from 'node:fs'; const error=new Error('plain'); const code=(error as NodeJS.ErrnoException).code; console.log(code??'missing');`,
	)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/no-optional-widening") {
		test.Fatalf("want optional host fields proven before the qualified cast, got %v", err)
	}
}

func TestNodeFSFileKeepsDetachedMethodRefusal(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`import {statSync} from 'node:fs'; const stat=statSync('x'); const method=stat.isFile; console.log(method()?'file':'other');`,
	)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "isFile") {
		test.Fatalf("want existing unbound-method refusal naming isFile, got %v", err)
	}
}

func TestNodeFSFileBufferBorrow(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, fsAgreementSource(
		test,
		`import {openSync,closeSync,readSync,readFileSync,writeFileSync,unlinkSync} from 'node:fs'; import {Buffer} from 'node:buffer'; const bytes=Buffer.from('b'); writeFileSync('x',Buffer.from('a')); const fd=openSync('x','r'); const count=readSync(fd,bytes,0,1,null); console.log(count.toString()); closeSync(fd); console.log(bytes.toString()); writeFileSync('x',bytes); console.log(readFileSync('x','utf8')); unlinkSync('x');`,
	))
}

func TestNodeFSFileScratchOptionsBorrow(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import {rmSync,existsSync} from 'node:fs'; const options={recursive:true,force:true}; rmSync('missing',options); console.log(existsSync('missing')?'present':'absent');`,
		`import {mkdtempSync,existsSync,rmSync} from 'node:fs'; const directory=mkdtempSync('prefix',{encoding:'utf8'}); console.log(existsSync(directory)?'created':'missing'); rmSync(directory,{recursive:true});`,
	} {
		source = strings.ReplaceAll(source, "'prefix'", strconv.Quote(filepath.Join(test.TempDir(), "prefix")))
		lowersAndAgreesWithNode(test, fsAgreementSource(test, source))
	}
}

func TestNodeFSFileScratchOverloadsAreNamed(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ source, member string }{
		{`import {mkdtempSync} from 'node:fs'; mkdtempSync('prefix','buffer');`, "mkdtempSync"},
		{`import {rmSync} from 'node:fs'; rmSync('missing',{maxRetries:2});`, "rmSync"},
		{
			`import {rmSync} from 'node:fs'; function flag():boolean {console.log('effect');return true;} rmSync('missing',{force:flag()});`,
			"rmSync",
		},
	} {
		_, err := lowerSource(test, one.source)
		var missing *NotYet
		if !errors.As(err, &missing) || !strings.Contains(err.Error(), one.member) {
			test.Fatalf("%s: want named NotYet, got %v", one.member, err)
		}
	}
}
