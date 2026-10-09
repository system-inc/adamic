import pathlib,subprocess,json,os,time,gzip,shutil,re
root=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/flow-defense/evidence')
results=json.loads((p/'results.json').read_text());base=subprocess.check_output(['git','show','HEAD:internal/flow/liveness.go'],cwd=root).decode()
target='TestFlowLiveness____oracle_testdata_timsort_a_63f1305dfb95'
rows=[target,'TestDebuggerHasNoFlowInstruction','TestFlowProgram_testdata_joins_a_091c4a59e83f','TestFlowProgram_testdata_mutations_a_fe30ed94ac7e','TestFlowProgram____oracle_testdata_node_buffer_digest_twice_a_eee6c7ec2841','TestFlowProgram____oracle_testdata_pow_fractional_a_633be2799c28']
try:
 for result in results:
  if not result.get('timeout'):continue
  i=result['mutant'];subprocess.run(['git','apply',str(p/(i+'.diff'))],cwd=root,check=True)
  env=dict(os.environ,TMPDIR='/tmp/flow-defense/tmp',ADAMIC_BUILD_CACHE_DIR='/tmp/flow-defense/cache/'+i+'-bounded')
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/flow/','-run','^('+'|'.join(rows)+')$'];start=time.monotonic()
  with (p/(i+'-bounded.log')).open('w') as out:r=subprocess.run(command,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  result['bounded']={'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command),'rows':rows,'status':r.returncode,'wall':round(time.monotonic()-start,3),'rows_failed':[],'rows_passed':[],'first_assertions':{}}
  with (p/(i+'-bounded.log')).open() as inp:
   for line in inp:
    try:x=json.loads(line)
    except:continue
    t=x.get('Test','');output=x.get('Output','')
    if t and '/' not in t and x.get('Action') in ['pass','fail']:result['bounded']['rows_'+('passed' if x['Action']=='pass' else 'failed')].append(t)
    if t and re.search(r'\w+_test.go:\d+:',output) and t not in result['bounded']['first_assertions']:result['bounded']['first_assertions'][t]=output.strip()
  root.joinpath('internal/flow/liveness.go').write_text(base)
  (p/'results.json').write_text(json.dumps(results,indent=2));print(i,'bounded',result['bounded']['wall'],result['bounded']['rows_failed'],flush=True)
  if (p/(i+'-bounded.log')).stat().st_size>20_000_000:
   with (p/(i+'-bounded.log')).open('rb') as inp,gzip.open(p/(i+'-bounded.log.gz'),'wb',compresslevel=1) as out:shutil.copyfileobj(inp,out)
   (p/(i+'-bounded.log')).unlink()
finally:root.joinpath('internal/flow/liveness.go').write_text(base)
