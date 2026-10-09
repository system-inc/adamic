import pathlib,subprocess,json,time,os
root=pathlib.Path(__file__).resolve().parent
source=pathlib.Path('stage1/cohere/markdownblocks/leaves.ts')
original=source.read_text()
results=[]
try:
 for m in json.loads((root/'plan.json').read_text()):
  assert original.count(m['from'])==1,m
  source.write_text(original.replace(m['from'],m['to'],1))
  (root/(m['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',str(source)]))
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-markdown-products/cache/'+m['mutant']
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^(TestProduct_MarkdownLeaf.*|TestMarkdownLeafComposition.*)$']
  start=time.monotonic()
  with (root/(m['mutant']+'.log')).open('w') as f: result=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
  events=[]
  for line in (root/(m['mutant']+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  results.append({**m,'command':cmd,'seconds':round(time.monotonic()-start,3),'exit':result.returncode,'rows_failed':sorted({e['Test'].split('/')[0] for e in events if e.get('Test') and e.get('Action')=='fail'}),'rows_passed':sorted({e['Test'] for e in events if e.get('Test') and '/' not in e['Test'] and e.get('Action')=='pass'}),'failing_lines':[e['Output'].strip() for e in events if 'differs:' in e.get('Output','') or 'disagreed' in e.get('Output','')]})
  (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
  source.write_text(original)
finally:source.write_text(original)
