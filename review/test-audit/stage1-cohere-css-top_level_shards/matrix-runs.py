exec(open('/tmp/u080_mutate.py').read())
scope=json.loads((p/'scope.json').read_text());prefix='TestThePortParsesAsGoCohereDoes_';pattern='^('+'|'.join(scope['bounded_members']+[scope['setup']])+')$';env=os.environ.copy();env.update(ADAMIC_CSS_FIXTURES='/tmp/u080-css-fixtures',ADAMIC_CSS_LIBRARY='/tmp/u080-css-library')
def run(id,pattern=pattern):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run',pattern];start=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 ev=[]
 for l in (p/(id+'.log')).read_text().splitlines():
  try:ev.append(json.loads(l))
  except:pass
 end=next((e for e in reversed(ev) if e.get('Action') in ['pass','fail'] and 'Test' not in e),{});failed=[e['Test'] for e in ev if e.get('Action')=='fail' and 'Test' in e];passed=[e['Test'] for e in ev if e.get('Action')=='pass' and 'Test' in e]
 (p/(id+'-run.json')).write_text(json.dumps({'command':'selector='+id+'; ADAMIC_CSS_FIXTURES=/tmp/u080-css-fixtures ADAMIC_CSS_LIBRARY=/tmp/u080-css-library '+' '.join(cmd),'wall':time.monotonic()-start,'exit':r.returncode,'seconds':end.get('Elapsed'),'failed':failed,'passed':passed,'timeout':any('test timed out' in e.get('Output','') for e in ev)},indent=2));return r.returncode
assert run('switched-baseline')==0,'Red switched baseline, stop'
for id in ['M01','M02','M03','M04','P01']:
 pathlib.Path('/tmp/u080-mutant').write_text(id);run(id)
pathlib.Path('/tmp/u080-mutant').write_text('')
for f,s in base.items():pathlib.Path(f).write_text(s)
