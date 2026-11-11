import pathlib,time,json,subprocess,hashlib,shutil,sys
root=pathlib.Path('/workspace/overload-head');out=root/'stage3/census/overload-result-10b6c33'
def waitstat(path):
 p=pathlib.Path(path)
 while not p.exists():time.sleep(5)
 data=json.loads(p.read_text())
 if data['exit_code']!=0:raise RuntimeError(str(p)+': '+str(data))
def run(name,cmd,cwd=root):
 print('start '+name,flush=True)
 with open('/tmp/overload-'+name+'.log','w') as log:
  r=subprocess.run(cmd,cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
 print(name+' child_exit='+str(r.returncode),flush=True)
 if r.returncode:raise RuntimeError(name+' failed; see log')
print('waiting for apply',flush=True)
waitstat('/tmp/overload-apply.time.json')
files={}
for p in pathlib.Path('/workspace/overload-adapted/src').rglob('*'):
 if p.is_file():files[str(p.relative_to('/workspace/overload-adapted'))]=hashlib.sha256(p.read_bytes()).hexdigest()
(out/'adapted-source-hashes.json').write_text(json.dumps(files,indent=2)+'\n')
shutil.copyfile('/workspace/overload-adapted/patch-set.md',out/'patch-set.md')
run('checkpoint-apply',['python3','/workspace/overload-checkpoint.py','apply complete'])
print('waiting for both builds and setup',flush=True)
for label in ['parent','head']:
 for kind in ['entry-build','checker-build']:waitstat('/workspace/overload-build-'+label+'/'+kind+'.time.json')
while 'setup: done on ' not in pathlib.Path('/tmp/overload-setup.log').read_text():time.sleep(5)
for label in ['parent','head']:
 d=pathlib.Path('/workspace/overload-build-'+label)
 (out/(label+'-binary-hashes.json')).write_text(json.dumps({p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in [d/'entry-census',d/'census']},indent=2)+'\n')
run('parent-measure',['bash','/workspace/overload-measure.sh','parent'])
run('checkpoint-parent',['python3','/workspace/overload-checkpoint.py','parent run complete'])
run('head-measure',['bash','/workspace/overload-measure.sh','head'])
run('checkpoint-head',['python3','/workspace/overload-checkpoint.py','head run complete'])
run('pair',['python3','/workspace/overload-run.py',str(out/'pair.log'),str(out/'pair.time.json'),'python3','/workspace/overload-parent/stage3/census/pair/pair.py','/workspace/overload-build-parent/latent.jsonl','/workspace/overload-build-head/latent.jsonl','/workspace/overload-adapted','--typescript','/workspace/overload-parent/stage3/api/node_modules/typescript','--json',str(out/'delta.json'),'--report',str(out/'pair-report.md')])
run('analyze',['python3','/workspace/overload-analyze.py'])
for name in ['analyze.py','report.py','orchestrate.py']:
 shutil.copyfile('/workspace/overload-'+name,out/name)
run('report',['python3','/workspace/overload-report.py'])
print('all measurements and report complete',flush=True)
