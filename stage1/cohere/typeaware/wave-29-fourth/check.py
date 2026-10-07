#!/usr/bin/env python3
"""Independent production Go versus numeric AST kernels; not a full lint oracle."""
import hashlib,json,os,re,subprocess,sys,time
from pathlib import Path
s=Path(__file__).resolve().parent;r=s.parents[3];d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
c=Path(os.environ.get('ADAMIC_COMPILER','/workspace/wave29-latest-kernel-adamic'))
commands=[]
def run(name,args,cwd=r,env=None):
 start=time.monotonic_ns()
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:
  p=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode,ns=time.monotonic_ns()-start));(d/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 if p.returncode:raise RuntimeError(f'{name}: exit {p.returncode}; see logs')
 return (d/(name+'.stdout')).read_bytes()
v=r/'cohere/internal/lint/rules/react/adamic_wave29_fourth_test.go'
(d/'overlay.json').write_text(json.dumps({'Replace':{str(v):str(s/'testdata/kernel_oracle_test.go')}}))
run('go',['go','test','-overlay',d/'overlay.json','./internal/lint/rules/react','-run','^TestWave29FourthKernels$','-count=1','-v'],cwd=r/'cohere',env=dict(os.environ,ADAMIC_FOURTH_ARTIFACTS=str(d)))
truth=(d/'go.expected').read_bytes();frames=d/'cases.frames'
assert len(truth.splitlines())>=79
kinds=json.loads((d/'kinds.json').read_text())
for name,value in kinds.items():assert f'export const {name} = {value};' in (s/'syntax_kinds.a').read_text()
for name,values in json.loads((d/'listeners.json').read_text()).items():
 manifest=json.loads((s/'rules'/name.replace('/','-')/'rule.json').read_text());assert manifest==dict(name=name,kinds=values)
 module=(s/'rules'/name.replace('/','-')/'rule.a').read_text();assert ('export const listenerKinds: readonly string[] = '+json.dumps(values)+';') in module
run('native-build',[c,'build',s/'kernel_probe.a','-o',d/'native'])
assert run('native',[d/'native',frames])==truth and not(d/'native.stderr').read_bytes()
assert run('node',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',s/'kernel_probe.a',frames])==truth
assert not(d/'node.stderr').read_bytes()
js=run('javascript-build',[c,'js',s/'kernel_probe.a']);(d/'probe.mjs').write_bytes(js)
assert run('javascript',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',d/'probe.mjs',frames])==truth
assert not(d/'javascript.stderr').read_bytes()
run('asan-build',[c,'build',s/'kernel_probe.a','-o',d/'asan','--sanitize'])
env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1')
assert run('asan',[d/'asan',frames],env=env)==truth and not(d/'asan.stderr').read_bytes()
def imports(text,parent,redirect=None):
 def replace(m):
  target=(parent/m[2]).resolve();target=redirect.get(target,target) if redirect else target
  return 'from '+repr(str(target))
 return re.sub(r"from\s+(['\"])(\.[^'\"]+)\1",replace,text)
mutants=[
 ('fragment','react-jsx-fragments',"module.text === 'react'","module.text === 'preact'"),
 ('undef','react-jsx-no-undef',"name.includes('-')","name.includes('_')"),
 ('context','react-jsx-no-constructed-context-values','return first.node >= 0 || node.b < 0 ? first : construction(tree, tree.get(node.b), depth + 1);','return first.node < 0 || node.b < 0 ? first : construction(tree, tree.get(node.b), depth + 1);'),
]
for name,directory,before,after in mutants:
 module=s/'rules'/directory/'rule.a';text=module.read_text();assert text.count(before)>=1
 local=d/('mutant-'+name);local.mkdir(exist_ok=True);mutated=local/'rule.a'
 mutated.write_text(imports(text.replace(before,after,1),module.parent))
 (local/'probe.a').write_text(imports((s/'kernel_probe.a').read_text(),s,{module.resolve():mutated}))
 run(name+'-build',[c,'build',local/'probe.a','-o',local/'native'])
 got=run(name,[local/'native',frames]);assert got!=truth and not(d/(name+'.stderr')).read_bytes()
 differences=sum(a!=b for a,b in zip(truth.splitlines(),got.splitlines()))
 print(name+': compiles, exit 0, empty stderr; Go bytes catch '+str(differences)+' differing rows',flush=True)
print('PASS: '+str(len(truth.splitlines()))+' kernel/callback rows, '+str(len(truth))+' identical bytes; sha256='+hashlib.sha256(truth).hexdigest(),flush=True)
print('Go/native/Node/emitted JS/sanitized native agree; 3 listener manifests match live registrations',flush=True)
print('No full source-rule findings/corpora, bridge/lifetime suite or full lint timing is claimed',flush=True)
