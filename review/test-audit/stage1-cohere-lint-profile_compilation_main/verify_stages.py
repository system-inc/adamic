from pathlib import Path
import subprocess,time,json,concurrent.futures,os
p=Path('review/test-audit/stage1-cohere-lint-profile_compilation_main');d=Path('/workspace/u111-verify');flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2'];runtime=Path('internal/native/runtime').resolve();results=[]
def verify(mid):
 q=d/mid;q.mkdir(exist_ok=True)
 if not (q/'main.ts').exists():
  import re
  root=Path.cwd()
  for name in ['main.ts','lint.ts']:
   f='stage1/cohere/lint/'+name;target=q/f;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(subprocess.check_output(['git','show','HEAD:'+f],text=True))
  z=subprocess.run(['git','apply','--unsafe-paths','--directory='+str(q),str((p/(mid+'.diff')).resolve())],capture_output=True,text=True);assert z.returncode==0,z.stderr
  for name in ['main.ts','lint.ts']:
   source=q/'stage1/cohere/lint'/name
   def rewrite(m):
    spec=m.group(2)
    if not spec.startswith('.'):return m.group(0)
    absolute=((root/'stage1/cohere/lint'/name).parent/spec).resolve()
    if absolute==root/'stage1/cohere/lint/lint.ts':absolute=q/'lint.ts'
    return m.group(1)+str(absolute)+m.group(3)
   (q/name).write_text(re.sub(r"(from\s+['\"])([^'\"]+)(['\"])",rewrite,source.read_text()))
 env=dict(os.environ);env['ADAMIC_BUILD_CACHE_DIR']=str(q/'cache')
 cmd=['timeout','90',str(d/'adamic'),'c',str(q/'main.ts')];st=time.monotonic()
 with (q/'main.c').open('w') as out,(p/(mid+'-emit-verify.log')).open('w') as log:r=subprocess.run(cmd,stdout=out,stderr=log,env=env)
 x=dict(id=mid,emit_command=cmd,emit_status=r.returncode,emit_wall=time.monotonic()-st)
 if r.returncode==0:
  cmd=['timeout','90','clang']+flags+['-I',str(runtime),'-o',str(q/'scanner'),str(q/'main.c')]+[str(f) for f in sorted(runtime.glob('*.c'))]+['-lm'];st=time.monotonic()
  with (p/(mid+'-clang-verify.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  x.update(clang_command=cmd,clang_status=r.returncode,clang_wall=time.monotonic()-st,binary_exists=(q/'scanner').exists())
 z=subprocess.run(['git','apply','--check','--cached',str(p/(mid+'.diff'))],capture_output=True,text=True);x['apply_status']=z.returncode
 return x
with concurrent.futures.ThreadPoolExecutor(2) as pool:
 futures=[pool.submit(verify,mid) for mid in ['M01','M02','M03','P01']]
 for future in concurrent.futures.as_completed(futures):
  results.append(future.result());(p/'native-staged-verification.json').write_text(json.dumps(results,indent=2))
