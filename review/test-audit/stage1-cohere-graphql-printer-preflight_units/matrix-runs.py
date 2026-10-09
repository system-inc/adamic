exec(open('/tmp/u099_switch.py').read())
import os,time,shlex
scope=json.loads((p/'scope.json').read_text());names=[n for g in scope['groups'] for n in g['members']];pattern='^('+'|'.join(names)+')$';env=os.environ.copy();env['ADAMIC_GRAPHQL_PRETTIER']='/tmp/u099-prettier'
def run(id):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run',pattern];start=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 events=[]
 for l in (p/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 end=next((e for e in reversed(events) if e.get('Action') in ['pass','fail'] and 'Test' not in e),{});meta={'command':'ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier '+shlex.join(cmd),'selector_file':'/tmp/u099-mutant','selector_value':pathlib.Path('/tmp/u099-mutant').read_text(),'exit':r.returncode,'wall':time.monotonic()-start,'seconds':end.get('Elapsed'),'failed':[e['Test'] for e in events if e.get('Action')=='fail' and 'Test'in e],'passed':[e['Test'] for e in events if e.get('Action')=='pass' and 'Test'in e],'timeout':any('test timed out' in e.get('Output','') for e in events)};(p/(id+'-run.json')).write_text(json.dumps(meta,indent=2));print(id,meta['seconds'],r.returncode,len(meta['failed']),flush=True);return r.returncode
try:
 assert run('switched-baseline')==0,'Red switched baseline'
 for id in ['M01','M02','M03','M04','P01']:
  pathlib.Path('/tmp/u099-mutant').write_text(id);run(id)
finally:
 pathlib.Path('/tmp/u099-mutant').write_text('')
 for f,text in base.items():pathlib.Path(f).write_text(text)
