#!/usr/bin/env python3
"""Private native/Go byte oracle, preserving all subprocess output in files."""
from pathlib import Path
import argparse, gzip, hashlib, json, os, re, shutil, subprocess, time

parser=argparse.ArgumentParser()
parser.add_argument('directory',type=Path)
parser.add_argument('--adamic',default='/workspace/typeaware-wave-04/final/adamic')
parser.add_argument('--compiler-manifest')
parser.add_argument('--compiler-config')
parser.add_argument('--repository-manifest')
parser.add_argument('--blocking-controls',type=Path)
parser.add_argument('--process-controls',type=Path)
args=parser.parse_args()
r=Path(__file__).resolve().parents[4];own=Path(__file__).resolve().parent;out=args.directory.resolve();out.mkdir(parents=True,exist_ok=True)
sequence=0

def run(label,command,cwd=r,environment=None,expected=0):
 global sequence
 sequence+=1;stem=out/f'{sequence:03d}-{label}';started=time.perf_counter()
 with stem.with_suffix('.stdout').open('wb') as stdout,stem.with_suffix('.stderr').open('wb') as stderr:
  result=subprocess.run([str(x) for x in command],cwd=cwd,env=environment,stdout=stdout,stderr=stderr)
 elapsed=time.perf_counter()-started
 a=stem.with_suffix('.stdout').read_bytes();b=stem.with_suffix('.stderr').read_bytes()
 if result.returncode!=expected:raise RuntimeError(f'{label}: exit {result.returncode}; see {stem}.stderr')
 print(f'{label}: exit {result.returncode}, {len(a)} stdout bytes, {len(b)} stderr bytes, {elapsed:.6f}s',flush=True)
 return a,b,elapsed

adapter=own/'_adapter';adapter.mkdir(exist_ok=True)
for name in ['main.go','profile.go','boundary.c']:
 data=(r/'bridge/tsgo/archive'/name).read_text()
 if name=='main.go':
  if data.count('program.Inspect(text(file)')!=1:raise RuntimeError('shared archive shape changed')
  data=data.replace('program.Inspect(text(file)','program.InspectWave04Next(text(file)')
 (adapter/name).write_text(data)
