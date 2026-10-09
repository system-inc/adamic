from pathlib import Path
import subprocess,os,time,json
base=os.environ.copy();base['ADAMIC_YAML_LIBRARY']='/tmp/u153/library';results=[]
for ident,regex in [('W1','^TestScalarsMatchGoPlantedFailure$'),('W2','^TestSchemaMutants$'),('W3','^TestUnistMutants$'),('P1','^TestProduct_YamlScalarsLowered$'),('P2','^TestProduct_YamlScalarsNative$'),('P3','^TestProduct_YamlScalarsGo$'),('S1','^TestProduct_YamlScalarsLowered$'),('S2','^TestProduct_YamlScalarsNative$'),('S3','^TestProduct_YamlScalarsGo$')]:
 env=base.copy();env['ADAMIC_AUDIT_SELECTOR']=ident
 if ident.startswith('S'):env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u153/cache/'+ident
 path=Path('/tmp/u153/'+ident+'.log');start=time.monotonic()
 cmd=['timeout','95','go','test','-overlay=/tmp/u153/harness-overlay.json','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',regex]
 with path.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(id=ident,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd),selector=ident));Path('/tmp/u153/harness-results.json').write_text(json.dumps(results,indent=2));print(ident,r.returncode,results[-1]['wall'],flush=True)
