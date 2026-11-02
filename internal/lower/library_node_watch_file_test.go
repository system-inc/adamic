package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

const watchSource = `import * as fs from 'node:fs';
function observe(curr: fs.Stats, prev: fs.Stats): void { console.log(String(curr.mtimeMs) + " " + String(prev.size)); }
fs.watchFile('x', {persistent:true, interval:250}, observe);
fs.unwatchFile('x', observe);
fs.unwatchFile('x');`

func TestNodeWatchFilePollingForms(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, watchSource)
	if err != nil {
		t.Fatal(err)
	}
	if !program.HasPromises() {
		t.Fatal("watchFile must run the loop without an explicit Promise")
	}
	operations := map[string]int{}
	walk(program.Main, func(node any) bool {
		if call, ok := node.(ir.NodeFSFile); ok {
			operations[call.Operation]++
		}
		return true
	})
	if operations["host_watch_file"] != 1 || operations["host_unwatch_file"] != 2 {
		t.Fatal(operations)
	}
}
func TestNodeWatchFileOtherFormsRefused(t *testing.T) {
	t.Parallel()
	cases := []struct{ code, want string }{
		{`const watcher = fs.watchFile('x', {persistent:true, interval:5}, observe);`, "used as a value"},
		{`fs.watchFile('x', observe);`, "outside the stat-polling signature"},
		{`fs.watchFile('x', {persistent:true, interval:5, bigint:true}, (_c:fs.BigIntStats,_p:fs.BigIntStats):void=>{});`, "options"},
		{`fs.watchFile('x', {persistent:true, interval:5}, (_c:fs.Stats,_p:fs.Stats):number=>1);`, "listener outside"},
		{`fs.watchFile('x', {persistent:true, interval:5}, (_c:fs.Stats):void=>{});`, "listener outside"},
		{`fs.watchFile?.('x', {persistent:true, interval:5}, observe);`, "optional call"},
	}
	prefix := `import * as fs from 'node:fs'; function observe(_c:fs.Stats,_p:fs.Stats):void{}`
	for _, c := range cases {
		_, err := lowerSource(t, prefix+c.code)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: want %q, got %v", c.code, c.want, err)
		}
	}
}

func TestNodeWatchFileTscCloseShape(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `import * as fs from 'node:fs';
 function start(path:string): {close:()=>void} {
 fs.watchFile(path,{persistent:true,interval:250}, changed);
 return {close:()=>fs.unwatchFile(path,changed)};
 function changed(curr:fs.Stats,prev:fs.Stats):void {console.log(path+String(curr.size-prev.size));}
 }
 start('x').close();`)
	if err != nil {
		t.Fatal(err)
	}
}
func TestNodeWatchFileVoidViewCannotHideOwnedResult(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `import * as fs from 'node:fs';
 function stringResult(_curr:fs.Stats,_prev:fs.Stats):string {return 'owned';}
 const listener:(curr:fs.Stats,prev:fs.Stats)=>void=stringResult;
 fs.watchFile('x',{persistent:true,interval:250},listener);`)
	if err == nil || !strings.Contains(err.Error(), "listener outside") {
		t.Fatal(err)
	}
}

func TestNodeWatchFileUsesLibrarySymbols(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import {watchFile as poll,unwatchFile as stop,Stats} from 'node:fs';
 function changed(_curr:Stats,_prev:Stats):void{}
 poll('x',{persistent:true,interval:250},changed);stop('x',changed);`,
		`function watchFile(text:string):string{return text;}console.log(watchFile('own'));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
