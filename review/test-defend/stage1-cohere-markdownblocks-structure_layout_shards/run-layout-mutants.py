import pathlib,json,time,subprocess,os
root=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/defend-markdown');out=root/'review/test-defend/stage1-cohere-markdownblocks-structure_layout_shards';plans=json.loads((out/'planned-mutants.json').read_text());runs=[]
for m in plans[1:]:
 f=root/m['file'];original=f.read_text();env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/m['id']);env['ADAMIC_NATIVE_SPLIT']='1';env['ADAMIC_MARKDOWNWIDTH_DEPS']=str(p/'width-deps')
 try:
  f.write_text(original.replace(m['old'],m['new'],1))
  for family in ['Structure','Table']:
   regex='^TestMarkdown'+family+'Layout(Union|_[0-9]{3})$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',regex];start=time.monotonic()
   log=p/(m['id']+'-'+family+'.log')
   with log.open('w') as target:r=subprocess.run(cmd,env=env,stdout=target,stderr=subprocess.STDOUT)
   events=[]
   for line in log.read_text().splitlines():
    try:events.append(json.loads(line))
    except:pass
   cooked=any('test timed out' in e.get('Output','') for e in events)
   item=dict(mutant=m['id'],family=family,command=' '.join(cmd),cache=env['ADAMIC_BUILD_CACHE_DIR'],status=r.returncode,wall=time.monotonic()-start,cooked=cooked,failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test'in e]);runs.append(item);(out/'layout-runs.json').write_text(json.dumps(runs,indent=2));print(item,flush=True)
   if cooked:
    # The first cold run may leave completed immutable products; rerun this family bounded.
    start=time.monotonic()
    with (p/(m['id']+'-'+family+'-retry.log')).open('w') as target:rr=subprocess.run(cmd,env=env,stdout=target,stderr=subprocess.STDOUT)
    runs.append(dict(mutant=m['id'],family=family+' retry',command=' '.join(cmd),cache=env['ADAMIC_BUILD_CACHE_DIR'],status=rr.returncode,wall=time.monotonic()-start));(out/'layout-runs.json').write_text(json.dumps(runs,indent=2));print('retry',m['id'],family,rr.returncode,flush=True)
 finally:f.write_text(original)
(p/'layout-done').write_text('done')
