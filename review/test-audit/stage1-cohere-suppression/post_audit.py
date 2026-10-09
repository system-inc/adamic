import pathlib,json,subprocess,time,os,re
out=pathlib.Path('review/test-audit/stage1-cohere-suppression');limit=time.monotonic()+120
while not (out/'runs.json').exists() or not any(r['name']=='final-clean' for r in json.loads((out/'runs.json').read_text())):
 if time.monotonic()>limit:raise SystemExit('main runner still active, narrowing required')
 time.sleep(1)
plan=json.loads((out/'frozen-plan.json').read_text());rec=[]
def run(name,cmd,env=None):
 start=time.monotonic()
 with (out/(name+'.log')).open('w')as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 d={'name':name,'command':' '.join(cmd),'wall':time.monotonic()-start,'exit':r.returncode};rec.append(d);(out/'post-runs.json').write_text(json.dumps(rec,indent=2));print(d,flush=True);return r.returncode
for m in plan['mutants'][:3]:
 f=pathlib.Path(m['file']);original=f.read_text();assert original.count(m['from'])==1
 try:
  f.write_text(original.replace(m['from'],m['to']));env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u138/cache/'+m['id']);run(m['id']+'-build-validated',['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/suppression/main.ts','-o','/tmp/u138/'+m['id'],'--sanitize'],env)
 finally:f.write_text(original)
for name,pattern in [('gap','^TestEachGapStandsWhereGapsMdSaysItDoes$'),('port','^TestThePortParsesAsGoCohereDoes$/^natively')]:
 env=dict(os.environ)
 if name=='port':env['NODE_V8_COVERAGE']='/tmp/u138/node-coverage'
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native','-coverprofile='+str(out/(name+'.cover')),'./stage1/cohere/suppression/','-run',pattern];run(name+'-coverage',cmd,env)
 if (out/(name+'.cover')).exists():
  with (out/(name+'-functions-covered.txt')).open('w')as f:subprocess.run(['go','tool','cover','-func='+str(out/(name+'.cover'))],stdout=f,stderr=subprocess.STDOUT)
# V8 names and counters prove which port functions Node actually executed.
covered=[]
for file in pathlib.Path('/tmp/u138/node-coverage').glob('*.json'):
 data=json.loads(file.read_text())
 for script in data.get('result',[]):
  if not any(script.get('url','').endswith('/'+n) for n in ['directives.ts','scan.ts','suppression.ts','main.ts']):continue
  for fn in script.get('functions',[]):
   if any(r.get('count',0)>0 for r in fn.get('ranges',[])):covered.append({'url':script['url'],'function':fn['functionName'],'ranges':fn['ranges']})
(out/'port-functions-v8.json').write_text(json.dumps(covered,indent=2)+'\n')
