import collections,hashlib,json,subprocess
from pathlib import Path
here=Path(__file__).resolve().parent;root=here.parents[2];source=Path('/workspace/graph-boundary-adapted')
reference=json.loads((here/'reference-adapted-manifest.json').read_text())
changed=[]
for name,expected in reference.items():
 p=source/name
 if not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest()!=expected:changed.append(name)
assert not changed,changed
assert subprocess.check_output(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],text=True).strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
assert not subprocess.check_output(['git','diff','bcbe5d12','--','internal','stage3/adapt','stage3/api','go.mod','go.work'],cwd=root,text=True).strip()
rows=[json.loads(s) for s in Path('/workspace/scratch/jawqwv7/selftest-boundary.jsonl').read_text().splitlines()]
assert len(rows)==2
assert collections.Counter(s['kind'] for s in rows[0]['sites'])==dict(views=2,record_literals=1,dynamic_reads=2,checked_object_calls=1)
assert rows[0]['graph_types_nonempty'] and rows[0]['status']=='lowered'
assert rows[1]['status']=='stopped' and rows[1]['sites']==[]
fixture=[json.loads(s) for s in Path('/workspace/scratch/jawqwv7/fixture-latent.jsonl').read_text().splitlines()]
assert len(fixture[0]['files'])==2 and len(fixture)==3
assert fixture[0]['latent_mode']=='first-error' and fixture[0]['status']=='measurement'
result=dict(entry_reach_fixture_passed=True,reference_files_matched=len(reference),typescript_pin_matched=True,compiler_and_adapter_unchanged=True,selftest_passed=True)
Path('/workspace/scratch/jawqwv7/apply-validation.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
