import pathlib,subprocess,os,json,time,difflib
repo=pathlib.Path('/workspace/adamic'); root=repo/'review/test-audit/stage1-cohere-static_single_assignment'; results=[]
def run(mid,file,change,pattern,vet=None):
 p=repo/file; original=p.read_text(); changed=change(original)
 (root/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 p.write_text(changed);env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u136/cache/'+mid;env['AUDIT_CLANG_LOG']=str(root/'logs'/(mid+'-clang.jsonl'));env['PATH']='/tmp/u136/bin:'+env['PATH']
 try:
  if vet:
   with (root/'logs'/(mid+'-vet.log')).open('w') as f:rc=subprocess.run(['go','vet',vet],stdout=f,stderr=subprocess.STDOUT,env=env).returncode
   assert rc==0
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/static_single_assignment/','-run',pattern]
  t=time.monotonic()
  with (root/'logs'/(mid+'.log')).open('w') as f:rc=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,env=env).returncode
  results.append(dict(id=mid,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command)+' > '+str(root/'logs'/(mid+'.log'))+' 2>&1',wall_seconds=time.monotonic()-t,returncode=rc))
  (root/'probe-witness-runs.json').write_text(json.dumps(results,indent=2)+'\n')
 finally:p.write_text(original)
run('P1','stage1/cohere/static_single_assignment/main.ts',lambda s:s[:s.index('const casesPath =')],'^TestThePortMakesWhatGoCohereMakes$')
def emptylower(s):
 a=s.index('func Lower(');b=s.index('\ntype lowering struct',a)
 return (s[:a]+'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn nil, nil\n}\n'+s[b:]).replace('\n\t"fmt"','').replace('\n\t"path/filepath"','')
run('P2','internal/lower/lower.go',emptylower,'^TestEachGapStandsWhereGapsMdSaysItDoes$','./internal/lower/')
def comparator(s):
 a=s.index('func firstDifference(');b=s.index('\n// lowered',a)
 return s[:a]+'func firstDifference(got string, want string) string {\n\treturn ""\n}\n'+s[b:]
run('W1','stage1/cohere/static_single_assignment/static_single_assignment_test.go',comparator,'.','./stage1/cohere/static_single_assignment/')
