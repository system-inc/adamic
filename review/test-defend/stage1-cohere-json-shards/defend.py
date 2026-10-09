import json, subprocess, time, os, difflib, pathlib
p=pathlib.Path(__file__).parent
base='92d196011e78208b45b3768dcf2c8c88e7ec132d'
menu=json.loads((p/'menu.txt').read_text());selection=json.loads((p/'matrix-selection.txt').read_text())
results=[]
def run(label,cmd,env):
 t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 result={'label':label,'command':cmd,'exit':r.returncode,'wall':time.monotonic()-t,'env':{k:env[k] for k in ['ADAMIC_BUILD_CACHE_DIR','ADAMIC_JSON_BENCH','NODE_V8_COVERAGE'] if k in env}}
 with (p/'commands.txt').open('a') as f:f.write(json.dumps(result)+'\n')
 return result
scratch=pathlib.Path('/tmp/json-defend');scratch.mkdir(exist_ok=True)
cases=scratch/'cost-cases.txt';cases.write_text('probe.json\t[0,-0,1,-1,1.50,1E+03,1e-003,1e309,1e-400,9007199254740993]\n')
node=['timeout','10','node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/cohere/json/main.ts','--cases',str(cases)]
env=os.environ.copy();env['NODE_V8_COVERAGE']='/tmp/json-defend/cost/baseline';baseline=run('cost-baseline',node,env);assert baseline['exit']==0
for m in menu:
 file=pathlib.Path(m['file']);original=subprocess.check_output(['git','show',base+':'+m['file']],text=True)
 changed=original.replace(m['old'],m['new'],1)
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (p/(m['id']+'.diff')).write_text(diff);subprocess.run(['git','apply','--check',str(p/(m['id']+'.diff'))],check=True);subprocess.run(['git','apply',str(p/(m['id']+'.diff'))],check=True)
 try:
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/json-defend/cache/'+m['id'];env['ADAMIC_JSON_BENCH']='1';env['ADAMIC_JSON_PRETTIER']='/tmp/json-defend/prettier'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',selection['selector']]
  matrix=run(m['id'],cmd,env)
  env['NODE_V8_COVERAGE']='/tmp/json-defend/cost/'+m['id'];witness=run('cost-'+m['id'],node,env)
  events=[]
  for s in (p/(m['id']+'.log')).read_text().splitlines():
   try:events.append(json.loads(s))
   except ValueError:pass
  record=dict(m,matrix=matrix,witness=witness,rows_failed=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],rows_passed=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test')],rows_skipped=[e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')],failing_output=[e['Output'].strip() for e in events if '.go:' in e.get('Output','') and any(e.get('Test','')==x.get('Test') for x in events if x.get('Action')=='fail')],witness_same=(p/'cost-baseline.log').read_bytes()==(p/('cost-'+m['id']+'.log')).read_bytes())
  results.append(record);(p/'results.txt').write_text(json.dumps(results,indent=2)+'\n');print(m['id'],matrix['exit'],record['rows_failed'],flush=True)
 finally:subprocess.run(['git','apply','--reverse',str(p/(m['id']+'.diff'))],check=True)
 if matrix['exit'] not in [0,1]:break
