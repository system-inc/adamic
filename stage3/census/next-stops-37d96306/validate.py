"""Verify 316 paired observations, limits, source/binary integrity and exact comparisons."""
import hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/next-stops-37d96306');tree=scratch/'adapted'
read=lambda p:json.loads(p.read_text())
provenance=read(base/'provenance.json')
for revision,sha in [('head',provenance['sha']),('base',provenance['baseline'])]:
 assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=scratch/revision,text=True).strip()==sha
 info=subprocess.check_output(['go','version','-m',str(scratch/(revision+'-census'))],text=True)
 assert 'vcs.revision='+sha in info and 'vcs.modified=false' in info
 assert info==read(base/'build-info.json')[revision]
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=scratch/'typescript',text=True).strip()==provenance['upstream']
expected=list(map(str,sorted((tree/'src/compiler').rglob('*.ts'))));assert len(expected)==79
statuses={str(p.relative_to(base)):read(p) for p in base.rglob('*-status.json')}
for s in statuses.values():assert isinstance(s['exit_code'],int) and s['exit_code']==0 and s['seconds']<1800 and s['process_group_cleaned'] and not s['stopping_reason']
for revision in ['head','base']:
 out=base/revision/'CommonJSBundle';rows=[]
 for i in range(1,80):
  pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
  assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
  assert (out/f'root-{i:02}-status.json').exists();rows.append(pair[0])
 assert [r['roots'][0] for r in rows]==expected
 assert all(r['module_format']=='CommonJSBundle' for r in rows)
 assert sum(g['count'] for g in read(out/'grouped-first-stops.json'))==79
 assert all(g.get('phase') and g.get('emitter') for g in read(out/'grouped-first-stops.json'))
 for a,z in [(1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79)]:assert (out/f'group-{a}-{z}-status.json').exists()
 binary=read(base/'binaries.json')[revision];assert hashlib.sha256(Path(binary['path']).read_bytes()).hexdigest()==binary['sha256']
 assert subprocess.check_output(['git','status','--porcelain','--untracked-files=no'],cwd=scratch/revision,text=True)==''
manifest=read(base/'adapted-source-manifest.json');assert manifest==read(base/'adapted-source-before-measurement.json')
assert all(hashlib.sha256((tree/n).read_bytes()).hexdigest()==h for n,h in manifest.items())
cs=read(base/'comparison.json');baseline=read(base/'baseline.json');assert len(cs)==79
for c in cs:
 assert c['before']==baseline[c['root']] and c['unchanged']==(c['before']==c['after'])
 # No diagnostic or fix-only change may be hidden by an equal stop tuple.
 assert c['unchanged']==c['observation_unchanged']
provenance=read(base/'provenance.json')
assert subprocess.check_output(['git','diff',provenance['baseline'],provenance['sha'],'--','stage3'],text=True)==''
assert all(n.startswith('stage3/census/next-stops-37d96306/') for n in subprocess.check_output(['git','diff',provenance['baseline'],'--name-only'],text=True).splitlines())
assert not list(base.rglob('*.go'))
assert not [p for p in base.rglob('*') if p.is_file() and p.stat().st_size>20_000_000]
processes=subprocess.check_output(['ps','-eo','stat=,comm='],text=True).splitlines()
assert not [line for line in processes if any(n in line for n in ['head-census','base-census']) and not line.split()[0].startswith('Z')]
result=dict(roots_per_revision=79,raw_observations=316,duplicates_agree=True,all_child_exits={n:s['exit_code'] for n,s in statuses.items()},maximum_child_seconds=max(s['seconds'] for s in statuses.values()),adapted_source_hashes_verified=True,binaries_verified=True,measurement_worktrees_clean=True,stage3_identical=True,no_live_census_children=True)
(base/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print('validation passed')
