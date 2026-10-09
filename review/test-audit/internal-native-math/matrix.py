import pathlib,os,subprocess,json,time,signal
out=pathlib.Path('review/test-audit/internal-native-math');plan=json.loads((out/'plan.json').read_text());rows=['TestMathAndToFixedMatchJavaScript','TestMaybeNumbersPackIntoOneDouble','TestNodeBufferRuntimeWithoutDeclarations','TestNormalizeRandomMatchesNode'];pattern='^('+'|'.join(rows)+')$';results=[]
for mid in ['control']+[x['id'] for x in plan['mutants']+plan['probes']]:
 env=os.environ.copy();env['ADAMIC_MUTANT']='' if mid=='control' else mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u049/cache/'+mid
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern];start=time.monotonic()
 with open(out/(mid+'.log'),'w') as log:
  p=subprocess.Popen(cmd,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try:code=p.wait(timeout=120)
  except subprocess.TimeoutExpired:code=124
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 events=[]
 for l in (out/(mid+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 failures=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')))
 results.append(dict(id=mid,command='ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),exit=code,wall=time.monotonic()-start,failures=failures,rows=rows))
 (out/'matrix.json').write_text(json.dumps(results,indent=2));print(mid,code,failures,round(results[-1]['wall'],3),flush=True)
 if mid=='control' and code:break
