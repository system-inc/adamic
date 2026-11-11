import os,pathlib,json,time,subprocess,sys
out=pathlib.Path(__file__).resolve().parent
root=pathlib.Path('/workspace/adamic')
def run(name,command,env=None):
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,env=env)
 (out/(name+'.exit.json')).write_text(json.dumps({'command':command,'exit':r.returncode},indent=2)+'\n')
 print(name+' child exit='+str(r.returncode),flush=True)
 if r.returncode:sys.exit(r.returncode)
while not (out/'candidate-prepare-pinned.exit.json').exists():time.sleep(5)
assert json.loads((out/'candidate-prepare-pinned.exit.json').read_text())['exit']==0
p=json.loads((out/'candidate-prepare-pinned.log').read_text())
artifact=pathlib.Path(p['artifact']);tree=artifact/'tree'
after=pathlib.Path('/workspace/tracing-after')
if not after.exists():after.symlink_to(tree,target_is_directory=True)
measured=pathlib.Path('/workspace/tracing-measured-after/src')
expected={p.relative_to(measured):p.read_bytes() for p in measured.rglob('*') if p.is_file()}
actual={p.relative_to(tree/'src'):p.read_bytes() for p in (tree/'src').rglob('*') if p.is_file()}
assert actual==expected,'fresh apply differs from measured source'
(out/'source-equality.json').write_text(json.dumps({'files':len(actual),'identical':True},indent=2)+'\n')
(out/'artifact-ready.json').write_bytes((artifact/'ready.json').read_bytes())
for phase in ['apply','install','build','harness']:(out/('candidate-'+phase+'.log')).write_bytes((artifact/(phase+'.log')).read_bytes())
run('normal-proof',['python3',str(out/'normal-proof.py')])
while not (out/'core-production-first-stop.exit.json').exists():time.sleep(5)
run('verify-census',['python3',str(out/'verify-census.py')])
env=dict(os.environ,STAGE3_ARTIFACT=str(artifact),ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-6.0.3')
run('lane',['python3','-B','stage3/lane/measure_shards.py','/workspace/tracing-lane-run','--concurrency','2'],env)
ledger=json.loads(pathlib.Path('/workspace/tracing-lane-run/ledger.json').read_text())
assert len(ledger['units'])==398
assert all(r['exit']==0 and r['status']=='pass' for r in ledger['units'])
(out/'lane-ledger.json').write_text(json.dumps(ledger,indent=2)+'\n')
print('all final proofs passed',flush=True)
