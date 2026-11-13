"""Compare the complete before/after lane union, every count and baseline byte."""
import collections,difflib,hashlib,json,pathlib,os
repo=pathlib.Path('/workspace/adamic');base=pathlib.Path('/workspace/scratch')
output=pathlib.Path(__file__).resolve().parent
suffix=os.environ.get('TIMER_REPORT_SUFFIX','')
manifest=json.loads((repo/'stage3/lane/shards.json').read_text());units=manifest['units']
roots={'main':base/os.environ.get('TIMER_MAIN_LANE','timer-lane-main'),'adapted':base/os.environ.get('TIMER_AFTER_LANE','timer-lane-final')}
reports={label:{} for label in roots}
counts={label:collections.Counter() for label in roots}
rows=[]
for unit in units:
 name=unit['name'];pair={}
 for label,root in roots.items():
  report=json.loads((root/name/'report.json').read_text())
  assert report['status']=='pass',(label,name,report['errors'])
  assert report['counts']==unit['expect'],(label,name,report['counts'])
  assert report['seconds']<30,(label,name,report['seconds'])
  reports[label][name]=report;counts[label].update(report['counts']);pair[label]=report
  observations=json.loads((root/name/'observations.json').read_text())
  names=collections.Counter((r['task']['runner'],r['task']['file'],' '.join(r['name']),r['status']) for r in observations)
  pair[label+'Names']=names
  paths=report['baseline_diffs']
  assert paths==(['api/typescript.d.ts'] if name=='stage3-lane-byte-comparisons' else []),(label,name,paths)
 assert pair['main']['counts']==pair['adapted']['counts']
 assert pair['mainNames']==pair['adaptedNames'],name
 if name!='stage3-lane-byte-comparisons':
  assert (roots['main']/name/'oracle/baseline.diff').read_bytes()==(roots['adapted']/name/'oracle/baseline.diff').read_bytes()==b'',name
 rows.append(dict(name=name,counts=pair['adapted']['counts'],mainSeconds=pair['main']['seconds'],adaptedSeconds=pair['adapted']['seconds'],testIdentityAndStatusEqual=True,nonApiBaselinesByteIdentical=True))
expected={'pass':106366,'fail':1,'skip':0}
assert all(dict(c)==expected for c in counts.values()),counts
saved={label:root/'stage3-lane-byte-comparisons/saved-api' for label,root in roots.items()}
assert (saved['main']/'reference.d.ts').read_bytes()==(saved['adapted']/'reference.d.ts').read_bytes()
a=(saved['main']/'local.d.ts').read_text();b=(saved['adapted']/'local.d.ts').read_text()
diff=list(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='main/api/typescript.d.ts',tofile='adapted/api/typescript.d.ts'))
removed=[line[1:].strip() for line in diff if line.startswith('-') and not line.startswith('---')]
added=[line[1:].strip() for line in diff if line.startswith('+') and not line.startswith('+++')]
sanctions=json.loads((repo/'stage3/adapt/99-correctness-timer-closures/api-sanctions.json').read_text())
assert collections.Counter(removed)==collections.Counter(row['before'] for row in sanctions),(removed,sanctions)
assert collections.Counter(added)==collections.Counter(row['after'] for row in sanctions),(added,sanctions)
(output/('api-between-main-and-adapted'+suffix+'.diff.txt')).write_text(''.join(diff))
result=dict(status='pass',units=len(units),counts=expected,everyShardCountMatchesMain=True,everyTestIdentityAndStatusMatchesMain=True,allNonApiBaselinesByteIdentical=True,apiChangesExactlySixSanctionedMembers=True,maximumSeconds={label:max(r['seconds'] for r in rows.values()) for label,rows in reports.items()},shards=rows)
(output/('lane-comparison'+suffix+'.json')).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='shards'},indent=2))
