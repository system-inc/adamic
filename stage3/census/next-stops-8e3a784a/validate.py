"""Validate measured roots, provenance, source integrity, limits and evidence."""
import hashlib,json,os,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent;out=base/os.environ['CENSUS_MODULE_FORMAT'];tree=Path('/workspace/scratch/next-stops-8e3a784a-adapted')
expected=list(map(str,sorted((tree/'src/compiler').rglob('*.ts'))));assert len(expected)==79
rows=[]
for i in range(1,80):
 pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
 assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
 status=json.loads((out/f'root-{i:02}-status.json').read_text());assert status['exit_code']==0 and status['seconds']<1800 and status['process_group_cleaned']
 rows.append(pair[0])
assert [r['roots'][0] for r in rows]==expected
assert sum(g['count'] for g in json.loads((out/'grouped-first-stops.json').read_text()))==79
for g in json.loads((out/'grouped-first-stops.json').read_text()):assert g['phase'] and g['owner']
manifest=json.loads((base/'adapted-source-manifest.json').read_text());assert all(hashlib.sha256((tree/n).read_bytes()).hexdigest()==h for n,h in manifest.items())
assert manifest==json.loads((base/'adapted-source-before-remaining.json').read_text())
binary=json.loads((base/'binaries.json').read_text())['ordinary'];assert hashlib.sha256(Path(binary['path']).read_bytes()).hexdigest()==binary['sha256']
assert all(json.loads((out/f'group-{a}-{z}-status.json').read_text())['exit_code']==0 for a,z in [(1,10),(11,20),(21,30),(31,40),(41,50),(51,60),(61,70),(71,79)])
assert all(r['module_format']==os.environ['CENSUS_MODULE_FORMAT'] for r in rows)
comparisons=json.loads((out/'comparison.json').read_text());baseline=json.loads((out/'baseline.json').read_text())
for row,c in zip(rows,comparisons):
 root=row['roots'][0].split('/src/compiler/')[-1];after=[row.get('where','-').split('/src/compiler/')[-1],row['kind'],row.get('reason','-')]
 assert c['root']==root and c['before']==baseline[root] and c['after']==after and c['unchanged']==(c['before']==after)
sha=json.loads((base/'provenance.json').read_text())['sha']
assert all(n.startswith('stage3/census/next-stops-8e3a784a/') for n in subprocess.check_output(['git','diff',sha,'--name-only'],text=True).splitlines())
assert not [p for p in base.rglob('*') if p.is_file() and p.stat().st_size>20_000_000]
statuses={str(p.relative_to(base)):json.loads(p.read_text()) for p in base.rglob('*-status.json')}
assert all(isinstance(s['exit_code'],int) and s['seconds']<1800 and s['process_group_cleaned'] for s in statuses.values())
processes=subprocess.check_output(['ps','-eo','stat=,comm='],text=True).splitlines();assert not [line for line in processes if Path(binary['path']).name[:15] in line and not line.split()[0].startswith('Z')]
result=dict(roots=79,observations=158,duplicates_agree=True,root_exits_zero=True,source_clean_outside_evidence=True,all_child_exits={n:s['exit_code'] for n,s in statuses.items()},all_child_durations_below_1800=True,adapted_source_hashes_verified=True,no_live_census_children=True)
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n');print('validation passed')
