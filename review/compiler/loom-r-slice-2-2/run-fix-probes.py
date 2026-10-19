import json,os,re,subprocess,time,sys
from pathlib import Path
root=Path.cwd();out=Path(__file__).resolve().parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4]);env=dict(os.environ,GOMAXPROCS='4')
results=[r for r in json.loads((out/'fix-mutants-initial.json').read_text()) if r['caught']] if '--resume' in sys.argv else []
def execute(name,pkg,pattern,path,before,after,marker,physical=False):
 source=Path(path).read_text();assert source.count(before)==1
 folder=out/'fix-mutants'/name;folder.mkdir(parents=True,exist_ok=True)
 planted=folder/('source.go.txt' if path.endswith('.go') else 'source.txt');planted.write_text(source.replace(before,after))
 command=['timeout','180','go','test',pkg,'-run',pattern,'-count=1','-v','-timeout=90s']
 if not physical:
  overlay=folder/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/path):str(planted)}}));command+=['-overlay='+str(overlay)]
 start=time.monotonic()
 try:
  if physical:Path(path).write_text(source.replace(before,after))
  with (folder/'test.log').open('w') as f:ran=subprocess.run(command,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=190)
 finally:
  if physical:Path(path).write_text(source)
 text=(folder/'test.log').read_text();caught=ran.returncode==1 and marker in text and '--- FAIL:' in text and 'build failed' not in text and 'test timed out' not in text
 row=dict(mutant=name,caught=caught,exit=ran.returncode,seconds=round(time.monotonic()-start,3),command=command,marker=marker);results.append(row)
 (out/'fix-mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True);assert caught
if '--resume' not in sys.argv: execute('colliding-flow-name','./cmd/adamic-gate','^TestEveryTestHasOneDescriptiveName$','internal/flow/callback_widening_test.go','func TestFlowCallbackWideningBooleanMap(','func TestCallbackWideningBooleanMap(','lives in two or more packages',True)
if '--resume' not in sys.argv: execute('unlisted-empty-message','./internal/native','^TestRuntimeStaticsAreListed$','internal/native/testdata/runtime-statics/storage.txt','`exceptions.c:empty_message:1`\n','','unlisted mutable storage empty_message',True)
execute('immortal-units-write','./internal/native','^TestErrorUndefinedMessageHasNoOwnProperty$','internal/native/runtime/string_index.c','if (adamic_reference_count(&string->heap) != 0) {','if (true) {','exit status 4')
# Compile each physically mutated header with the real branch compiler. Apply the README's
# exact error-header predicate to its fresh diagnostic; restore the program in all cases.
rows=json.loads((out/'acheck-final.json').read_text());headers=[]
for r in rows:
 if r['status'] not in ['refused','type error']:continue
 path=Path(r['path']);original=path.read_bytes();head,body=original.split(b'\n',1)
 for name,mutated in [('removed',body),('wrong',head+b' incorrect\n'+body)]:
  try:
   path.write_bytes(mutated)
   ran=subprocess.run(['timeout','30','/workspace/scratch/loom-r-slice-2-2-adamic','c',str(path)],capture_output=True,text=True,timeout=35)
  finally:path.write_bytes(original)
  m=re.search(r'Adamic 0\.1 refuses ([^\n]+?);',ran.stderr)
  expected='// a-check: refused '+m[1] if m else '// a-check: type error TS'+re.search(r'\bTS(\d+)\b',ran.stderr)[1]
  observed=mutated.decode().splitlines()[0];caught=ran.returncode==1 and observed!=expected
  row=dict(path=str(path),mutant=name,caught=caught,compiler_exit=ran.returncode,observed=observed,expected=expected,stderr=ran.stderr);headers.append(row);assert caught
(out/'header-mutants.json').write_text(json.dumps(headers,indent=2)+'\n');print('Live header probes caught:',len(headers),flush=True)