(adapter/'tsgo.h').write_bytes((r/'bridge/tsgo/tsgo.h').read_bytes())
virtual=r/'cohere/adamic_wave04_next_oracle.go';overlay=out/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'testdata/oracle.go')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',out/'oracle',virtual],cwd=r/'cohere')
for sanitized in [False,True]:
 name='asan' if sanitized else 'normal';archive=out/f'checker-{name}.a';binary=out/f'native-{name}'
 environment=dict(os.environ)
 if sanitized:environment.update(CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
 run(f'{name}-archive',['go','build','-buildmode=c-archive','-o',archive,adapter],environment=environment)
 command=[args.adamic,'build',own/'suite.a','-o',binary,'--tsgo',archive]
 if sanitized:command.append('--sanitize')
 run(f'{name}-native-build',command)

fixture=(r/'cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout_test.go').read_text()
blocks=re.findall(r'\[\]string\{\s*((?:"(?:[^"\\]|\\.)*"\s*,?\s*)+)\}',fixture)
arrays=[[json.loads(t) for t in re.findall(r'"(?:[^"\\]|\\.)*"',b)] for b in blocks]
controls=out/'controls';controls.mkdir(exist_ok=True)
node=next(a for a in arrays if any('namespace NodeJS' in line for line in a));(controls/'timers.d.ts').write_text('\n'.join(node)+'\n')
prefix='\n'.join(next(a for a in arrays if any('declare function generateImage' in line for line in a)))+'\n'
sources=['\n'.join(a)+'\n' for a in arrays if any('Promise.race' in line or 'Promise.all' in line for line in a)]
sources += [
 "const timeout=new Promise((_r,reject)=>((setTimeout(reject,milliseconds))));Promise.race([work(),timeout]);",
 "let timeout=new Promise((_r,reject)=>setTimeout(reject,milliseconds));Promise.race([work(),timeout]);",
 "Promise.race([work(),new Promise((_r,reject)=>{const timer=setTimeout(reject,milliseconds);use({timer});})]);",
 "Promise.race([work(),new Promise((_r,reject)=>{let timer;timer=setTimeout(reject,milliseconds);})]);",
 "Promise.race([work(),new Promise((_r,reject)=>{let timer;let other;other=timer=setTimeout(reject,milliseconds);})]);",
 "Promise.race([work(),new Promise((_r,reject)=>{const timer=setTimeout(reject,milliseconds);function shadow(timer:number){return timer;}use(shadow);})]);",
 "Promise.race([work(),new Promise((_r,reject)=>{class Timer{start(){setTimeout(reject,milliseconds);}}})]);",
 "/* 世界 🌍 */\r\nPromise.race([work(),new Promise((_r,reject)=>{ void (setTimeout(reject,milliseconds)); })]);\r\n",
 '/// <reference lib="dom" />\nPromise.race([work(),new Promise((_r,reject)=>{window.setTimeout(reject,milliseconds);})]);',
]
paths=[]
for i,source in enumerate(sources):
 path=controls/f'control-{i:03d}.a';path.write_text(prefix+source);paths.append(path)
config=controls/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022']},'files':['timers.d.ts']}))
manifest=controls/'manifest';manifest.write_text('\n'.join(str(p) for p in paths)+'\n')
truth=None;records=[];truths={};positive=set()
corpora=[('controls',config,manifest)]
if args.process_controls:corpora.append(('process-controls',args.process_controls/'tsconfig.json',args.process_controls/'manifest'))
if args.blocking_controls:
 for case in json.loads((args.blocking_controls/'corpora.json').read_text()):corpora.append((case['name'],Path(case['config']),Path(case['manifest'])))
if args.repository_manifest:corpora.append(('repository',r/'tsconfig.json',Path(args.repository_manifest)))
if args.compiler_manifest:corpora.append(('compiler',Path(args.compiler_config),Path(args.compiler_manifest)))
for name,config_path,manifest_path in corpora:
 expected,stderr,elapsed=run(name+'-go',[out/'oracle',config_path,manifest_path])
 truths[name]=(config_path,manifest_path,expected)
 for rule in ['no-uncleared-race-timeout','no-process-exit-after-output','require-blocking-standard-streams']:
  if ('nexus/correctness-'+rule+'\t').encode() in expected:positive.add(rule)
 if name=='controls':
  truth=expected
  if not re.search(rb'findings [1-9]\d*\n$',truth):raise RuntimeError('positive timeout controls absent')
 for variant in ['normal','asan']:
  actual,error,native_elapsed=run(name+'-'+variant,[out/f'native-{variant}',config_path,manifest_path])
  if error or actual!=expected:raise RuntimeError(name+' '+variant+' differs from Go or has sanitizer stderr')
  records.append(dict(corpus=name,variant=variant,bytes=len(actual),sha256=hashlib.sha256(actual).hexdigest(),go_seconds=elapsed,native_seconds=native_elapsed))
  print(name+' '+variant+': complete finding bytes match',flush=True)
(out/'comparisons.json').write_text(json.dumps(records,indent=2)+'\n')
if positive!=set(['no-uncleared-race-timeout','no-process-exit-after-output','require-blocking-standard-streams']):raise RuntimeError('positive rule controls absent: '+str(positive))
# Each span mutant compiles and exits normally, and only byte comparison kills it.
mutants=[]
for filename,rule in [('no_uncleared_race_timeout.a','no-uncleared-race-timeout'),('no_process_exit_after_output.a','no-process-exit-after-output'),('require_blocking_standard_streams.a','require-blocking-standard-streams')]:
 original=(own/filename).read_text();before='this.rules.byte(node.end),' if filename.startswith('no_uncleared') else 'f.rules.byte(node.end),';after=before.replace('),',') + 1,')
 if original.count(before)!=1:raise RuntimeError('nonunique mutant '+filename)
 mutant_dir=out/('mutant-'+rule);mutant_dir.mkdir(exist_ok=True)
 for p in own.glob('*.a'):
  source=p.read_text().replace("from '../", "from '"+str(own.parent)+"/").replace("from '../../../typescript/", "from '"+str(r/'stage1/typescript')+'/')
  if p.name==filename:source=source.replace(before,after)
  (mutant_dir/p.name).write_text(source)
 binary=out/('mutant-native-'+rule)
 run('mutant-build-'+rule,[args.adamic,'build',mutant_dir/'suite.a','-o',binary,'--tsgo',out/'checker-normal.a'])
 config_path,manifest_path,expected=next(v for v in truths.values() if ('nexus/correctness-'+rule+'\t').encode() in v[2])
 actual,error,_=run('mutant-run-'+rule,[binary,config_path,manifest_path])
 if error or actual==expected:raise RuntimeError('mutant survived or failed outside comparison: '+rule)
 first=next(i for i,(a,b) in enumerate(zip(actual,expected)) if a!=b)
 print(f'{rule} mutant: compiled, exit 0, empty stderr; byte oracle catches byte {first}',flush=True)
 mutants.append(dict(rule=rule,first_differing_byte=first,exit=0,stderr_bytes=0))
(out/'mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
run('released-build',[args.adamic,'build',own/'released.a','-o',out/'released','--tsgo',out/'checker-normal.a'])
probe=out/'probe.a';probe.write_text('x;')
a,b,_=run('released-run',[out/'released',controls/'tsconfig.json',probe],expected=70)
if a or b!=b'adamic: panic: invalid or released checker handle\n':raise RuntimeError('released handle escaped')
# Retaining a released handle must make the expected panic check fail.
main=adapter/'main.go';original=main.read_text();before='delete(programs.live, uint64(handle))'
if original.count(before)!=1:raise RuntimeError('nonunique registry mutant')
try:
 main.write_text(original.replace(before,'// mutant: retain the released handle'))
 run('registry-mutant-archive',['go','build','-buildmode=c-archive','-o',out/'registry-mutant.a',adapter])
finally:main.write_text(original)
run('registry-mutant-build',[args.adamic,'build',own/'released.a','-o',out/'registry-mutant','--tsgo',out/'registry-mutant.a'])
a,b,_=run('registry-mutant-run',[out/'registry-mutant',controls/'tsconfig.json',probe])
if b or not a:raise RuntimeError('registry mutant failed outside the released-handle check')
(out/'released-mutant.json').write_text(json.dumps(dict(expected_exit=70,mutant_exit=0,stdout_bytes=len(a),stderr_bytes=0,caught_by='released-handle panic assertion'))+'\n')
print('released-handle registry mutant: compiled, exit 0, empty stderr; expected panic assertion catches it',flush=True)
# Preserve corpus inputs and hashes, including roots that produce no finding.
inputs=[]
for name,_,m in corpora:
 for path in m.read_text().splitlines():inputs.append(dict(corpus=name,path=path,sha256=hashlib.sha256(Path(path).read_bytes()).hexdigest()))
(out/'input-sha256.json').write_text(json.dumps(inputs,indent=2)+'\n')
print('PASS three claimed rules: positive controls, requested corpora, ASan/UBSan/leaks, three byte mutants and released handle',flush=True)
