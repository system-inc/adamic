import pathlib,json,subprocess,time,sys
out=pathlib.Path(__file__).resolve().parent
root=pathlib.Path('/workspace/adamic')
def run(name,args,cwd=root):
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(args,cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
 (out/(name+'.exit.json')).write_text(json.dumps({'command':args,'exit':r.returncode},indent=2)+'\n')
 print(name+' child exit='+str(r.returncode),flush=True)
 if r.returncode:sys.exit(r.returncode)
while not (out/'lane.exit.json').exists():time.sleep(5)
assert json.loads((out/'lane.exit.json').read_text())['exit']==0
run('collect-lane',['python3',str(out/'collect-lane.py')])
assert subprocess.check_output(['git','status','--porcelain'],cwd=root,text=True)==''
base=subprocess.check_output(['git','rev-parse','HEAD^'],cwd=root,text=True).strip();assert base.startswith('838809f7a')
assert subprocess.check_output(['git','rev-list','--count',base+'..HEAD'],cwd=root,text=True).strip()=='1'
paths=subprocess.check_output(['git','diff','--name-only',base,'HEAD'],cwd=root,text=True).splitlines()
allowed=['stage3/adapt/62-temporary-tracing-write/','stage3/adapt/63-temporary-tracing-legend/','stage3/adapt/99-correctness-tracing-dumptypes/']
assert all(p=='stage3/upstream/LEDGER.md' or any(p.startswith(a) for a in allowed) for p in paths)
run('implementation-push',['git','push','origin','HEAD:refs/heads/typescript/tracing-fs-views'])
evidence=pathlib.Path('/workspace/tracing-evidence')
run('evidence-add',['git','add','review/tracing-fs-views'],evidence)
# Keep self-referential commit/push logs outside the committed evidence directory.
for name,args in [('evidence-commit',['git','commit','-m','baobab: complete all 398 shards and tracing retirement proofs']),('evidence-push',['git','push','origin','HEAD:refs/heads/typescript/tracing-fs-views-evidence'])]:
 with pathlib.Path('/tmp/tracing-final-'+name+'.log').open('w') as log:r=subprocess.run(args,cwd=evidence,stdout=log,stderr=subprocess.STDOUT)
 print(name+' child exit='+str(r.returncode),flush=True)
 if r.returncode:sys.exit(r.returncode)
print('implementation and complete evidence published',flush=True)
