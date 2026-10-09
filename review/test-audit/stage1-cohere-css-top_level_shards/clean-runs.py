import pathlib,re,json,subprocess,time,os
p=pathlib.Path('review/test-audit/stage1-cohere-css-top_level_shards');prefix='TestThePortParsesAsGoCohereDoes_';src=pathlib.Path('stage1/cohere/css/top_level_shards_test.go').read_text();allnames=re.findall(r'^func (TestThePortParsesAsGoCohereDoes_\w+)\(',src,re.M);listed=pathlib.Path('/tmp/u080-list.log').read_text().splitlines();assert len(allnames)==389 and all(n in listed for n in allnames)
# Fixed regular spacing, selected before production mutations are planned or run.
selected=[prefix+f'{i:03}' for i in range(0,388,12)];(p/'scope.json').write_text(json.dumps({'start_commit':'cf79ecec3723604428ab91ebcb283400d05a1548','setup':prefix+'Setup','family_members':[n for n in allnames if n!=prefix+'Setup'],'bounded_members':selected,'missing':[]},indent=2))
env=os.environ.copy();env.update(ADAMIC_CSS_FIXTURES='/tmp/u080-css-fixtures',ADAMIC_CSS_LIBRARY='/tmp/u080-css-library')
def run(id,pattern):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run',pattern];start=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 ev=[json.loads(l) for l in (p/(id+'.log')).read_text().splitlines() if l.startswith('{')];end=next((e for e in reversed(ev) if e.get('Action') in ['pass','fail'] and 'Test' not in e),{});(p/(id+'-run.json')).write_text(json.dumps({'command':'ADAMIC_CSS_FIXTURES=/tmp/u080-css-fixtures ADAMIC_CSS_LIBRARY=/tmp/u080-css-library '+' '.join(cmd),'wall':time.monotonic()-start,'exit':r.returncode,'seconds':end.get('Elapsed'),'timeout':any('test timed out' in e.get('Output','') for e in ev)},indent=2))
 return r.returncode
pattern='^('+'|'.join(selected+[prefix+'Setup'])+')$'
assert run('bounded-baseline',pattern)==0,'Red bounded baseline, stop audit'
for i in range(3):run('setup-time'+str(i+1),'^'+prefix+'Setup$')
for i in range(3):run('bounded-family-time'+str(i+1),'^('+'|'.join(selected)+')$')
for i in range(3):run('full-family-time'+str(i+1),'^'+prefix+'[0-9]{3}$')
