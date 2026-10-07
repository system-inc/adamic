"""Owned continuation validation; does not edit the shared harness or dispatcher."""
import argparse, gzip, hashlib, json, os, pathlib, re, shutil, subprocess, time

REPO=pathlib.Path(__file__).resolve().parents[4]
OWN=pathlib.Path(__file__).resolve().parent
NAMES=['nexus/correctness-require-child-process-error-listener','nexus/correctness-require-response-status-check','nexus/performance-no-independent-await-in-loop']

def main():
 p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);p.add_argument('--compiler');p.add_argument('--initial',action='store_true');args=p.parse_args()
 scratch=pathlib.Path(args.scratch).resolve();scratch.mkdir(parents=True,exist_ok=True);records=[]
 def run(name,command,env=None,expect=0):
  start=time.monotonic()
  with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:
   result=subprocess.run(command,cwd=REPO,stdout=out,stderr=err,env=env)
  record={'name':name,'command':[str(x) for x in command],'exit':result.returncode,'seconds':time.monotonic()-start};records.append(record)
  (scratch/'runs.json').write_text(json.dumps(records,indent=2)+'\n')
  if result.returncode!=expect:raise RuntimeError(str(record)+'\n'+(scratch/(name+'.stderr')).read_text())
  return (scratch/(name+'.stdout')).read_bytes()
 def overlay(name,replacements):
  path=scratch/(name+'.json');path.write_text(json.dumps({'Replace':replacements}));return path
 archive=scratch/'checker.a';asan_archive=scratch/'checker-asan.a';stage0=scratch/'adamic';native=scratch/'native';asan=scratch/'native-asan';oracle=scratch/'oracle'
 run('stage0-build',['go','build','-o',stage0,'./cmd/adamic'])
 run('archive-build',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
 run('native-build',[stage0,'build',OWN/'suite.a','-o',native,'--tsgo',archive])
 virtual=REPO/'cohere/adamic_wave06nextoracle.go'
 go_overlay=overlay('oracle-overlay',{str(virtual):str(OWN/'testdata/oracle.go')})
 run('oracle-build',['go','-C',REPO/'cohere','build','-overlay',go_overlay,'-o',oracle,virtual])
 globals_file=scratch/'globals.d.ts';globals_file.write_text("""declare module 'node:child_process' {
 interface ChildProcess {on(event:string,listener:(...args:unknown[])=>void):ChildProcess;once(event:string,listener:(...args:unknown[])=>void):ChildProcess;addListener(event:string,listener:(...args:unknown[])=>void):ChildProcess;prependListener(event:string,listener:(...args:unknown[])=>void):ChildProcess;prependOnceListener(event:string,listener:(...args:unknown[])=>void):ChildProcess;off(event:string,listener:(...args:unknown[])=>void):ChildProcess;removeListener(event:string,listener:(...args:unknown[])=>void):ChildProcess;removeAllListeners(event?:string):ChildProcess;setMaxListeners(n:number):ChildProcess;unref():void;kill():void;pid:number;stdout:unknown;}
 export function spawn(command:string,args?:string[]):ChildProcess;
 export function fork(file:string):ChildProcess;
 export function exec(command:string):ChildProcess;
}
""")
 config=scratch/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','moduleResolution':'bundler','lib':['ES2022','DOM']},'files':[str(globals_file)]}))
 child=[
 "spawn('x');", "spawn('x').on('close',()=>{});", "spawn('x').on('error',()=>{});", "spawn('x').once('error',()=>{}).unref();", "spawn('x').off('error',()=>{}).unref();", "const c=spawn('x');c.kill();", "const c=spawn('x');c.on('error',()=>{});", "const c=spawn('x');const copy=c;", "function f(){return spawn('x');}", "const c=spawn('x');use(c);", "const c=spawn('x');const x={c};", "const c=spawn('x');export {c};", "spawn('x').on(event,()=>{});", "spawn('x').on('close' as 'close'|'exit',()=>{});", "fork('x');", "exec('x');", "const c=spawn('x');type T=typeof c;", "const c=spawn('x');if(c)c.kill();", "let c=spawn('x');c=spawn('y');", "const c=spawn('x');function f(c:{on:(s:string,f:()=>void)=>void}){c.on('error',()=>{});}",
 ]
 response=[
 "await fetch(url);", "const r=await fetch(url);", "const r=await fetch(url);return r.json();", "return (await fetch(url)).json();", "const r=await fetch(url);if(!r.ok)throw 1;return r.json();", "const r=await fetch(url);const data=await r.json();if(!r.ok)throw data;return data;", "const r=await fetch(url);const text=await r.text();throw text;", "const r=await fetch(url);use(r);return r.json();", "const r=await fetch(url);if(flag)use(r.status);return r.json();", "const r=await fetch(url);return r.body;", "const r=await fetch(url);return r.headers;", "let r=await fetch(url);return r.json();", "const r=await fetch(url);function later(){return r.ok;}return r.json();", "const r=await fetch(url);const other={r};return r.json();", "return fetch(url).then(r=>r.json());", "return fetch(url).then(r=>{if(!r.ok)throw 1;return r.json();});", "const r=await fetch(url);try{return r.json();}finally{use(r.ok);}", "const r=await fetch(url);if(flag)return r.json();else throw 1;", "const r=await fetch(url);return r['json']();", "const r=await fetch(url);return r.statusText;",
 ]
 loops=[
 "for(const item of items){await work(item);}", "for(const item of items){await item;}", "for(const item of items){results.push(await work(item));}", "for(const item of items){results.push(await work(item));use(results.length);}", "for(const item of items){counter+=await work(item);}", "for(const item of items){await work(item);console.log(item);}", "for(const item of items){await sleep(1);await work(item);}", "for(const item of items){if(await work(item))return;}", "for(const item of items){const x=await work(item);const y=x;if(y)break;}", "for(const item of items){if(!item)throw 1;await work(item);}", "for(const item of items){try{results.push(await work(item));}catch(e){console.error(e);}}", "for(let i=0;i<items.length;i++){await work(items[i]);}", "for(let i=0;i<items.length;i++){i++;await work(items[i]);}", "while(flag){await work(1);}", "for await(const item of items){await work(item);}", "for(const item of items){for(const x of items){await work(x);}}", "for(const item of items){await printItem(item);}", "for(const item of items){seen.add(item);await work(item);if(seen.has(item))use(item);}", "for(const item of items){this.cache.push(await this.fetch(item));}", "for(const item of items){await work(item);continue;}",
 ]
 child += ["import * as children from 'node:child_process';children.spawn('x');", "import {spawn as launch} from 'node:child_process';launch('x');", "function f(){function spawn(x:string){return x;}spawn('x');}", "const c=spawn('x');if(!c)throw 1;c.kill();", "spawn('x').on('close' as 'close'|'error',()=>{});", "const c=spawn('x');c.removeAllListeners().once('error',()=>{});", "spawn('x').on;", "spawn('x')['on']('error',()=>{});"]
 response += ["function fetch(x:string){return Promise.resolve({json(){return 1;}})}const r=await fetch(url);return r.json();", "const r=await fetch(url);if(flag){return r.json();}use(r.ok);return 1;", "const r=await fetch(url);try{return r.json();}catch{return 1;}", "const r=await fetch(url);try{throw 1;}catch{use(r.ok);}return r.json();", "return fetch(url).then(r=>{const t=r.text();throw t;});", "const r=await fetch(url);return (r as Response).json();", "for(const u of [url]){const r=await fetch(u);use(await r.json());}", "return (await fetch(url)).body;"]
 loops += ["for(const item of items){var x=await work(item);use(x);}", "for(const item of items){const x=await work(item);if(x)continue;results.push(x);}", "for(const item of items){if(await work(item)){throw 1;}}", "for(const item of items){this.cache.push(await this.fetch(item));use(this.cache.length);}", "for(const item of items){const local:number[]=[];local.pop();await work(item);}", "for(let i=0;i<items.length;++i){await work(items[i]);}", "for(let i=0;i<items.length;i+=1){await work(items[i]);}", "for(const item of items){try{await work(item);}catch(e){results.push(0);}}"]
 controls=[]
 for group,examples in [('child',child),('response',response),('loop',loops)]:
  for i,example in enumerate(examples):
   file=scratch/(group+'-'+str(i)+'.a')
   if group=='child':source="import {spawn,fork,exec} from 'node:child_process';declare const event:string;declare function use(x:unknown):void;"+example+'\nexport {};'
   elif group=='response':source='declare const url:string;declare const flag:boolean;declare function use(x:unknown):void;async function f(){'+example+'}\nexport {};'
   else:source='declare const items:number[];declare const results:number[];declare let counter:number;declare const flag:boolean;declare const seen:Set<number>;declare function work(x:number):Promise<number>;declare function sleep(x:number):Promise<void>;declare function use(x:unknown):void;function printItem(x:number){console.log(x);return work(x);}class Example{cache:number[]=[];fetch(x:number){return work(x);}async f(){'+example+'}}\nexport {};'
   file.write_text(source);controls.append(file)
 manifest=scratch/'controls.manifest';manifest.write_text(''.join(str(x)+'\n' for x in controls))
 def compare(name,binary,configuration,inputs,timed=False):
  env=os.environ.copy()
  if timed:env['ADAMIC_TSGO_TIMING']='1'
  native_output=run(name+'-native',[binary,configuration,inputs],env)
  go_output=run(name+'-go',[oracle,configuration,inputs])
  if native_output!=go_output:
   a=native_output.decode().splitlines();b=go_output.decode().splitlines()
   import difflib
   difference='\n'.join(difflib.unified_diff(b,a,fromfile='Go',tofile='native'))
   (scratch/(name+'-diff.txt')).write_text(difference)
   raise AssertionError(difference)
  print(name, len(native_output),'identical bytes',native_output.decode().splitlines()[-1],flush=True)
  return go_output
 truth=compare('controls',native,config,manifest)
 for name in NAMES:assert name.encode() in truth,name+' missing positive control'
 if args.initial:return
 env=os.environ.copy();env.update(CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
 run('archive-asan-build',['go','build','-buildmode=c-archive','-o',asan_archive,'./bridge/tsgo/archive'],env)
 run('native-asan-build',[stage0,'build',OWN/'suite.a','-o',asan,'--tsgo',asan_archive,'--sanitize'])
 compare('controls-asan',asan,config,manifest)
 assert (scratch/'controls-asan-native.stderr').read_bytes()==b''
 for folder,old in [('child_process_error_listener','childProcessWithoutErrorListener'),('response_status_check','bodyReadWithoutStatusCheck'),('independent_await_in_loop','independentAwaitInLoop')]:
  directory=scratch/(folder+'-source');shutil.copytree(OWN,directory,dirs_exist_ok=True)
  for file in directory.rglob('*.a'):
   original=OWN/file.relative_to(directory)
   text=file.read_text();text=re.sub(r"(['\"])([^'\"]+\.ts)\1",lambda m:repr(str((original.parent/m.group(2)).resolve())),text)
   if file==directory/folder/'index.a':assert text.count("'"+old+"'")==1;text=text.replace("'"+old+"'", "'"+old+"Mutant'")
   file.write_text(text)
  binary=scratch/(folder+'-mutant');run(folder+'-mutant-build',[stage0,'build',directory/'suite.a','-o',binary,'--tsgo',archive])
  output=run(folder+'-mutant-run',[binary,config,manifest]);assert output!=truth and (scratch/(folder+'-mutant-run.stderr')).read_bytes()==b''
  print(folder,'mutant exits 0, Go byte comparison catches it',flush=True)
 # The new question follows the archive's existing released-handle boundary.
 probe=scratch/'released.a';probe.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const a=programArguments();const p=tsgoProgram(a[0]??'',[a[1]??'']);tsgoRelease(p);console.log(tsgoInspect(p,a[1]??'',0,1,'CallExpression','wave06-declarations\\ncall'));")
 released=scratch/'released';run('released-build',[stage0,'build',probe,'-o',released,'--tsgo',archive]);run('released-run',[released,config,controls[0]],expect=70)
 assert (scratch/'released-run.stderr').read_text()=='adamic: panic: invalid or released checker handle\n'
 # Mutation of the Go fact projection, never of the oracle.
 helper=REPO/'bridge/tsgo/checker/wave06_declarations.go';source=helper.read_text();needle='declarations = append(declarations, signature.Declaration())';assert source.count(needle)==1
 mutated=scratch/'question.go';mutated.write_text(source.replace(needle,'declarations = append(declarations, node)'))
 fact_overlay=overlay('question-mutant-overlay',{str(helper):str(mutated)})
 fact_archive=scratch/'question-mutant.a';run('question-mutant-archive',['go','build','-overlay',fact_overlay,'-buildmode=c-archive','-o',fact_archive,'./bridge/tsgo/archive'])
 fact_binary=scratch/'question-mutant';run('question-mutant-build',[stage0,'build',OWN/'suite.a','-o',fact_binary,'--tsgo',fact_archive]);output=run('question-mutant-run',[fact_binary,config,manifest]);assert output!=truth and (scratch/'question-mutant-run.stderr').read_bytes()==b''
 print('fact mutant exits 0, Go finding bytes catch the wrong resolved declaration; released handle panic 70 verified',flush=True)
 if args.compiler:
  compiler=pathlib.Path(args.compiler).resolve();pin=subprocess.check_output(['git','-C',compiler,'rev-parse','HEAD'],text=True).strip();assert pin=='050880ce59e30b356b686bd3144efe24f875ebc8'
  for name,base,configuration in [('compiler',compiler,compiler/'src/compiler/tsconfig.json'),('repository',REPO,REPO/'tsconfig.json')]:
   relative=(OWN.parent/'validation-coverage'/(name+'.manifest')).read_text().splitlines();inputs=scratch/(name+'.manifest');inputs.write_text(''.join(str(base/x)+'\n' for x in relative if x))
   compare(name,native,configuration,inputs);compare(name+'-asan',asan,configuration,inputs);assert (scratch/(name+'-asan-native.stderr')).read_bytes()==b''
   compare(name+'-timed',native,configuration,inputs,True)
 print('PASS',flush=True)

if __name__=='__main__':main()
