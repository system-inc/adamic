import pathlib,json,gzip,collections,hashlib
out=pathlib.Path(__file__).resolve().parent
roots=json.loads((out/'roots.json').read_text());coverage=json.loads((out/'coverage.json').read_text());sites=json.loads((out/'sites.json').read_text());counts=json.loads((out/'by-callee.json').read_text());funcs=json.loads((out/'public-functions.json').read_text())
assert len(roots)==79 and len(set(roots))==79
assert sorted(s['file'] for s in coverage)==roots
raw_observations=set();diagnostics=set()
for s in coverage:
 stem=s['file'].replace('/','__').removesuffix('.ts');p=out/'raw'/(stem+'.jsonl');p=p if p.exists() else pathlib.Path(str(p)+'.gz')
 assert p.exists(),s['file']
 rows=[]
 with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as stream:
  for line in stream:
   try:rows.append(json.loads(line))
   except json.JSONDecodeError:assert s['child_exit']!=0,'truncated successful stream'
 assert not rows or rows[0].get('latent_mode')=='full'
 if rows:diagnostics.add((rows[0].get('checker_rejected'),len(rows[0].get('diagnostics',[]))))
 if s['child_exit']==0:
  assert rows[-1].get('status')=='relation_report'
  assert [r['file'] for r in rows if 'file' in r]==['/workspace/scratch/baobab-adapted/src/compiler/'+s['file']]
 for r in rows:
  for f in r.get('findings') or []:
   if 'SourceFileLike' in f.get('reason',''):raw_observations.add(tuple(f.get(k,'') for k in ['kind','where','reason','text','phase','unit']))
actual={tuple(f.get(k,'') for k in ['kind','where','reason','text','phase','unit']) for f in sites}
assert actual==raw_observations
physical={(f['location'],f['reason']) for f in sites}
assert sum(r['sites'] for r in counts)==len(physical)
assert all(r['reads_optional_method'] is not None for r in counts),'unreviewed callee'
assert all(f['callee'] in {i['function'] for i in funcs} for f in sites),'non-call or unreviewed callee needs explicit audit'
assert not any(p.stat().st_size>1000000 for p in (out/'raw').iterdir() if p.is_file() and p.suffix!='.gz'),'uncompressed raw evidence over 1 MB'
for p in out.rglob('*.go'):raise AssertionError('Go evidence must be .go.txt: '+str(p))
for name in ['REPORT.md','README.md']:
 text=(out/name).read_text();assert '\u2013' not in text and '\u2014' not in text
result=dict(attempted=len(coverage),completed=sum(s['child_exit']==0 and s['relation_trailer'] for s in coverage),timeouts=[s['file'] for s in coverage if s['over_five_minutes']],raw_observations=len(raw_observations),physical_sites=len(physical),diagnostic_states=[list(x) for x in sorted(diagnostics)],audit='passed')
(out/'audit-result.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
