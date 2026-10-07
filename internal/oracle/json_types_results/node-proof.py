import os,subprocess,pathlib,json,re,concurrent.futures
root=pathlib.Path('/workspace/adamic-json-types-scratch');runner=root/'oracle/node.mjs';after=runner.read_bytes();before=pathlib.Path('/tmp/json-types-original-node.mjs').read_bytes()
fixtures=[]
for name in ['library_json_decode_test.go','library_json_encode_test.go']:
 text=(root/'internal/oracle'/name).read_text().split('func init()',1)[0]
 fixtures.extend(re.findall(r'"(json_[^"]+\.a)"',text))
records={};env=dict(os.environ,JSON_GO_LOG='/tmp/json-types-final-byte-go.log',PATH='/tmp/json-types-go-wrapper:'+os.environ['PATH']);pathlib.Path(env['JSON_GO_LOG']).write_text('')
def run(name):
 cmd=['node','--disable-warning=ExperimentalWarning',str(runner),str(root/'internal/oracle/testdata'/name)]
 p=subprocess.run(cmd,cwd=root,env=env,capture_output=True,timeout=180)
 return name,{'exit':p.returncode,'stdout':p.stdout.hex(),'stderr':p.stderr.hex()}
try:
 for variant in ['before','after']:
  runner.write_bytes(before if variant=='before' else after)
  if variant=='after':
   binary='/tmp/json-types-final-byte-helper';subprocess.run(['go','build','-o',binary,'./oracle/json_types.go'],cwd=root,env=env,check=True,stdout=open('/tmp/json-types-final-byte-build.log','w'),stderr=subprocess.STDOUT);env['ADAMIC_ORACLE_JSON_TYPES']=binary
  with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:records[variant]=dict(pool.map(run,fixtures))
finally:runner.write_bytes(after)
pathlib.Path('/tmp/json-types-final-byte-results.json').write_text(json.dumps(records,indent=2))
assert records['before']==records['after'],'output mismatch'
assert all(v['exit']==0 for v in records['after'].values()),'source fixture failed'
print('byte-identical stdout/stderr/exit:',len(fixtures),'fixtures')
