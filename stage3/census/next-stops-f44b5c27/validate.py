import collections,json,subprocess,hashlib
from pathlib import Path
out=Path(__file__).resolve().parent
rows=[]
for i in range(1,80):
 pair=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
 assert len(pair)==2
 assert {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
 assert json.loads((out/f'root-{i:02}-status.json').read_text())['exit_code']==0
 rows.append(pair[0])
assert len({r['roots'][0] for r in rows})==79
normalize=lambda s:s.split('/src/compiler/')[-1]
old=json.loads((out/'previous-grouped-first-stops.json').read_text())
baseline={r:(g['where'],g['kind'],g['reason']) for g in old for r in g['roots']}
comparisons=json.loads((out/'comparison.json').read_text())
for row,c in zip(rows,comparisons):
 root=normalize(row['roots'][0]);after=(normalize(row.get('where','-')),row['kind'],row.get('reason','-'))
 assert tuple(c['before'])==baseline[root] and tuple(c['after'])==after
 assert c['classification']==('unchanged' if baseline[root]==after else 'moved' if after in baseline.values() else 'new')
groups=json.loads((out/'grouped-first-stops.json').read_text())
assert sum(g['count'] for g in groups)==79
phases=json.loads((out/'stop-phases.json').read_text())
for g in groups:
 assert g['where'] in phases
 if g['new']:
  assert phases[g['where']]['owner']
  assert phases[g['where']]['phase'] in ['refusal_scan','lowering','lowering (initialization preflight)']
manifest=json.loads((out/'adapted-source-manifest.json').read_text())
tree=Path('/workspace/scratch/next-stops-f44b5c27-adapted')
assert all(hashlib.sha256((tree/name).read_bytes()).hexdigest()==sha for name,sha in manifest.items())
binary=json.loads((out/'binaries.json').read_text())['ordinary']
assert hashlib.sha256(Path(binary['path']).read_bytes()).hexdigest()==binary['sha256']
assert sorted(r for g in groups for r in g['roots'])==sorted(baseline)
sha=json.loads((out/'provenance.json').read_text())['sha']
assert subprocess.run(['git','merge-base','--is-ancestor',sha,'HEAD']).returncode==0
assert all(n.startswith('stage3/census/next-stops-f44b5c27/') for n in subprocess.check_output(['git','diff',sha,'--name-only'],text=True).splitlines())
assert not list(out.rglob('*.go'))
processes=subprocess.check_output(['ps','-eo','comm'],text=True).splitlines()
assert 'next-stops-cens' not in processes and 'next-stops-censu' not in processes
failed=json.loads((out/'build-before-setup-status.json').read_text())
assert failed['exit_code']==1 and failed['seconds']<1800 and failed['process_group_cleaned']
statuses=[json.loads(f.read_text()) for f in out.glob('*-status.json') if f.name!='build-before-setup-status.json']
assert all(s['exit_code']==0 and s['seconds']<1800 for s in statuses)
assert all(s['process_group_cleaned'] for s in statuses)
result=dict(all_final_children_exit_zero=True,documented_initial_build_exit=1,all_recorded_child_seconds_below_1800=True,roots=79,raw_observations=158,duplicates_agree=True,baseline_comparison_verified=True,tracked_source_clean=True,no_live_census_children=True,sha=sha,classification=dict(collections.Counter(c['classification'] for c in comparisons)))
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
