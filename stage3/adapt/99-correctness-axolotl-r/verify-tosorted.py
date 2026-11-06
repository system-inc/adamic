"""Verify completed full applies without modifying either measured source tree."""
import hashlib,json,os,shutil,subprocess
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/tosorted/step2');raw=s/'raw';out=repo/'stage3/adapt/99-correctness-axolotl-r'
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source');results={}
def run(name,args,cwd=repo,extra=None):
 with (raw/(name+'.log')).open('w') as log:r=subprocess.run(args,cwd=cwd,env={**env,**(extra or {})},stdout=log,stderr=log)
 results[name]={'command':[str(a) for a in args],'cwd':str(cwd),'exit':r.returncode};(raw/'verification.json').write_text(json.dumps(results,indent=2)+'\n');assert r.returncode==0,(name,r.returncode);print(name,'exit 0',flush=True)
pre=s/'contract-input/src/compiler';pre.mkdir(parents=True)
for file in ['core.ts','checker.ts']:(pre/file).write_bytes(subprocess.check_output(['git','-C',str(s/'before'),'show','refs/stage3/99-correctness-axolotl-r-before:src/compiler/'+file]))
run('contract',['node',out/'contract.cjs',s/'contract-input',s/'bin/adamic'],extra={'ADAMIC_RELATION_EVIDENCE':str(raw/'contract-relation')})
run('relation-mutants',['node',out/'relation-mutants.cjs',s/'after',s/'bin/adamic',raw/'actual-relation-mutants'])
run('adapter-replay',['node',out/'adapt.cjs',s/'after'])
roots=sorted((s/'after/src/compiler').rglob('*.ts'));assert len(roots)==79
(raw/'types-roots.json').write_text(json.dumps([str(p) for p in roots],indent=2)+'\n')
run('types-after',[s/'bin/adamic','types',*roots])
for side in ['before','after']:
 tree=s/side
 if not (tree/'node_modules').exists():(tree/'node_modules').symlink_to('/workspace/typescript-source/node_modules')
 run('build-'+side,['npm','run','build:compiler'],tree)
run('comparison',['node',out/'compare-tosorted.cjs',s/'before',s/'after',s/'after'])
comparison=json.loads((raw/'comparison.log').read_text());assert comparison['emittedJavascriptFiles']==9
(out/'tosorted-comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
manifest={str(p.relative_to(s/'after/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/'after/src').rglob('*')) if p.is_file()}
(s/'ready-for-census.json').write_text(json.dumps(manifest,indent=2)+'\n')
mutant=s/'mutable-arm-mutant';shutil.copytree(s/'after/src',mutant/'src',symlinks=True)
run('mutable-arm-mutant-patch',['node',out/'mutable-arm-mutant.cjs',mutant])
run('mutable-arm-mutant',[s/'bin/tool',mutant/'src/compiler/binder.ts',raw/'mutable-arm-mutant.jsonl'])
rows=[json.loads(line) for line in (raw/'mutable-arm-mutant.jsonl').read_text().splitlines()];assert len(rows)==2
assert all(r['kind']=='Refused' and r['where'].endswith('/core.ts:1049:34') and r['reason']=="a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold" for r in rows)
run('ordinary-after-binder',[s/'bin/tool',s/'after/src/compiler/binder.ts',raw/'ordinary-after-binder.jsonl'])
print('All verification checks passed',flush=True)
