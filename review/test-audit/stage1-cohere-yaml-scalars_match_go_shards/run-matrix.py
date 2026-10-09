from pathlib import Path
import json,subprocess,os,time
root=Path('/workspace/adamic');evidence=root/'review/test-audit/stage1-cohere-yaml-scalars_match_go_shards';menu=json.loads((evidence/'menu.json').read_text());originals=json.loads(Path('/tmp/u153/originals.json').read_text());base=os.environ.copy();base['ADAMIC_YAML_LIBRARY']='/tmp/u153/library';results=[]
# Wait for clean timing runs to finish before changing any production source.
while not Path('/tmp/u153/timings-done').exists():time.sleep(1)
timing=json.loads(Path('/tmp/u153/timings.json').read_text())
if any(r['status'] for r in timing):print('STOP: red isolated baseline',flush=True);raise SystemExit(1)
regexes={'M1':'^(TestScalarsMatchGo(Union|_[0-9]{3})|TestUnistMatchesGo)$','M2':'^(TestSchemaMatchesGo|TestUnistMatchesGo)$','M3':'^TestUnistMatchesGo$','M4':'^(TestScalarsMatchGo(Union|_[0-9]{3})|TestSchemaMatchesGo|TestSpeedCostProbes|TestUnistMatchesGo|TestWidthsMatchGo)$'}
def run(ident,regex,cache):
 env=base.copy();env['ADAMIC_BUILD_CACHE_DIR']=cache
 cmd=['timeout','95','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',regex];start=time.monotonic()
 with Path('/tmp/u153/'+ident+'.log').open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(id=ident,regex=regex,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd),cache=cache));Path('/tmp/u153/matrix-runs.json').write_text(json.dumps(results,indent=2));print(ident,r.returncode,results[-1]['wall'],flush=True)
 return r.returncode
try:
 for m in menu:
  p=root/m['file'];p.write_text(originals[m['file']].replace(m['old'],m['new'],1));cache='/tmp/u153/cache/'+m['id'];status=run(m['id'],regexes[m['id']],cache)
  text=Path('/tmp/u153/'+m['id']+'.log').read_text()
  if status==124 or 'panic: test timed out' in text:
   for name in ['TestScalarsMatchGo(Union|_[0-9]{3})','TestSchemaMatchesGo','TestSpeedCostProbes','TestUnistMatchesGo','TestWidthsMatchGo']:
    if name.split('(')[0] in regexes[m['id']]:run(m['id']+'-'+name.split('(')[0],'^'+name+'$',cache)
  p.write_text(originals[m['file']])
 # Empty script entry produces no stdout. Keep only an argument read so native main exists.
 for ident,file,regex in [('P4','scalar_main.ts','^TestScalarsMatchGo(Union|_[0-9]{3})$'),('P5','schema_main.ts','^TestSchemaMatchesGo$'),('P6','unist_main.ts','^TestUnistMatchesGo$'),('P7','width_main.ts','^TestWidthsMatchGo$')]:
  p=root/'stage1/cohere/yaml'/file;original=p.read_text();empty="import { programArguments } from 'adamic';\nprogramArguments();\n"
  import difflib
  (evidence/(ident+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),empty.splitlines(True),fromfile='a/stage1/cohere/yaml/'+file,tofile='b/stage1/cohere/yaml/'+file)))
  try:p.write_text(empty);run(ident,regex,'/tmp/u153/cache/'+ident)
  finally:p.write_text(original)
finally:
 for file,s in originals.items():(root/file).write_text(s)
Path('/tmp/u153/matrix-done').write_text('done')
