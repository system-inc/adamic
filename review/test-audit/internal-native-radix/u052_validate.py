import pathlib,json,subprocess,tempfile,shutil,time
out=pathlib.Path('review/test-audit/internal-native-radix');root=pathlib.Path.cwd();rt=root/'internal/native/runtime'
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-DADAMIC_COUNT','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all']
start=time.monotonic();results=[]
for diff in sorted(out.glob('*.diff')):
 if diff.name=='switch.diff':continue
 with tempfile.TemporaryDirectory(prefix='u052-vet-') as tmp:
  tmp=pathlib.Path(tmp);shutil.copytree(rt,tmp/'internal/native/runtime')
  patch=subprocess.run(['git','apply',str(root/diff)],cwd=tmp,capture_output=True,text=True)
  meta=next((m for m in json.loads((out/'mutant-plan.json').read_text()) if m['id']==diff.stem),None)
  if meta:f=meta['file']
  else:f=next(line[6:] for line in diff.read_text().splitlines() if line.startswith('+++ b/'))
  if f.endswith('.h'):f='internal/native/runtime/string.c'
  cmd=['clang',*flags,'-I',str(tmp/'internal/native/runtime'),'-c',str(tmp/f),'-o',str(tmp/'out.o')]
  with (out/(diff.stem+'-compile.log')).open('w') as log:
   log.write(' '.join(cmd)+'\n'+patch.stderr);r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  results.append({'id':diff.stem,'apply_exit':patch.returncode,'compile_exit':r.returncode})
(out/'standalone-validation.json').write_text(json.dumps({'wall':time.monotonic()-start,'results':results},indent=2))
