package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

const watchSource = `import * as fs from 'node:fs';
function observe(curr: fs.Stats, prev: fs.Stats): void { console.log(String(curr.mtimeMs) + " " + String(prev.size)); }
fs.watchFile('x', {persistent:true, interval:250}, observe);
fs.unwatchFile('x', observe);
fs.unwatchFile('x');`

func TestNodeWatchFilePollingForms(test *testing.T) {
	test.Parallel()
	program, err := lowerSource(test, watchSource)
	if err != nil {
		test.Fatal(err)
	}
	if !program.HasPromises() {
		test.Fatal("watchFile must run the loop without an explicit Promise")
	}
	operations := map[string]int{}
	walk(program.Main, func(node any) bool {
		if call, ok := node.(ir.NodeFSFile); ok {
			operations[call.Operation]++
		}
		return true
	})
	if operations["host_watch_file"] != 1 || operations["host_unwatch_file"] != 2 {
		test.Fatal(operations)
	}
}
func TestNodeWatchFileOtherFormsRefused(test *testing.T) {
	test.Parallel()
	cases := []struct{ code, want string }{
		{`const watcher = fs.watchFile('x', {persistent:true, interval:5}, observe);`, "used as a value"},
		{`fs.watchFile('x', observe);`, "outside the stat-polling signature"},
		{
			`fs.watchFile('x', {persistent:true, interval:5, bigint:true}, (_c:fs.BigIntStats,_p:fs.BigIntStats):void=>{});`,
			"options",
		},
		{`fs.watchFile('x', {persistent:true, interval:5}, (_c:fs.Stats,_p:fs.Stats):number=>1);`, "listener outside"},
		{`fs.watchFile('x', {persistent:true, interval:5}, (_c:fs.Stats):void=>{});`, "listener outside"},
		{`fs.watchFile?.('x', {persistent:true, interval:5}, observe);`, "optional call"},
	}
	prefix := `import * as fs from 'node:fs'; function observe(_c:fs.Stats,_p:fs.Stats):void{}`
	for _, c := range cases {
		_, err := lowerSource(test, prefix+c.code)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			test.Errorf("%s: want %q, got %v", c.code, c.want, err)
		}
	}
}

func TestNodeWatchFileTscCloseShape(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `import * as fs from 'node:fs';
 function start(path:string): {close:()=>void} {
 fs.watchFile(path,{persistent:true,interval:250}, changed);
 return {close:()=>fs.unwatchFile(path,changed)};
 function changed(curr:fs.Stats,prev:fs.Stats):void {console.log(path+String(curr.size-prev.size));}
 }
 start('x').close();`)
	if err != nil {
		test.Fatal(err)
	}
}
func TestNodeWatchFileVoidViewCannotHideOwnedResult(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `import * as fs from 'node:fs';
 function stringResult(_curr:fs.Stats,_prev:fs.Stats):string {return 'owned';}
 const listener:(curr:fs.Stats,prev:fs.Stats)=>void=stringResult;
 fs.watchFile('x',{persistent:true,interval:250},listener);`)
	if err == nil || !strings.Contains(err.Error(), "listener outside") {
		test.Fatal(err)
	}
}

func TestNodeWatchFileUsesLibrarySymbols(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import {watchFile as poll,unwatchFile as stop,Stats} from 'node:fs';
 function changed(_curr:Stats,_prev:Stats):void{}
 poll('x',{persistent:true,interval:250},changed);stop('x',changed);`,
		`function watchFile(text:string):string{return text;}console.log(watchFile('own'));`,
	} {
		if _, err := lowerSource(test, source); err != nil {
			test.Fatal(err)
		}
	}
}

func TestNodeWatchFileCallbackViews(test *testing.T) {
	test.Parallel()
	for _, parameter := range []struct{ name, of, want string }{
		{"h11", "{blocks:number}", "node:fs.StatsBase.blocks"},
		{"any", "any", "hides a host object"},
		{"unknown", "unknown", "hides a host object"},
	} {
		test.Run(parameter.name, func(test *testing.T) {
			_, err := lowerSource(test, `import * as fs from 'node:fs';
function changed(curr:`+parameter.of+`,prev:fs.Stats):void{}
fs.watchFile('x',{persistent:true,interval:10},changed);`)
			var notYet *NotYet
			if !errors.As(err, &notYet) ||
				!strings.Contains(notYet.What, parameter.want) ||
				!strings.Contains(notYet.Where, "main.a:3:") {
				test.Fatalf("want argument-located NotYet %q, got %v", parameter.want, err)
			}
		})
	}
}
func TestNodeWatchFileMtimeUnaryPlus(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `import * as fs from 'node:fs';
 function changed(curr:fs.Stats,prev:fs.Stats):void { console.log(String(+curr.mtime === 0) + " " + String(+prev.mtime === 0) + " " + String(+curr.mtime === +prev.mtime)); }
 fs.watchFile('x',{persistent:true,interval:10},changed);
 fs.unwatchFile('x',changed);`)
	if err != nil {
		test.Fatal(err)
	}
}

func TestNodeWatchFileOptionTypeRefusals(test *testing.T) {
	test.Parallel()
	prefix := `import * as fs from 'node:fs'; function observe(_curr:fs.Stats,_prev:fs.Stats):void{}`
	for _, probe := range []struct{ name, code, want string }{
		{
			"persistent",
			`fs.watchFile('x',{persistent:undefined,interval:5},observe);`,
			"node:fs.watchFile with a non-boolean persistent",
		},
		{
			"interval",
			`fs.watchFile('x',{persistent:true,interval:undefined},observe);`,
			"node:fs.watchFile with a non-number interval",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			_, err := lowerSource(test, prefix+probe.code)
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.want {
				test.Fatalf("want exact NotYet %q, got %v", probe.want, err)
			}
		})
	}
}
