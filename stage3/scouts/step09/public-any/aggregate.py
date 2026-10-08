#!/usr/bin/env python3
"""Join every CLI capture to its acceptance project and retain complete shape observations."""
import collections
import hashlib
import json
from pathlib import Path
import sys
H=Path(__file__).resolve().parent
observations,acceptance=map(Path,sys.argv[1:])
records=[]
for file in sorted(observations.glob('*.json')):
 row=json.loads(file.read_text());row['project']=Path(row['cwd']).name;records.append(row)
report=json.loads((acceptance/'report.json').read_text())
assert report['cases']==report['passed']==301 and not report['failed']
assert len(records)==301 and len({r['project'] for r in records})==301
expected={p.parent.name for p in acceptance.glob('*/actual.exit')}
assert {r['project'] for r in records}==expected,'missing project instrumentation'
by_site={}
for row in records:
 for event in row['records']:
  target=by_site.setdefault(event['site'],dict(projects=set(),phases=collections.Counter(),domains={}))
  target['projects'].add(row['project']);target['phases'][event['phase']]+=event['count']
  key=json.dumps([event['phase'],event['shape']],sort_keys=True)
  if key not in target['domains']:target['domains'][key]=dict(phase=event['phase'],shape=event['shape'],count=0)
  target['domains'][key]['count']+=event['count']
for target in by_site.values():
 target['projects']=sorted(target['projects']);target['phases']=dict(target['phases']);target['domains']=list(target['domains'].values())
result=dict(acceptance=report,projects=len(records),by_site=by_site,
            excluded_observation_claims=['watch/build timers not invoked by noEmit acceptance','public-only tree-shaken APIs not present in CLI'],
            captures_sha256={r['project']:hashlib.sha256(json.dumps(r,sort_keys=True).encode()).hexdigest() for r in records})
(H/'evidence/observations.json').write_text(json.dumps(result,indent=2)+'\n')
import gzip
with gzip.open(H/'evidence/project-captures.json.gz','wt') as f:json.dump(records,f)
print('PASS: all 301 projects have instrumentation captures and unchanged CLI goldens')
for name,value in sorted(by_site.items()):print(name,len(value['projects']),value['phases'])
