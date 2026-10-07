#!/usr/bin/env python3
"""Real production callbacks versus the owned raw-checker-fact decision kernel."""
import hashlib,json,os,re,subprocess,sys,time
from pathlib import Path
s=Path(__file__).resolve().parent;r=s.parents[6];d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
c=Path(os.environ.get('ADAMIC_COMPILER','/workspace/wave29-b8fb-adamic'));records=[]
def run(name,args,cwd=r,env=None):
 start=time.monotonic_ns()
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:p=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=out,stderr=err)
 records.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode,ns=time.monotonic_ns()-start));(d/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
 assert p.returncode==0,(name,p.returncode)
 return (d/(name+'.stdout')).read_bytes()
v=r/'cohere/adamic_wave29_undef_resolution.go';(d/'overlay.json').write_text(json.dumps({'Replace':{str(v):str(s/'resolution_oracle.go')}}))
run('go-build',['go','build','-overlay',d/'overlay.json','-o',d/'oracle',v],cwd=r/'cohere')
(d/'dep.ts').write_text('export const Component=null; export namespace Foo { export const App=null; }\n')
(d/'ambient.d.ts').write_text('declare const External:any;\n')
cases=[
 ('missing','const value=<App/>;',True),('local','const App=null;const value=<App/>;',True),
 ('function','function App(){} const value=<App/>;',True),('class','class App{} const value=<App/>;',True),
 ('enum','enum App {One} const value=<App/>;',False),('enum-member','enum Other {App} const value=<App/>;',False),
 ('import','import {Component as App} from "./dep";const value=<App/>;',True),
 ('import-equals','import App=require("./dep");const value=<App/>;',False),
 ('interface','interface App{} const value=<App/>;',False),('type-alias','type App=string;const value=<App/>;',False),
 ('scope-in','{const App=null;const value=<App/>;}',True),('scope-out','{const App=null;}const value=<App/>;',True),
 ('parameter','const f=(App:any)=><App/>;',False),('lib-global','const value=<Map/>;',True),
 ('ambient-global','const value=<External/>;',True),
 ('own-global','declare global {var G:any;} const value=<G/>;',False),
 ('intrinsic','const value=<div/>;',True),('custom','const value=<Foo-bar/>;',True),
 ('underscore','const value=<_foo/>;',True),('dollar','const value=<$foo/>;',True),
 ('unicode','/*😀*/const value=<테스트/>;',True),('declared-unicode','const 테스트=null;/*😀*/const value=<테스트/>;',True),
]
paths=[];profile=[]
for label,code,js in cases:
 for extension in (['tsx','cjs'] if js else ['tsx']):
  for available in [False,True]:
   for allow in [False,True]:
    path=d/(label+'-'+extension+'-'+str(int(available))+str(int(allow))+'.'+extension)
    path.write_text(code+'export {};\n');paths.append(str(path))
    profile.append(dict(rule='react/jsx-no-undef',options={'allowGlobals':allow},checkerAvailable=available))
(d/'config.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'jsx':'preserve','module':'ESNext','allowJs':True},'files':paths+[str(d/'ambient.d.ts')]}))
(d/'manifest').write_text('\n'.join(paths)+'\n');(d/'profile.json').write_text(json.dumps(profile));frames=d/'facts.frames'
truth=run('go',[d/'oracle',d/'config.json',d/'manifest',d/'profile.json'],env=dict(os.environ,ADAMIC_RESOLUTION_FACTS=str(frames)))
assert b'jsxIdentifierNotDefined' in truth and b'findings 0' not in truth
probe=s/'resolution_probe.a'
run('native-build',[c,'build',probe,'-o',d/'native']);assert run('native',[d/'native',frames])==truth and not(d/'native.stderr').read_bytes()
node=['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs'];assert run('node',node+[probe,frames])==truth and not(d/'node.stderr').read_bytes()
(d/'probe.mjs').write_bytes(run('javascript-build',[c,'js',probe]));assert run('javascript',node+[d/'probe.mjs',frames])==truth and not(d/'javascript.stderr').read_bytes()
run('asan-build',[c,'build',probe,'-o',d/'asan','--sanitize']);assert run('asan',[d/'asan',frames],env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1'))==truth and not(d/'asan.stderr').read_bytes()
mutants=[('globals','allowGlobals ||','false ||'),('commonjs',"endsWith('.cjs')","endsWith('.mjs')"),('declaration','|| declaredInFile','|| false'),('checker','if(!checkerAvailable)','if(false)'),('resolved','if(resolved &&','if(false &&'),('message','missing import.','missing export.'),('positions','code < 128 ? 1','code < 128 ? 2')]
for name,before,after in mutants:
 module=s.parent/('positions.a' if name=='positions' else 'resolution.a')
 text=module.read_text();assert text.count(before)==1
 local=d/('mutant-'+name);local.mkdir(exist_ok=True);mutated=local/module.name;mutated.write_text(text.replace(before,after))
 def replace(m):
  target=(probe.parent/m[2]).resolve()
  if target==module.resolve():target=mutated
  if target==(s.parent/'rule.a').resolve():target=local/'rule.a'
  return 'from '+repr(str(target))
 ruletext=(s.parent/'rule.a').read_text()
 def rule_import(m):
  target=(s.parent/m[2]).resolve()
  if target==module.resolve():target=mutated
  return 'from '+repr(str(target))
 (local/'rule.a').write_text(re.sub(r"from\s+(['\"])(\.[^'\"]+)\1",rule_import,ruletext))
 localprobe=local/'probe.a';localprobe.write_text(re.sub(r"from\s+(['\"])(\.[^'\"]+)\1",replace,probe.read_text()))
 run(name+'-build',[c,'build',localprobe,'-o',local/'native']);got=run(name,[local/'native',frames]);assert got!=truth and not(d/(name+'.stderr')).read_bytes()
 print(name+': compiled, exit 0, empty stderr; only production Go bytes catch it',flush=True)
print('PASS: '+str(len(paths))+' real source/option/checker-availability controls; '+truth.splitlines()[-1].decode()+'; '+str(len(truth))+' identical bytes; sha256='+hashlib.sha256(truth).hexdigest(),flush=True)
print('Go/native/Node/emitted JavaScript/ASan/UBSan/leaks agree on findings, ranges, descriptions and zero fixes/suggestions; raw facts adapter remains test-only',flush=True)
