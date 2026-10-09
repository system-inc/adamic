import pathlib,json,subprocess,time
root=pathlib.Path('/workspace/adamic');out=pathlib.Path('/tmp/def-graphql/evidence');plans=json.loads((out/'plan.json').read_text())
def events(path):
 es=[]
 for l in path.read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return es
expected={e['Test'] for e in events(out/'reached-baseline-parallel2.log') if e.get('Test') and '/' not in e['Test'] and e.get('Action')=='pass'}
results=[]
for p in plans:
 es=events(out/(p['mutant']+'-matrix.log'))
 if not any('panic: test timed out' in e.get('Output','') for e in es):continue
 observed={e['Test'] for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail']}
 unknown=expected-observed
 selector='^(TestPrinterWhitespaceGap_[0-9]+'+''.join('|'+x for x in sorted(unknown))+')$'
 path=root/p['file'];original=path.read_text()
 try:
  changed=original.replace(p['before'],p['after'])
  if 'extra_before' in p:changed=changed.replace(p['extra_before'],p['extra_after'])
  path.write_text(changed)
  cmd=f"source /workspace/adamic-tools/env.sh\nADAMIC_GRAPHQL_PRETTIER=/tmp/def-graphql/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/def-graphql/cache/{p['mutant']} timeout 120 go test -json -count=1 -timeout 90s -parallel 2 ./stage1/cohere/graphql/printer/ -run '{selector}'"
  (out/(p['mutant']+'-repair-command.txt')).write_text(cmd)
  start=time.monotonic()
  with (out/(p['mutant']+'-repair.log')).open('w') as log:code=subprocess.run(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  results.append(dict(mutant=p['mutant'],unknown_rows_rerun=sorted(unknown),selector=selector,exit=code,wall_seconds=time.monotonic()-start))
 finally:path.write_text(original)
(out/'repairs.json').write_text(json.dumps(results,indent=2))
