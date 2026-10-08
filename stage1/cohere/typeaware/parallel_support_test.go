package typeaware

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

var typeawareRunGuard = childguard.Options{FirstOutput: 2 * time.Minute, Stall: time.Minute, Ceiling: 10 * time.Minute}

// Each child owns its output counter and directory. Parent-owned sources,
// archives, executables and emitted modules are immutable until all children end.
func (h *harness) child(t *testing.T) *harness {
	t.Helper()
	directory := filepath.Join(h.directory, strings.ReplaceAll(strings.TrimPrefix(t.Name(), "TestTypeAwareAgreementAndMutants/"), "/", "-"))
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	return &harness{t: t, repository: h.repository, directory: directory}
}

// Node cannot link the checker ABI. A tiny persistent ABI client supplies only
// checker calls; parsing, rule decisions and output remain in the tested source
// or emitted JavaScript. Bridge overlays therefore run unchanged on both backends.
func (h *harness) bridgeClient(name, archive string) string {
	source := h.write(name+".c", `#include "tsgo.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
static size_t natural(void) { char line[64]; if(!fgets(line,sizeof line,stdin)) exit(0); return strtoull(line,0,10); }
static tsgo_view field(void) { size_t n=natural(); char *s=malloc(n+1); if(fread(s,1,n,stdin)!=n) exit(2); s[n]=0; return (tsgo_view){s,n}; }
int main(void) {
 for(;;) {
  size_t count=natural(); tsgo_view *f=calloc(count,sizeof *f);
  for(size_t i=0;i<count;i++) f[i]=field();
  tsgo_buffer answer={0},error={0}; int status=TSGO_ARGUMENT; char handleText[64];
  if(!strcmp(f[0].data,"create") && count>=2) {
   tsgo_handle handle=0; status=tsgo_create(f[1],f+2,count-2,&handle,&error);
   if(!status) { answer.data=handleText; answer.length=snprintf(handleText,sizeof handleText,"%llu",(unsigned long long)handle); }
  } else if(!strcmp(f[0].data,"parts") && count==6) {
   status=tsgo_type_parts(strtoull(f[1].data,0,10),f[2],strtoull(f[3].data,0,10),strtoull(f[4].data,0,10),f[5],&answer,&error);
  } else if(!strcmp(f[0].data,"release") && count==2) {
   status=tsgo_release(strtoull(f[1].data,0,10),&error);
  }
  tsgo_buffer reply=status?error:answer;
  printf("%d\n%zu\n",status,reply.length); fwrite(reply.data,1,reply.length,stdout); fflush(stdout);
  if(answer.data!=handleText) tsgo_buffer_free(&answer); tsgo_buffer_free(&error);
  for(size_t i=0;i<count;i++) free((void*)f[i].data); free(f);
 }
}
`)
	binary := filepath.Join(h.directory, name)
	h.must(name+"-build", exec.Command("clang", "-O2", "-I", filepath.Join(h.repository, "bridge/tsgo"), source, archive, "-lpthread", "-ldl", "-lm", "-o", binary))
	return binary
}

