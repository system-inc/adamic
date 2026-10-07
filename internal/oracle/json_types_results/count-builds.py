import os,subprocess,pathlib,json,time
root=pathlib.Path('/workspace/adamic-json-types-scratch');runner=root/'oracle/node.mjs';after=runner.read_bytes();before=pathlib.Path('/tmp/json-types-original-node.mjs').read_bytes();rows=[]
try:
 for variant in ['before','after']:
  runner.write_bytes(before if variant=='before' else after)
  invocations=pathlib.Path('/tmp/json-types-full-'+variant+'-go.jsonl');invocations.write_text('')
  env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',JSON_GO_LOG=str(invocations),PATH='/tmp/json-types-go-wrapper:'+os.environ['PATH'])
  cmd=['/tmp/json-types-'+variant+'.test','-test.count=1','-test.timeout=30m']
  start=time.monotonic()
  with open('/tmp/json-types-full-'+variant+'.log','w') as log: p=subprocess.run(cmd,cwd=root/'internal/oracle',env=env,stdout=log,stderr=subprocess.STDOUT)
  calls=[json.loads(line) for line in invocations.read_text().splitlines()]; builds=[c for c in calls if c and c[0] in ['run','build']]
  row={'variant':variant,'exit':p.returncode,'seconds':time.monotonic()-start,'goInvocations':calls,'buildRequests':len(builds),'command':'ADAMIC_GATE_UNCACHED=1 JSON_GO_LOG='+str(invocations)+' PATH=/tmp/json-types-go-wrapper:$PATH '+' '.join(cmd)};rows.append(row);print(json.dumps(row),flush=True);pathlib.Path('/tmp/json-types-full-count.json').write_text(json.dumps(rows,indent=2))
finally:runner.write_bytes(after)
