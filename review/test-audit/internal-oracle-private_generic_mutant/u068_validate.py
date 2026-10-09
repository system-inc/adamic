import pathlib,json,subprocess,tempfile,shutil,difflib,time
root=pathlib.Path.cwd();p=root/'review/test-audit/internal-oracle-private_generic_mutant';start=time.monotonic()
# Remove imports made unused by the Lower empty-answer probe.
diff=p/'P01.diff';original=subprocess.check_output(['git','show','HEAD:internal/lower/lower.go'],text=True)
with tempfile.TemporaryDirectory(prefix='u068-p01-') as tmp:
 tmp=pathlib.Path(tmp);(tmp/'internal/lower').mkdir(parents=True);(tmp/'internal/lower/lower.go').write_text(original);subprocess.run(['git','apply',str(diff)],cwd=tmp,check=True);changed=(tmp/'internal/lower/lower.go').read_text().replace('\n\t"fmt"','').replace('\n\t"path/filepath"','');diff.write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/internal/lower/lower.go',tofile='b/internal/lower/lower.go')))
results=[]
for id in ['M01','M02','M03','M04','P01','P02','P03','P04']:
 diff=p/(id+'.diff');file=next(l[6:] for l in diff.read_text().splitlines() if l.startswith('+++ b/'))
 with tempfile.TemporaryDirectory(prefix='u068-validate-') as tmp:
  tmp=pathlib.Path(tmp);target=tmp/file;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(subprocess.check_output(['git','show','HEAD:'+file],text=True));apply=subprocess.run(['git','apply',str(diff)],cwd=tmp,capture_output=True,text=True);assert apply.returncode==0,apply.stderr
  if file.endswith('.go'):
   replacements={}
   for f in ['internal/lower/lower.go','internal/native/emit.go','internal/native/emit_expressions.go','internal/javascript/javascript.go']:
    baseline=tmp/f;baseline.parent.mkdir(parents=True,exist_ok=True)
    if f!=file:baseline.write_text(subprocess.check_output(['git','show','HEAD:'+f],text=True))
    replacements[str(root/f)]=str(baseline)
   overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}));cmd=['go','vet','-overlay',str(overlay),'./'+str(pathlib.Path(file).parent)+'/']
  else:
   rt=tmp/'internal/native/runtime';modified=target.read_text();shutil.copytree(root/'internal/native/runtime',rt,dirs_exist_ok=True);target.write_text(modified);(rt/'string.c').write_text(subprocess.check_output(['git','show','HEAD:internal/native/runtime/string.c'],text=True));cmd=['clang','-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-DADAMIC_COUNT','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I',str(rt),'-c',str(rt/'string.c'),'-o',str(tmp/'out.o')]
  with (p/(id+'-validate.log')).open('w') as log:
   log.write(' '.join(cmd)+'\n');r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  results.append({'id':id,'apply_exit':apply.returncode,'compile_exit':r.returncode,'command':' '.join(cmd)})
(p/'standalone-validation.json').write_text(json.dumps({'wall':time.monotonic()-start,'results':results},indent=2))
