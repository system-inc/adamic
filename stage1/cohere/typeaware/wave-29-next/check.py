#!/usr/bin/env python3
"""Isolated rule comparison; no shared harness changes."""
import collections, hashlib, json, os, subprocess, sys, time
from pathlib import Path
repo=Path(__file__).resolve().parents[4]; source=Path(__file__).resolve().parent
d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
compiler=Path(os.environ.get('ADAMIC_COMPILER', '/workspace/wave29-regex-controls/adamic'));archive=Path('/workspace/wave29-next-checker.a')
commands=[]
def run(name,args,cwd=repo,expected=0):
 start=time.monotonic_ns()
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:
  p=subprocess.run([str(x) for x in args],cwd=cwd,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode,ns=time.monotonic_ns()-start))
 (d/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 if p.returncode!=expected: raise RuntimeError(f'{name}: exit {p.returncode}; see {d/(name+".stderr")}')
 return (d/(name+'.stdout')).read_bytes()
def frame(s):
 s=str(s);return str(len(s.encode('utf-16-le'))//2)+'\n'+s
def profiles(rows):
 s=frame(len(rows))
 for row in rows:
  options=row.get('options');entries=[];obj=False;objects=[];allow=False
  if row['rule']=='no-restricted-globals' and options:
   if isinstance(options[0],dict) and 'globals' in options[0]:
    flags=options[0];entries=flags['globals'];obj=flags.get('checkGlobalObject',False);objects=flags.get('globalObjects',[])
   else:entries=options
  if row['rule']=='no-shadow-restricted-names' and options:allow=not options.get('reportGlobalThis',True)
  s+=frame(row['rule'])+frame(len(entries))
  for entry in entries:s+=frame(entry if isinstance(entry,str) else entry['name'])+frame('' if isinstance(entry,str) else entry.get('message',''))
  s+=frame(int(obj))+frame(len(objects))+''.join(frame(x) for x in objects)+frame(int(allow))
 return s
rows=json.loads(run('extract',['go','run',source/'testdata/extract.go',repo]))
# Missing non-table controls and literal parser edge cases.
for rule,text,options in [
 ('no-restricted-globals','export const values={status}; export {status as renamed};',['status']),
 ('no-restricted-globals','export const values={status}; export {status as renamed}; const status=1;',['status']),
 ('no-restricted-globals','const {length:size}=String(); const x: Promise<void>=Promise.resolve();',['length','Promise']),
 ('no-restricted-globals','globalThis[0x10]; globalThis[("foo")]; globalThis[`foo`]; globalThis?.globalThis?.foo;',[{'globals':['16','foo'],'checkGlobalObject':True}]),
 ('no-restricted-globals','foo;',[{'name':'foo','message':'世界 🌍'}, {'name':'foo','message':'later'}]),
 ('no-restricted-globals','const a={event}; event; window.event;',[]),
 ('no-setter-return','Object.defineProperty({},"x",{set: value => (value+1)}); Object.defineProperties({}, {x:{["set"](value){return value;}}}); Object.create({}, {x:{[`set`]: value=>value}}); Reflect["defineProperty"]({},"x",{set: value=>value});',None),
 ('no-setter-return','let Object:any;Object.defineProperty({},"x",{set: value=>value});',None),
 ('no-setter-return','Object.defineProperty({set(v){return v;}},"x",{}); Object.create({}, {set(v){return v;}}); ({set x(v){return (()=>{return v;})();}});',None),
 ('no-shadow-restricted-names','var undefined; ({undefined}=obj);',None),
 ('no-shadow-restricted-names','function f(){var undefined;} var undefined; undefined=1;',None),
 ('no-shadow-restricted-names','var undefined; function f(){var undefined; undefined=1;}',None),
 ('no-shadow-restricted-names','var undefined = /* comment */ 1; var {x:NaN}=obj; enum E{undefined};',None),
 ('no-shadow-restricted-names','var globalThis; const eval=1;',{ 'reportGlobalThis':False}),
 ('no-shadow-restricted-names','/* 世界 🌍 */\r\nfunction NaN(undefined){return undefined;}\r\n',None),
]:rows.append(dict(rule=rule,source=text,options=options,environment='',function='local'))
(d/'fixtures.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
virtual=repo/'cohere/adamic_wave29_next_oracle.go'
(d/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle.go')}}))
oracle=d/'oracle';run('oracle-build',['go','build','-overlay',d/'overlay.json','-o',oracle,virtual],cwd=repo/'cohere')
run('native-build',[compiler,'build',source/'runner.a','-o',d/'native','--tsgo',archive])
groups=collections.defaultdict(list)
for row in rows:groups[row.get('environment','')].append(row)
results=[]; datasets=[]
for at,(env,fixtures) in enumerate(groups.items()):
 group=d/f'group-{at:02d}';group.mkdir(exist_ok=True);paths=[]
 for i,row in enumerate(fixtures):
  path=group/f'input-{i:03d}.a';path.write_text(row['source']+'\nexport {};\n');paths.append(str(path))
 config=group/'config.json';ambient=group/'environment.d.ts';ambient.write_text(''.join('declare var '+name+':any;\n' for name in env.split(',') if name and name!='globalThis'))
 config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'module':'ESNext'},'files':paths+[str(ambient)]}))
 manifest=group/'manifest';manifest.write_text('\n'.join(paths)+'\n')
 (group/'profiles.frames').write_text(profiles(fixtures));(group/'profiles.json').write_text(json.dumps(fixtures))
 args=[config,manifest];datasets.append((f'group-{at:02d}',args,group/'profiles.frames',group/'profiles.json'))
 truth=run(f'group-{at:02d}-go',[oracle,*args,group/'profiles.json'])
 got=run(f'group-{at:02d}-native',[d/'native',*args,group/'profiles.frames'])
 if truth!=got:
  a=truth.splitlines();b=got.splitlines()
  for i,(x,y) in enumerate(zip(a,b)):
   if x!=y:print('mismatch group',at,'line',i,'Go',a[max(i-1,0):i+3],'native',b[max(i-1,0):i+3]);break
  raise RuntimeError('configured bytes differ')
 results.append(truth);print(f'group {at}: {len(fixtures)} cases, {len(truth)} bytes, '+truth.splitlines()[-1].decode(),flush=True)
