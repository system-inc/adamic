import pathlib,subprocess,json,time,os
p=pathlib.Path('review/test-audit/internal-oracle-stage3_front');root=pathlib.Path.cwd();work=pathlib.Path('/tmp/u070/validation');work.mkdir(exist_ok=True);plan=json.loads((p/'plan.json').read_text());items=plan['mutants']+plan['probes'];basefiles={str(root/f):subprocess.check_output(['git','show','HEAD:'+f],text=True) for f in set(i['file'] for i in items)};replace={}
for i,(filename,text) in enumerate(basefiles.items()):
 if filename.endswith('.go'):
  target=work/('base-%d.go'%i);target.write_text(text);replace[filename]=str(target)
replace[str(root/'internal/lower/audit_mutant.go')]='';results=[];env=os.environ.copy();env['ADAMIC_MUTANT']=''
for item in items:
 mid=item['id'];filename=item['file'];dest=work/mid;dest.mkdir(exist_ok=True);target=dest/filename;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(basefiles[str(root/filename)]);t=time.monotonic();apply=['git','apply','--unsafe-paths','--directory='+str(dest),str((p/(mid+'.diff')).resolve())]
 with open(p/(mid+'-validation.log'),'w') as log:
  a=subprocess.run(apply,stdout=log,stderr=subprocess.STDOUT);cmd=[]
  if a.returncode:code=a.returncode
  elif filename.endswith('.go'):
   overlay=dict(replace);overlay[str(root/filename)]=str(target);manifest=work/(mid+'.json');manifest.write_text(json.dumps({'Replace':overlay}));cmd=['go','vet','-overlay='+str(manifest),'./'+str(pathlib.Path(filename).parent)+'/'];code=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  elif filename.endswith('.c'):
   flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'];cmd=['clang',*flags,'-I',str(root/'internal/native/runtime'),'-c',str(target),'-o',str(work/(mid+'.o'))];code=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  else:
   cmd=['go','run','./.audit-u070',str(target),str(work/(mid+'-port'))];code=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
 results.append(dict(id=mid,apply_command=apply,apply_exit=a.returncode,compile_command=cmd,compile_exit=code,wall_seconds=time.monotonic()-t));(p/'validation.json').write_text(json.dumps(results,indent=2));print(mid,code,flush=True)
