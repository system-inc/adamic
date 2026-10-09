from pathlib import Path
import json,subprocess,time,os,difflib
root=Path.cwd();p=root/'review/test-audit/stage1-cohere-estree-recovery_lowered_recipe';plan=json.loads((p/'plan.json').read_text());original=json.loads((p/'originals.json').read_text());records=[]
e=os.environ.copy();e['ADAMIC_ESTREE_LIBRARY']='/tmp/u087-estree-library'
rows=[('RecoveredGrammar','^TestRecoveredGrammar$'),('ScalarEdges_family','^(TestScalarEdgesUnion|TestScalarEdges_000|TestScalarEdges_001|TestScalarEdges_002|TestScalarEdges_003)$')]
probe=dict(id='P01',file='stage1/cohere/estree/pipeline.ts',before='export function answer(path: string, text: string): string {',after="export function answer(path: string, text: string): string {\n    return '';",line=10)
text=original[probe['file']];changed=text.replace(probe['before'],probe['after']);(p/'diffs/P01.diff').write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='a/'+probe['file'],tofile='b/'+probe['file'])))
try:
 for m in plan+[probe]:
  for f,t in original.items():
   if (root/f).read_text()!=t:(root/f).write_text(t)
  f=root/m['file'];f.write_text(original[m['file']].replace(m['before'],m['after']))
  check=subprocess.run(['git','apply','--reverse','--check',str(p/'diffs'/(m['id']+'.diff'))],stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  e['ADAMIC_BUILD_CACHE_DIR']='/tmp/u087/cache/'+m['id']
  for name,regex in rows:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex];s=time.monotonic()
   with (p/(m['id']+'.'+name+'.log')).open('w') as log:r=subprocess.run(cmd,env=e,stdout=log,stderr=subprocess.STDOUT)
   records.append(dict(id=m['id'],row=name,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-s,apply_check_exit=check.returncode,cache=e['ADAMIC_BUILD_CACHE_DIR']))
   (p/'matrix-commands.json').write_text(json.dumps(records,indent=2)+'\n');print(m['id'],name,r.returncode,round(records[-1]['wall_seconds'],2),flush=True)
finally:
 for f,t in original.items():
  if (root/f).read_text()!=t:(root/f).write_text(t)
