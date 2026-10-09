import pathlib,subprocess,json,time
out=pathlib.Path('review/test-audit/internal-native-math');work=pathlib.Path('/tmp/u049/validation');work.mkdir(exist_ok=True);results=[]
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
for diff in sorted(out.glob('*.diff')):
 start=time.monotonic();text=diff.read_text();file=text.splitlines()[0][6:];original=subprocess.check_output(['git','show','HEAD:'+file],text=True)
 # Apply each diff to an isolated exact-origin source file with git apply.
 target=work/file;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(original)
 cmd=['git','apply','--unsafe-paths','--directory='+str(work),str(diff.resolve())]
 with open(out/(diff.stem+'-compile.log'),'w') as log:
  apply=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  cc=['clang',*flags,'-I',str(pathlib.Path('internal/native/runtime').resolve()),'-c',str(target),'-o',str(work/(diff.stem+'.o'))]
  code=subprocess.run(cc,stdout=log,stderr=subprocess.STDOUT).returncode if apply.returncode==0 else apply.returncode
 results.append(dict(id=diff.stem,apply_command=cmd,compile_command=cc,exit=code,wall=time.monotonic()-start))
 (out/'validation.json').write_text(json.dumps(results,indent=2));print(diff.stem,code,flush=True)
