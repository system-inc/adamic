set -e
source /workspace/adamic-tools/env.sh
python3 /tmp/u037-survivors.py > /tmp/u037-survivor-progress.log 2>&1
python3 /tmp/u037-probes.py > /tmp/u037-probe-progress.log 2>&1
python3 /tmp/u037-report.py > /tmp/u037-report-progress.log 2>&1
python3 - <<'PY'
import pathlib,json,subprocess,os,difflib
p=pathlib.Path('review/test-audit/internal-lower-module_namespace');report=json.loads((p/'report.json').read_text());needed=sorted(set(r['subsumed_by'][0] for r in report if r['verdict']=='subsumed' and r['subsumer_seconds'] is None));times={}
for n in needed:
 values=[]
 for i in range(3):
  with (p/'logs'/('timing-'+n+'-'+str(i)+'.log')).open('w') as out:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+n+'$'],stdout=out,stderr=subprocess.STDOUT)
  assert r.returncode==0,n
  events=[json.loads(s) for s in (p/'logs'/('timing-'+n+'-'+str(i)+'.log')).read_text().splitlines() if s.startswith('{')];values.append(next(e['Elapsed'] for e in reversed(events) if e['Action']=='pass' and not e.get('Test')))
 times[n]=sorted(values)[1]
(p/'subsumer-times.json').write_text(json.dumps(times,indent=2))
# Save the joint no-answer probe and validate it separately.
plan=json.loads((p/'probe-plan.json').read_text());joint=[q for q in plan if q['id'] in ['P05','P06']];original={q['file']:pathlib.Path(q['file']).read_text() for q in joint};diff=''
try:
 for q in joint:
  before=original[q['file']];after=before.replace(q['entry'],q['entry']+'\n\tif true { '+q['empty']+' }');pathlib.Path(q['file']).write_text(after);diff+=''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+q['file'],tofile='b/'+q['file']))
 with (p/'logs'/'vet-P08.log').open('w') as out:subprocess.run(['go','vet','./internal/lower/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 (p/'probes'/'P08.diff').write_text(diff)
finally:
 for f,t in original.items():pathlib.Path(f).write_text(t)
# Remove the unnecessary blank-identifier no-op from M09; the switch already deletes the assignment exactly.
mutants=json.loads((p/'mutant-plan.json').read_text());m=next(q for q in mutants if q['id']=='M09');m['new']=m['new'].replace('_ = declaration','');m['kind']='drop statement';before=pathlib.Path(m['file']).read_text();after=before.replace(m['old'],m['new']);(p/'diffs'/'M09.diff').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
try:
 pathlib.Path(m['file']).write_text(after)
 with (p/'logs'/'vet-M09-repaired.log').open('w') as out:subprocess.run(['go','vet','./internal/lower/'],stdout=out,stderr=subprocess.STDOUT,check=True)
finally:pathlib.Path(m['file']).write_text(before)
(p/'mutant-plan.json').write_text(json.dumps(mutants,indent=2))
PY
python3 /tmp/u037-report.py > /tmp/u037-report-progress.log 2>&1
