from pathlib import Path
import os,sys,json,shutil
r=Path('/workspace/loose-slot-scratch/head');s=r.parent;d=r/'stage3/census/loose-slot-census';sys.path.insert(0,str(d));from measure import run
out=d/'evidence/controls';out.mkdir(parents=True,exist_ok=True)
fixture=r/'internal/oracle/testdata/to_primitive/arrays.a'
run([s/'head-census',fixture,out/'arrays-production.json'],out/'arrays-production',cwd=r)
print((out/'arrays-production.json').read_text(),flush=True)
control=s/'arrays-control';control.mkdir(exist_ok=True);shutil.copyfile(fixture,control/'arrays.a')
run([s/'latent',control,out/'arrays-latent.jsonl'],out/'arrays-latent',cwd=r,env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1'))
records=[json.loads(x) for x in (out/'arrays-latent.jsonl').read_text().splitlines()]
print('latent records',[(x.get('file'),len(x.get('findings',[]) or [])) for x in records],flush=True)
run([s/'type-count',fixture,out/'arrays-type.json'],out/'arrays-type',cwd=r)
run([s/'type-count',fixture,out/'arrays-mutant.json'],out/'arrays-mutant',cwd=r,env=dict(os.environ,LOOSE_MUTANT_DISABLED='1'))
a=json.loads((out/'arrays-type.json').read_text());m=json.loads((out/'arrays-mutant.json').read_text())
assert a['Events'];assert not m['Events'];print('instrument events',len(a['Events']),'mutant',len(m['Events']),flush=True)
