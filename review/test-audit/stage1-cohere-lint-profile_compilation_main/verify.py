from pathlib import Path
import subprocess,time,json,re,os
root=Path.cwd();p=root/'review/test-audit/stage1-cohere-lint-profile_compilation_main';d=Path('/workspace/u111-verify');d.mkdir(exist_ok=True);base=lambda f:subprocess.check_output(['git','show','HEAD:'+f],text=True)
cmd=['go','build','-o',str(d/'adamic'),'./cmd/adamic'];st=time.monotonic()
with (p/'compiler-cli-build.log').open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
assert r.returncode==0;(p/'compiler-cli-build.json').write_text(json.dumps({'command':cmd,'wall':time.monotonic()-st,'status':r.returncode}))
res=[]
for mid in ['M01','M02','M03','P01']:
 q=d/mid;q.mkdir(exist_ok=True)
 for name in ['main.ts','lint.ts']:
  f='stage1/cohere/lint/'+name;target=q/f;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(base(f))
 cmd=['git','apply','--unsafe-paths','--directory='+str(q),str(p/(mid+'.diff'))];r=subprocess.run(cmd,capture_output=True,text=True);assert r.returncode==0,r.stderr
 for name in ['main.ts','lint.ts']:
  source=q/'stage1/cohere/lint'/name;text=source.read_text()
  def rewrite(m):
   spec=m.group(2)
   if not spec.startswith('.'):return m.group(0)
   origin=(root/'stage1/cohere/lint'/name).parent/spec
   absolute=origin.resolve()
   if absolute==root/'stage1/cohere/lint/lint.ts':absolute=q/'lint.ts'
   return m.group(1)+str(absolute)+m.group(3)
  text=re.sub(r"(from\s+['\"])([^'\"]+)(['\"])",rewrite,text);(q/name).write_text(text)
 env=dict(os.environ);env['ADAMIC_BUILD_CACHE_DIR']=str(q/'cache')
 cmd=['timeout','90',str(d/'adamic'),'build',str(q/'main.ts'),'-o',str(q/'scanner')];st=time.monotonic()
 with (p/(mid+'-native-verify.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 x=dict(id=mid,command=cmd,status=r.returncode,wall=time.monotonic()-st,binary_exists=(q/'scanner').exists(),env={'ADAMIC_BUILD_CACHE_DIR':str(q/'cache')},copy='only imports rewritten to the identical absolute source graph; mutated main/lint copied from applied standalone diff')
 z=subprocess.run(['git','apply','--check','--cached',str(p/(mid+'.diff'))],capture_output=True,text=True);x['apply_status']=z.returncode;x['apply_error']=z.stderr
 res.append(x);(p/'native-verification.json').write_text(json.dumps(res,indent=2))
 if r.returncode:break