if '--controls-only' in sys.argv:sys.exit(0)
# The exact frozen compiler and repository population used by the original wave.
for name,config in [('compiler',Path('/workspace/wave29-typescript/src/compiler/tsconfig.json')),('repository',repo/'tsconfig.json')]:
 manifest=d/(name+'.manifest');manifest.write_bytes((Path('/workspace/wave29-regex-default')/(name+'.manifest')).read_bytes())
 args=[config,manifest];datasets.append((name,args,None,None))
 truth=run(name+'-go',[oracle,*args]);got=run(name+'-native',[d/'native',*args]);assert truth==got,name+' bytes differ'
 results.append(truth);print(name, len(manifest.read_text().splitlines()),len(truth),truth.splitlines()[-1].decode(),flush=True)
# Each fault changes a rule verdict but still compiles, exits 0 and writes no stderr.
for label,module,before,after in [
 ('globals','no_restricted_globals.a','!this.tree.symbol(index).source()','this.tree.symbol(index).source()'),
 ('setter','no_setter_return.a',"this.tree.kind(target) === 'SetAccessor' || this.descriptor(target)","this.tree.kind(target) === 'GetAccessor' || this.descriptor(target)"),
 ('shadow','no_shadow_restricted_names.a',"if(text === 'undefined' && this.safe(name)) { continue; }","if(text === 'undefined') { continue; }"),
]:
 local=d/(label+'-source');local.mkdir(exist_ok=True)
 for path in source.glob('*.a'):
  text=path.read_text()
  if path.name==module:assert text.count(before)==1;text=text.replace(before,after)
  text=text.replace("'../", "'"+str(source.parent)+'/').replace("'../../../typescript/", "'"+str(repo/'stage1/typescript')+'/')
  # runner's ../../../typescript replacement must precede its ../ prefix expansion.
  text=text.replace(str(source.parent)+'/../../typescript/',str(repo/'stage1/typescript')+'/')
  (local/path.name).write_text(text)
 binary=d/(label+'-mutant');run(label+'-mutant-build',[compiler,'build',local/'runner.a','-o',binary,'--tsgo',archive])
 killed=False
 for name,args,nativeProfile,goProfile in datasets:
  if nativeProfile is None:continue
  got=run(label+'-'+name,[binary,*args,nativeProfile]);truth=(d/(name+'-go.stdout')).read_bytes()
  assert not(d/(label+'-'+name+'.stderr')).read_bytes()
  if got!=truth:killed=True;print(label,'mutant: compiled, exit 0, empty stderr, byte comparison killed on',name,flush=True);break
 assert killed,label+' mutant survived'
oldcc=os.environ.get('CC');oldflags=os.environ.get('CGO_CFLAGS')
os.environ['CC']='clang';os.environ['CGO_CFLAGS']='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'
run('asan-archive',['go','build','-buildmode=c-archive','-o',d/'asan-checker.a','./bridge/tsgo/archive'])
for key,value in [('CC',oldcc),('CGO_CFLAGS',oldflags)]:
 if value is None:os.environ.pop(key,None)
 else:os.environ[key]=value
run('asan-build',[compiler,'build',source/'runner.a','-o',d/'asan-native','--tsgo',d/'asan-checker.a','--sanitize'])
os.environ['ASAN_OPTIONS']='detect_leaks=1:halt_on_error=1';os.environ['UBSAN_OPTIONS']='halt_on_error=1'
for name,args,nativeProfile,goProfile in datasets:
 got=run(name+'-asan',[d/'asan-native',*args,*([nativeProfile] if nativeProfile else [])]);truth=(d/(name+'-go.stdout')).read_bytes()
 assert got==truth and not(d/(name+'-asan.stderr')).read_bytes(),name+' sanitizer mismatch'
print('all configured and corpus streams match under ASan/UBSan/LeakSanitizer',flush=True)
for name,args,nativeProfile,goProfile in datasets:
 if nativeProfile:continue
 run(name+'-timed-go',[oracle,*args,'--count']);run(name+'-timed-native',[d/'native',*args,'--count'])
 native=next(x['ns'] for x in commands if x['name']==name+'-timed-native');go=next(x['ns'] for x in commands if x['name']==name+'-timed-go')
 print(name,'native ns',native,'Go ns',go,'ratio',native/go,flush=True)
print('combined stream sha256',hashlib.sha256(b''.join(results)).hexdigest(),flush=True)
