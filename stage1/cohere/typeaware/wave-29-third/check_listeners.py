#!/usr/bin/env python3
"""Numeric metadata versus live Go rule registrations; no shared driver edits."""
import json,os,re,subprocess,sys,time
from pathlib import Path
s=Path(__file__).resolve().parent;r=s.parents[3]
d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
c=Path('/workspace/wave29-regex-controls/adamic');archive=Path('/workspace/wave29-next-checker.a')
commands=[]
def run(name,args,cwd=r,env=None):
 start=time.monotonic_ns()
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:
  p=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode,ns=time.monotonic_ns()-start));(d/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 if p.returncode:raise RuntimeError(f'{name}: exit {p.returncode}; see logs')
 return (d/(name+'.stdout')).read_bytes()
v=r/'cohere/adamic_wave29_listener_oracle.go'
(d/'overlay.json').write_text(json.dumps({'Replace':{str(v):str(s/'testdata/listener_oracle.go')}}))
truth=run('go',['go','run','-overlay',d/'overlay.json',v],cwd=r/'cohere');assert len(truth.splitlines())==9
# Manifests are potential subscriptions, not factory registrations or full ports.
def manifest_bytes(manifests):
 return ''.join(row['name']+'\t'+''.join(str(kind)+',' for kind in row['kinds'])+'\n' for row in manifests).encode()
manifest_names=['id-denylist','id-match','nexus/concurrency-no-check-then-write','no-restricted-globals','no-setter-return','no-shadow-restricted-names','react-hooks/set-state-in-effect','react-hooks/set-state-in-render','react-hooks/static-components']
manifests=[json.loads((s/'rules'/name.replace('/','-')/'rule.json').read_text()) for name in manifest_names]
for name,row in zip(manifest_names,manifests):
 assert row['name']==name and row['kinds'] and all(type(kind) is int for kind in row['kinds'])
assert manifest_bytes(manifests)==truth
(d/'manifests.stdout').write_bytes(manifest_bytes(manifests))
for index,name in enumerate(manifest_names):
 mutated=json.loads(json.dumps(manifests));mutated[index]['kinds'][0]+=1
 assert manifest_bytes(mutated)!=truth
 print(name+': valid JSON wrong-kind mutant caught only by Go registration comparison',flush=True)
run('native-build',[c,'build',s/'listener_probe.a','-o',d/'native','--tsgo',archive])
assert run('native',[d/'native'])==truth and not(d/'native.stderr').read_bytes()
# Metadata itself never uses checker operations, so no checker shim is needed on Node.
assert run('node',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',s/'listener_probe.a'])==truth
assert not(d/'node.stderr').read_bytes()
run('asan-build',[c,'build',s/'listener_probe.a','-o',d/'asan','--tsgo',archive,'--sanitize'])
env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1')
assert run('asan',[d/'asan'],env=env)==truth and not(d/'asan.stderr').read_bytes()
rows=[
 ('denylist',s.parent/'id_denylist.a','listenerKinds'),('match',s.parent/'id_match.a','listenerKinds'),
 ('concurrency',s.parent/'concurrency_no_check_then_write.a','listenerKinds'),
 ('globals',s.parent/'wave-29-next/no_restricted_globals.a','listenerKinds'),
 ('setter',s.parent/'wave-29-next/no_setter_return.a','listenerKinds'),
 ('shadow',s.parent/'wave-29-next/no_shadow_restricted_names.a','listenerKinds'),
 ('effect',s/'react_listener_declarations.a','setStateInEffectKinds'),
 ('render',s/'react_listener_declarations.a','setStateInRenderKinds'),
 ('static',s/'react_listener_declarations.a','staticComponentsKinds'),
]
def imports(text,parent,redirect=None):
 def replace(m):
  target=(parent/m[2]).resolve();target=redirect.get(target,target) if redirect else target
  return 'from '+repr(str(target))
 return re.sub(r"from\s+(['\"])(\.[^'\"]+)\1",replace,text)
for name,module,export in rows:
 local=d/('mutant-'+name);local.mkdir(exist_ok=True)
 text=module.read_text();pattern=r'(export const '+export+r': readonly number\[\] = \[)(\d+)'
 matches=list(re.finditer(pattern,text));assert len(matches)==1
 text=re.sub(pattern,lambda m:m[1]+str(int(m[2])+1),text)
 mutated=local/module.name;mutated.write_text(imports(text,module.parent))
 (local/'listener_declarations.a').write_text(imports((s/'listener_declarations.a').read_text(),s,{module.resolve():mutated}))
 (local/'listener_probe.a').write_text((s/'listener_probe.a').read_text())
 run(name+'-build',[c,'build',local/'listener_probe.a','-o',local/'native','--tsgo',archive])
 got=run(name,[local/'native']);assert got!=truth and not(d/(name+'.stderr')).read_bytes()
 print(name+': wrong numeric subscription compiles and exits 0; only registration-byte comparison catches it',flush=True)
print('PASS: 9 numeric listener declarations; '+str(len(truth))+' bytes agree across Go/native/Node/sanitized native',flush=True)
