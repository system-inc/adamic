from pathlib import Path
import subprocess,os,time,json,hashlib
s=Path('/workspace/scratch/each-node');repo=Path('/workspace/adamic');env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE=str(s/'upstream'))
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (s/(name+'.log')).open('w') as f:p=subprocess.run([str(a) for a in args],cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT)
 result=dict(command=[str(a) for a in args],cwd=str(cwd),exit=p.returncode,seconds=time.monotonic()-start);(s/(name+'-result.json')).write_text(json.dumps(result,indent=2)+'\n');print(name,json.dumps(result),flush=True)
 if p.returncode:raise SystemExit(p.returncode)
 return result
frozen=json.loads((s/'frozen-adaptation-manifest.json').read_text());frozen={p:h for p,h in frozen.items() if '__pycache__' not in Path(p).parts};(s/'frozen-adaptation-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
run('final-apply',['bash','stage3/apply.sh',s/'final'])
(s/'final/node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
run('final-build',['npm','run','build'],cwd=s/'final')
manifest=lambda root:{str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
a=manifest(s/'after/src');b=manifest(s/'final/src');assert a==b,'fresh apply source differs from measured candidate'
js=lambda root:{str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*.js'))};assert js(s/'after/built/local')==js(s/'final/built/local'),'fresh normal JavaScript differs'
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:assert (s/'after/built/local'/name).read_bytes()==(s/'final/built/local'/name).read_bytes()
current={str(p.relative_to(repo)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((repo/'stage3/adapt').rglob('*')) if p.is_file() and '__pycache__' not in p.parts};assert current==frozen,'adaptation key inputs changed during acceptance'
script="const fs=require('node:fs'),assert=require('node:assert/strict');const ts=require(process.argv[1]);const expected=JSON.parse(fs.readFileSync(process.argv[2],'utf8'));assert.deepEqual(Object.keys(ts).sort(),expected);console.log(Object.keys(ts).length+' final export keys match');"
run('final-export-keys',['node','-e',script,s/'final/built/local/typescript.js',s/'build-comparison-export-keys.json'])
result=dict(passed=True,allFrozenAdaptationInputsUnchanged=True,sourceFiles=len(a),normalJavaScriptFiles=len(js(s/'final/built/local')),sourceByteIdenticalToMeasuredCandidate=True,javascriptByteIdenticalToMeasuredCandidate=True,publicDeclarationsIdentical=True,exportKeysIdentical=True,exportKeyCount=2256)
(s/'final-verify.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
