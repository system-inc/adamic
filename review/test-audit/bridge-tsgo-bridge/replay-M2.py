from pathlib import Path
import subprocess,time,json,os
p=Path('review/test-audit/bridge-tsgo-bridge'); f=Path('internal/lower/tsgo.go'); original=f.read_text(); rows=['TestTSGoRequiresLink','TestBridgeUnitsCoverEveryPiece','TestBridgeProductCacheIsVerified']; timings=[]
try:
 subprocess.run(['git','apply','--check',str(p/'diffs/M2.diff')],check=True); subprocess.run(['git','apply',str(p/'diffs/M2.diff')],check=True)
 for corpus in [False,True]:
  directory=p/'corpus' if corpus else p; env=os.environ.copy(); env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u001/cache/M2-final'; env.pop('ADAMIC_MUTANT',None)
  if corpus: env['ADAMIC_TSGO_CORPUS']='/tmp/u001/corpus'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./bridge/tsgo/','-run','^('+'|'.join(rows)+')$']; start=time.monotonic()
  with (directory/'M2.log').open('w') as out: result=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  events=[json.loads(s) for s in (directory/'M2.log').read_text().splitlines() if s.startswith('{')]; failed=[e['Test'] for e in events if e['Action']=='fail' and e.get('Test') in rows]; passed=[e['Test'] for e in events if e['Action']=='pass' and e.get('Test') in rows]
  record=dict(id='M2',command=('ADAMIC_TSGO_CORPUS=/tmp/u001/corpus ' if corpus else '')+'ADAMIC_BUILD_CACHE_DIR=/tmp/u001/cache/M2-final '+' '.join(cmd)+' (standalone M2.diff applied)',exit=result.returncode,wall_seconds=time.monotonic()-start,failed=failed,passed=passed,unknown=[r for r in rows if r not in failed+passed],bounded=True,matrix_rows=rows,failure_lines=[])
  records=json.loads((directory/'matrix.json').read_text()); records=[record if x['id']=='M2' else x for x in records]; (directory/'matrix.json').write_text(json.dumps(records,indent=2)+'\n'); timings.append(record); print('M2',corpus,result.returncode,passed,flush=True)
 with (p/'survivor-after.log').open('w') as out: result=subprocess.run(['timeout','90','go','run','./review/test-audit/bridge-tsgo-bridge/survivor-observe.go'],stdout=out,stderr=subprocess.STDOUT)
 assert result.returncode==0
finally:
 f.write_text(original)
 (p/'M2-final-replay-timings.json').write_text(json.dumps(timings,indent=2)+'\n')
