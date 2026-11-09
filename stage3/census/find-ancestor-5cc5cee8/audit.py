import json,gzip,hashlib,subprocess,re
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
def read(path):
 if not path.exists():path=Path(str(path)+'.gz')
 with (gzip.open(path,'rt') if path.suffix=='.gz' else path.open()) as f:return [json.loads(line) for line in f]
streams={s:read(out/(s+'-entry-run.jsonl')) for s in ['base','head']}
info=json.loads((out/'analysis.json').read_text());evidence={}
for side,rows in streams.items():
 assert len(rows)==83 and rows[-1]['status']=='relation_report'
 assert not rows[0]['checker_rejected'] and not rows[0]['diagnostics']
 units=[u for r in rows for u in r.get('units',[])];attempts=[u for u in units if 'relation_overflow' in u]
 relations=read(out/(side+'-entry-run-relations.jsonl'));timings=read(out/(side+'-entry-run-units.jsonl'))
 assert len(relations)==4061 and len(attempts)==4059 and len(timings)==4059
 assert relations[0]['where']==relations[-1]['where']=='program'
 assert [(r['where'],r['answer']) for r in relations[1:-1]]==[(u['where'],u['relation_overflow']) for u in attempts]
 assert [r['where'] for r in timings]==[u['where'] for u in attempts]
 onset=info[side]['onset_attempted_ordinal']
 assert all(not u['relation_overflow'] for u in attempts[:onset-1])
 assert all(u['relation_overflow'] for u in attempts[onset-1:])
 projected=read(out/(side+'-pair-input.jsonl.gz'))
 expected=json.loads(json.dumps(rows[:-1]))
 for r in expected:
  for u in r.get('units',[]):u.pop('relation_overflow',None)
 assert projected==expected
 findings=[f for r in rows for f in r.get('findings',[])]
 def clean(x):return x.replace('/workspace/scratch/bar-adapted/src/','')
 lines=sorted({' | '.join([clean(f.get(k,'')) for k in ['unit','phase','kind','where']]+[clean(f.get('reason',''))[:110]]) for f in findings})
 listpath=out/('findings-'+side+'.txt')
 if not listpath.exists():listpath=Path(str(listpath)+'.gz')
 with (gzip.open(listpath,'rt') if listpath.suffix=='.gz' else listpath.open()) as f:assert f.read().splitlines()==lines
 evidence[side]={'stream_uncompressed_sha256':hashlib.sha256(gzip.decompress((out/(side+'-entry-run.jsonl.gz')).read_bytes())).hexdigest(),'monotone_overflow':True,'projection_exact':True,'list_exact':True,'probe_rows':len(attempts)}
assert streams['base'][-1]['findings']==streams['head'][-1]['findings']
assert len(streams['base'][-1]['findings'])==1
for p in out.glob('*.status.json'):
 assert json.loads(p.read_text())['exit']==0,(p,p.read_text())
for side in ['base','head']:
 receipt=json.loads((out/('checkpoint-'+side+'-entry-run.json')).read_text());assert receipt['exit']==0
 assert subprocess.check_output(['git','rev-parse',receipt['sha']+'^'],cwd=repo,text=True).strip()==json.loads((out/'pins.json').read_text())['base']
 paths=subprocess.check_output(['git','diff-tree','--no-commit-id','--name-only','-r',receipt['sha']],cwd=repo,text=True).splitlines()
 assert all(x.startswith('stage3/census/find-ancestor-5cc5cee8/') for x in paths)
status=subprocess.check_output(['git','status','--porcelain','-uall'],cwd=repo,text=True).splitlines()
assert all(row[3:].startswith('stage3/census/find-ancestor-5cc5cee8/') for row in status)
processes=subprocess.check_output(['ps','-eo','pid,ppid,args'],text=True).splitlines()
live=[p for p in processes if re.search(r'(?:^|/)bar-(?:base|head)-bin(?: |$)',p)]
assert not live,live
(out/'processes-final.json').write_text(json.dumps({'live_census_processes':live,'all_started_sessions_completed':True},indent=2)+'\n')
(out/'audit.json').write_text(json.dumps({'streams':evidence,'run_end_findings_identical':True,'checkpoint_scope_verified':True,'working_tree_scope_verified':True,'all_child_statuses_zero':True,'process_cleanup_verified':True},indent=2)+'\n')
print('audit child_exit=0')
