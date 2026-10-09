#!/usr/bin/env python3
"""Assemble successful checks from an interrupted gate and exact completion scopes."""
import gzip,json,pathlib,sys
scratch=pathlib.Path(sys.argv[1]);destination=pathlib.Path(sys.argv[2])
plan=json.loads((scratch/'completion-plan.json').read_text())
def read(directory):
 rows=[json.loads(p.read_text()) for p in directory.glob('*.json')]
 result={(r['Test'],r['Number']):r for r in rows}
 assert len(result)==len(rows)
 return result
def beneath(name,prefix):return name==prefix or name.startswith(prefix+'/')
rows=read(scratch/'final-observations')
scopes=[('completion-objects',[plan['completed_native_subtree']]),('completion-rest',plan['completion_rest']),('completion-schedule-bench',[plan['schedule_subtrees'][0]+'/tsan/'+n for n in ['3','7','64','unset']]),('completion-schedule-objects',[plan['schedule_subtrees'][1]])]
for directory,prefixes in scopes:
 log=[json.loads(line) for line in (scratch/(directory+'.jsonl')).read_text().splitlines() if line.startswith('{')]
 assert log[-1]['Action']=='pass' and not any(r['Action']=='fail' for r in log), directory
 rows={key:value for key,value in rows.items() if not any(beneath(key[0],prefix) for prefix in prefixes)}
 observed=read(scratch/directory)
 for key,value in observed.items():
  if key in rows:assert rows[key]['Run']==value['Run'], ('repeated ancestor output differs',key)
  rows[key]=value
with gzip.open(destination,'wt',compresslevel=9) as output:
 for key in sorted(rows):output.write(json.dumps(rows[key],separators=(',',':'))+'\n')
print('assembled',len(rows),'execution records')
