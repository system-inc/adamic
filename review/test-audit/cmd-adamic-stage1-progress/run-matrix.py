from pathlib import Path
import subprocess,time,os,json,re,shlex
p=Path('review/test-audit/cmd-adamic-stage1-progress'); plan=json.loads((p/'plan.json').read_text()); rows=[s for s in (p/'list.log').read_text().splitlines() if s.startswith('Test')]; records=[]
for id in ['control']+[entry['id'] for entry in plan]:
 env=os.environ.copy();env['ADAMIC_MUTANT']='' if id=='control' else id
 command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-stage1-progress/','-run','.'];started=time.monotonic()
 with (p/(id+'.log')).open('w') as out:result=subprocess.run(command,env=env,stdout=out,stderr=subprocess.STDOUT)
 events=[]
 for line in (p/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 failed=[e['Test'] for e in events if e['Action']=='fail' and e.get('Test') in rows];passed=[e['Test'] for e in events if e['Action']=='pass' and e.get('Test') in rows]
 record=dict(id=id,command='ADAMIC_MUTANT='+shlex.quote(env['ADAMIC_MUTANT'])+' '+shlex.join(command)+' > '+str(p/(id+'.log'))+' 2>&1',exit=result.returncode,wall_seconds=time.monotonic()-started,failed=failed,passed=passed,unknown=[r for r in rows if r not in failed+passed],failure_lines=[e['Output'].strip() for e in events if e['Action']=='output' and e.get('Test') in failed and re.search(r'\w+_test.go:\d+:',e.get('Output',''))])
 records.append(record);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');print(id,result.returncode,failed,record['failure_lines'],flush=True)
 if record['unknown']:print('unknown rows require individual reruns',record['unknown'],flush=True)
 if id=='control' and result.returncode:raise RuntimeError('switch control red')
