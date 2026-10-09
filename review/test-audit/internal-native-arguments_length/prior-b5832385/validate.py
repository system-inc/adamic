import pathlib,subprocess,json,time
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-native-arguments_length';results=[]
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
for f in sorted(p.glob('*.diff')):
 if f.name=='switch.diff':continue
 patch=f.read_text();target=next(l[6:] for l in patch.splitlines() if l.startswith('+++ b/'));start=time.monotonic();a=subprocess.run(['git','apply','--check',str(f)],cwd=r,capture_output=True,text=True)
 if a.returncode:raise RuntimeError(a.stderr)
 subprocess.run(['git','apply',str(f)],cwd=r,check=True)
 cmd=['timeout','90','go','vet','./internal/native/'] if target.endswith('.go') else ['clang',*flags,'-I','internal/native/runtime','-c','internal/native/runtime/case.c','-o','/tmp/u045/validation.o']
 with (p/(f.stem+'-validation.log')).open('w') as log:v=subprocess.run(cmd,cwd=r,stdout=log,stderr=subprocess.STDOUT)
 subprocess.run(['git','restore','--source=HEAD','--',target],cwd=r,check=True)
 results.append(dict(id=f.stem,apply_check=a.returncode,compile_code=v.returncode,wall=time.monotonic()-start));(p/'final-validation.json').write_text(json.dumps(results,indent=2))
 if v.returncode:print('FAILED',f.stem,flush=True)
print('VALIDATED',len(results),flush=True)
