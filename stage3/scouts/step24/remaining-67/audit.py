#!/usr/bin/env python3
"""Check complete dispositions, measured outcomes and fixture headers."""
import json,re
from pathlib import Path
u=Path(__file__).resolve().parent
rows=json.loads((u/'dispositions.json').read_text()); comparison=json.loads((u/'parser-comparison.json').read_text())
fixtures={r['file']:r for r in json.loads((u/'fixture-results.json').read_text())}
assert len(rows)==67 and len({r['id'] for r in rows})==67
assert {r['id'] for r in rows}=={r['id'] for r in comparison['before']}
assert len(comparison['after'])==53 and not comparison['new']
assert sum(r['afterStatus']=='cleared' for r in rows)==14
for row in rows:
 assert row['witness'] in fixtures
 assert all(row['adaptations'].values())
 assert row['owner'] and row['disposition']
mutants=[]
for name,result in fixtures.items():
 file=u/'fixtures'/name;source=file.read_text()
 def check(text):assert text.splitlines()[0]=='// a-check: type error TS'+str(result['code'])
 check(source)
 for mutant in [source.split('\n',1)[1],source.replace('type error TS'+str(result['code']),'type error TS9999',1)]:
  try:check(mutant)
  except AssertionError:mutants.append({'file':name,'caught':'measured first diagnostic header'})
  else:raise AssertionError('header mutant survived')
 assert result['nodeExit']==0 and result['stderr']=='' and result['stdout']!=result['mutantStdout']
(u/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
print(json.dumps({'completeStops':len(rows),'remaining':len(comparison['after']),'witnesses':len(fixtures),'nodeMutants':len(fixtures),'headerMutants':len(mutants)}))
