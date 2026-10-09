import pathlib,json,subprocess,os,time
p=pathlib.Path('/tmp/def-oct6');r=pathlib.Path('/workspace/adamic');relevant=json.loads((p/'review-callers.json').read_text())['agree'];import re
patterns=['^(TestFractionalPowersReachRuntime|TestReviewProgramsNoLooseFiles|TestReviewProgramsRefuse|TestReviewProgramsSelfTest)$','^TestReviewProgramsAgreeWithNode$/^('+'|'.join([re.escape('smoke.a')]+[re.escape(x['name']) for x in relevant])+')$'];(p/'new-test-patterns.json').write_text(json.dumps(patterns,indent=2));out=[]
for mutant in ['clean','D01','D02','D03','D04']:
 file=None;original=None
 try:
  if mutant!='clean':
   menu=json.loads((p/'menu.json').read_text());file=[x['file'] for x in menu if x['id']==mutant][0];original=(r/file).read_text();subprocess.run(['git','apply',str(p/(mutant+'.diff'))],cwd=r,check=True)
  row={'mutant':mutant,'runs':[],'failed':[],'passed':[],'skipped':[]}
  for i,pat in enumerate(patterns):
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pat];env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':f'/tmp/def-oct6/cache/{mutant}'};start=time.monotonic()
   with (p/f'{mutant}-new-{i}.log').open('w') as f:q=subprocess.run(cmd,cwd=r,stdout=f,stderr=subprocess.STDOUT,env=env)
   events=[]
   for s in (p/f'{mutant}-new-{i}.log').read_text().splitlines():
    try:events.append(json.loads(s))
    except:pass
   row['failed'] +=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/'not in e['Test']];row['passed'] +=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/'not in e['Test']];row['skipped'] +=[e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')]
   row['runs'].append({'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+f' > {mutant}-new-{i}.log 2>&1','exit':q.returncode,'wall_seconds':time.monotonic()-start})
   if mutant=='clean' and q.returncode:raise RuntimeError('red new-test baseline')
  out.append(row);(p/'new-test-results.json').write_text(json.dumps(out,indent=2));print(mutant,row['failed'],flush=True)
 finally:
  if original is not None:(r/file).write_text(original)
