import pathlib,subprocess,os,json,hashlib
root=pathlib.Path('/workspace/adamic-json-types-scratch');source=root/'oracle/json_types.go';original=source.read_text();phase=pathlib.Path('/tmp/json-types-proof-phase');proof=pathlib.Path('/tmp/json-types-proof-go');proof.mkdir(exist_ok=True)
real='/workspace/adamic-tools/go/bin/go';wrapper=proof/'go'
wrapper.write_text('''#!/usr/bin/env python3
import os,sys,json,subprocess,pathlib,hashlib
p=subprocess.run(['''+repr(real)+''']+sys.argv[1:])
if p.returncode==0 and '-o' in sys.argv:
 binary=pathlib.Path(sys.argv[sys.argv.index('-o')+1]);label=pathlib.Path('/tmp/json-types-proof-phase').read_text();record={'phase':label,'arguments':sys.argv[1:],'sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'mode':oct(binary.stat().st_mode&0o777)}
 with open('/tmp/json-types-source-builds.jsonl','a') as log:log.write(json.dumps(record)+'\\n')
sys.exit(p.returncode)
''');wrapper.chmod(0o755)
pathlib.Path('/tmp/json-types-source-builds.jsonl').write_text('');env=dict(os.environ,PATH=str(proof)+':'+os.environ['PATH']);env.pop('ADAMIC_GATE_UNCACHED',None)
cmd=['/tmp/json-types-after.test','-test.run','^TestNativeAgreesWithNode/internal/oracle/testdata/json_decode_scalars.a$','-test.count=1','-test.timeout=30m','-test.v'];rows=[]
try:
 for label in ['before','mutant','restored','repeated']:
  text=original
  if label=='mutant':text=text.replace('if err = json.NewEncoder(os.Stdout).Encode(sources); err != nil {','for path, text := range sources { sources[path] = "console.log(\\"json_types source mutant\\");\\n" + text }\n\tif err = json.NewEncoder(os.Stdout).Encode(sources); err != nil {')
  source.write_text(text);phase.write_text(label)
  with open('/tmp/json-types-source-'+label+'.log','w') as stream:p=subprocess.run(cmd,cwd=root/'internal/oracle',env=env,stdout=stream,stderr=subprocess.STDOUT)
  rows.append({'phase':label,'exit':p.returncode});print(label,p.returncode,flush=True)
finally:source.write_text(original)
builds=[json.loads(s) for s in pathlib.Path('/tmp/json-types-source-builds.jsonl').read_text().splitlines()];assert [r['exit'] for r in rows]==[0,1,0,0],rows
assert builds[0]['sha256']!=builds[1]['sha256'];assert builds[0]['sha256']==builds[2]['sha256']==builds[3]['sha256']
pathlib.Path('/tmp/json-types-source-proof.json').write_text(json.dumps({'runs':rows,'builds':builds},indent=2))
