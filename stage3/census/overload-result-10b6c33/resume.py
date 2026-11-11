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
run('parent-measure-resume',['bash','/workspace/overload-measure.sh','parent'])
run('checkpoint-parent',['python3','/workspace/overload-checkpoint.py','parent run complete'])
run('head-measure',['bash','/workspace/overload-measure.sh','head'])
run('checkpoint-head',['python3','/workspace/overload-checkpoint.py','head run complete'])
run('pair',['python3','/workspace/overload-run.py',str(out/'pair.log'),str(out/'pair.time.json'),'python3','/workspace/overload-parent/stage3/census/pair/pair.py','/workspace/overload-build-parent/latent.jsonl','/workspace/overload-build-head/latent.jsonl','/workspace/overload-adapted','--typescript','/workspace/overload-parent/stage3/api/node_modules/typescript','--json',str(out/'delta.json'),'--report',str(out/'pair-report.md')])
run('analyze',['python3','/workspace/overload-analyze.py'])
for name in ['analyze.py','report.py','orchestrate.py']:
 shutil.copyfile('/workspace/overload-'+name,out/name)
run('report',['python3','/workspace/overload-report.py'])
print('all measurements and report complete',flush=True)
