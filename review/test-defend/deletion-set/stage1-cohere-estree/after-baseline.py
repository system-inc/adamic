import json,pathlib,time,subprocess,datetime
p=pathlib.Path('review/test-defend/deletion-set/stage1-cohere-estree')
while True:
 lines=(p/'baseline.log').read_text().splitlines();events=[]
 for line in lines:
  try:events.append(json.loads(line))
  except:pass
 end=next((e for e in reversed(events) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None)
 if end:break
 time.sleep(3)
start=json.loads((p/'baseline-process.json').read_text())['start_epoch']; finish=datetime.datetime.fromisoformat(end['Time'].replace('Z','+00:00')).timestamp()
result=dict(status=end['Action'],wall_seconds=finish-start,binary_seconds=end.get('Elapsed'),rows_failed=sorted({e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']}),rows_skipped=sorted({e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test') and '/' not in e['Test']}))
(p/'baseline-result.json').write_text(json.dumps(result,indent=2)+'\n');print('baseline',result,flush=True)
if end['Action']!='pass':raise SystemExit('Red baseline, no replay permitted')
raise SystemExit(subprocess.call(['python3',str(p/'replay.py')]))
