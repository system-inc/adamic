import json,os,pathlib,subprocess,time
p=pathlib.Path('review/test-defend/stage1-cohere-yaml-compose')
plan=json.loads((p/'plan.json').read_text())
functional=['TestComposeMatchGo','TestCSTMatchesGo','TestPropsMatchGo','TestScalarsMatchGo_[0-9]+','TestUnistMatchesGo','TestFormatterMatchesGo','TestFileDriver_[0-9]+','TestFileDriverUnion']
runs=[]
for m in plan:
 path=pathlib.Path(m['file']); original=path.read_text(); changed=original.replace(m['old'],m['new'],1)
 path.write_text(changed)
 diff=subprocess.run(['git','diff','--',str(path)],text=True,capture_output=True,check=True).stdout
 (p/(m['id']+'.diff')).write_text(diff)
 path.write_text(original)
 subprocess.run(['git','apply','--check',str(p/(m['id']+'.diff'))],check=True)
 path.write_text(changed)
 try:
  for row in functional:
   # Shared parser reaches every functional YAML layer. Main only reaches formatting/file mode.
   if m['id']=='D1' and row not in ['TestComposeMatchGo','TestUnistMatchesGo','TestFormatterMatchesGo','TestFileDriver_[0-9]+','TestFileDriverUnion']:continue
   if m['id']=='D3' and row not in ['TestFormatterMatchesGo','TestFileDriver_[0-9]+','TestFileDriverUnion']:continue
   env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']='/tmp/defend-yaml/library';env['ADAMIC_NATIVE_SPLIT']='1';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-yaml/cache/'+m['id']
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+row+')$']
   log=p/(m['id']+'-'+row.replace('[0-9]+','family')+'.log'); start=time.monotonic()
   with log.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
   records=[]
   for line in log.read_text().splitlines():
    try:records.append(json.loads(line))
    except ValueError:pass
   cooked=r.returncode==124 or any('test timed out' in x.get('Output','') for x in records)
   item=dict(mutant=m['id'],row=row,command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],log=str(log),exit=r.returncode,wall=time.monotonic()-start,cooked=cooked,failed=[x['Test'] for x in records if x.get('Action')=='fail' and x.get('Test') and '/' not in x['Test']],passed=[x['Test'] for x in records if x.get('Action')=='pass' and x.get('Test') and '/' not in x['Test']],assertions=[x.get('Output','').strip() for x in records if ': byte ' in x.get('Output','') or 'native control disagrees' in x.get('Output','')])
   if cooked:item['failed']=[]
   runs.append(item);(p/'matrix.json').write_text(json.dumps(runs,indent=2)+'\n')
   print(m['id'],row,r.returncode,round(item['wall'],1),'cooked' if cooked else '',flush=True)
 finally:path.write_text(original)
