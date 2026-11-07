package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeLibraryNamesUnimplementedMembers(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ source, name string }{
		{`import {copyFileSync as copy} from 'node:fs'; copy('x','y');`, "node:fs.copyFileSync"},
		{`import * as fs from 'node:fs'; fs.readFile('x',()=>{});`, "node:fs.readFile"},
		{`import {toNamespacedPath} from 'node:path'; toNamespacedPath('x');`, "node:path.toNamespacedPath"},
		{`import {userInfo} from 'node:os'; userInfo();`, "node:os.userInfo"},
		{`import {randomUUID} from 'node:crypto'; randomUUID();`, "node:crypto.randomUUID"},
		{`import {Buffer} from 'node:buffer'; Buffer.byteLength('x');`, "node:buffer.BufferConstructor.byteLength"},
		{`import {sep} from 'node:path'; console.log(sep);`, "node:path.sep"},
		{`import {readFile} from 'node:fs'; const saved=readFile;`, "node:fs.readFile"},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, one.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, one.name) {
				test.Fatalf("want named NotYet %s, got %v", one.name, err)
			}
		})
	}
}

func TestNodeLibraryDoesNotRefuseTypeOnlyOrUserNames(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import type {Stats} from 'node:fs'; type Size=Pick<Stats,'size'>; console.log('okay');`,
		`import type {Stats} from 'node:fs'; function existsSync(path:string):boolean {return path==='x';} console.log(existsSync('x')?'yes':'no');`,
	} {
		lowersAndAgreesWithNode(test, source)
	}
}

// Behavior cannot observe checker member ownership for uncalled typed parameters.
func TestNodeLibraryDistinguishesReceiverOwners(test *testing.T) {
	test.Parallel()
	file := filepath.Join(test.TempDir(), "main.a")
	source := `import type {Stats,Dirent} from 'node:fs'; function stats(value:Stats):void{value.isFile();} function dirent(value:Dirent):void{value.isFile();}`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{file})
	if err != nil {
		test.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	lowering := lowering{program: program, checker: checked}
	names := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			names[lowering.nodeLibraryMember(node)] = true
		}
		node.ForEachChild(visit)
		return false
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	for _, name := range []string{"node:fs.StatsBase.isFile", "node:fs.Dirent.isFile"} {
		if !names[name] {
			test.Fatalf("missing distinct member %s in %v", name, names)
		}
	}
}

func TestNodeLibraryQualifiedTypeAssertion(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`import type {Stats} from 'node:fs'; const error=(new Error('plain') as NodeJS.ErrnoException); console.log('checked');`,
	)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/no-optional-widening") {
		test.Fatalf("want the qualified name checked without inventing optional fields, got %v", err)
	}
}