func (h *harness) nodeRuntime() (string, string) {
	adapter := h.write("checker-adapter.ts", `import {panic, readTextFile, writeTextFile} from 'adamic';
function ask(fields: readonly string[]): string {
 const sent=writeTextFile('__typeaware_request__',JSON.stringify(fields) ?? panic('checker request could not be encoded'));
 if(sent.kind==='Error') { panic(sent.message); }
 const reply=readTextFile('__typeaware_response__');
 if(reply.kind==='Error') { panic(reply.message); }
 return reply.text;
}
export function tsgoProgram(config:string,files:readonly string[]):number { return Number.parseInt(ask(['create',config,...files]),10); }
export function tsgoTypeParts(handle:number,path:string,start:number,end:number,kind:string):string { return ask(['parts',String(handle),path,String(start),String(end),kind]); }
export function tsgoRelease(handle:number):void { ask(['release',String(handle)]); }
`)
	runtime := h.write("checker-runtime.mjs", `import {spawn} from 'node:child_process';
import {readSync,writeSync} from 'node:fs';
import * as ordinary from '`+filepath.Join(h.repository, "oracle/adamic.mjs")+`';
export * from '`+filepath.Join(h.repository, "oracle/adamic.mjs")+`';
const child=spawn(process.env.ADAMIC_TYPEAWARE_CLIENT,[],{stdio:['pipe','pipe','inherit']});
child.unref();
for(const stream of [child.stdin,child.stdout]) { stream._handle.setBlocking(true); stream._handle.unref(); }
let closing=false;
process.on('beforeExit',()=>{ if(!closing) { closing=true; child.stdin.end(); child.ref(); } });
process.on('exit',()=>child.kill());
const input=child.stdin._handle.fd,output=child.stdout._handle.fd;
function read(n) { const buffer=Buffer.alloc(n); let at=0; while(at<n) { const count=readSync(output,buffer,at,n-at); if(count===0) ordinary.panic('checker client closed'); at+=count; } return buffer; }
function natural() { let line=''; for(;;) { const byte=read(1)[0]; if(byte===10) return Number(line); line+=String.fromCharCode(byte); } }
let response='';
export function writeTextFile(path,text) {
 if(path!=='__typeaware_request__') return ordinary.writeTextFile(path,text);
 const fields=JSON.parse(text).map(value=>Buffer.from(value));
 const request=Buffer.concat([Buffer.from(fields.length+'\n'),...fields.flatMap(field=>[Buffer.from(field.length+'\n'),field])]);
 let at=0; while(at<request.length) at+=writeSync(input,request,at,request.length-at);
 const status=natural(); response=read(natural()).toString('utf8');
 if(status!==0) ordinary.panic(response);
 return {kind:'Ok'};
}
export function readTextFile(path) { return path==='__typeaware_response__'?{kind:'Ok',text:response}:ordinary.readTextFile(path); }
`)
	runner := h.write("checker-node.mjs", `import {readFileSync} from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
import {pathToFileURL} from 'node:url';
registerHooks({
 resolve(specifier,context,next) { return specifier==='adamic'?{url:pathToFileURL('`+runtime+`').href,shortCircuit:true}:next(specifier,context); },
 load(url,context,next) { return url.endsWith('.ts')||url.endsWith('.a')?{format:'module',source:stripTypeScriptTypes(readFileSync(new URL(url),'utf8'),{mode:'transform'}),shortCircuit:true}:next(url,context); }
});
const entry=process.argv[2];process.argv.splice(1,1);
await import(pathToFileURL('`+runtime+`').href);
await import(pathToFileURL(entry).href);
`)
	return adapter, runner
}

// Substitute only the external checker transport. The emitter receives normal
// TypeScript functions, so this does not add support to the production backend.
func (h *harness) javascriptEntry(name, entry, adapter string) (string, string) {
	h.t.Helper()
	copySource := func(path, name string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		for _, symbol := range []string{"tsgoProgram", "tsgoTypeParts", "tsgoRelease"} {
			source = strings.ReplaceAll(source, ", "+symbol, "")
			source = strings.ReplaceAll(source, symbol+", ", "")
		}
		source = "import {tsgoProgram, tsgoTypeParts, tsgoRelease} from '" + adapter + "';\n" + source
		source = strings.ReplaceAll(source, "../../typescript", filepath.Join(h.repository, "stage1/typescript"))
		source = strings.ReplaceAll(source, "../lint", filepath.Join(h.repository, "stage1/cohere/lint"))
		return h.write(name, source)
	}
	source := copySource(entry, name+"-source-node.ts")
	data, err := os.ReadFile(source)
	if err != nil {
		h.t.Fatal(err)
	}
	if strings.Contains(string(data), "./unary_minus.ts") {
		unary := copySource(filepath.Join(h.repository, "stage1/cohere/typeaware/unary_minus.ts"), name+"-unary.ts")
		data = []byte(strings.ReplaceAll(string(data), "./unary_minus.ts", unary))
	} else if strings.Contains(string(data), "./wrong-node.ts") {
		unary := copySource(filepath.Join(h.directory, "wrong-node.ts"), name+"-unary.ts")
		data = []byte(strings.ReplaceAll(string(data), "./wrong-node.ts", unary))
	}
	if err = os.WriteFile(source, data, 0644); err != nil {
		h.t.Fatal(err)
	}
	program, err := load.Load([]string{source})
	if err != nil {
		h.t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		h.t.Fatal(err)
	}
	emitted := h.write(name+"-emitted.mjs", javascript.JavaScript(lowered))
	return source, emitted
}

func (h *harness) javascriptRun(name, runner, entry, client string, args ...string) result {
	command := exec.Command("node", append([]string{"--disable-warning=ExperimentalWarning", runner, entry}, args...)...)
	command.Env = append(os.Environ(), "ADAMIC_TYPEAWARE_CLIENT="+client)
	return h.run(name, command)
}
