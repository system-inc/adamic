import pathlib,subprocess,json,time,os
root=pathlib.Path('.').resolve();p=root/'review/test-audit/stage1-cohere-formatfiles-formatfiles_products';plan=json.loads((p/'plan.json').read_text());rs=[];origin=root/'/tmp/u091/validation/origin';origin=pathlib.Path('/tmp/u091/validation/origin');origin.mkdir(parents=True,exist_ok=True)
go_files=['stage1/cohere/formatfiles/formatfiles_prepare_test.go','stage1/cohere/formatfiles/formatfiles_shards_test.go']
for f in go_files:
 target=origin/f;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(subprocess.check_output(['git','show','HEAD:'+f]))
for m in plan:
 mid=m['id'];work=pathlib.Path('/tmp/u091/validation')/mid;work.mkdir(parents=True,exist_ok=True);f=m['file'];target=work/f;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(subprocess.check_output(['git','show','HEAD:'+f]));t=time.monotonic()
 with open(p/(mid+'-validation.log'),'w') as log:
  apply=['git','apply','--unsafe-paths','--directory='+str(work),str(p/(mid+'.diff'))];a=subprocess.run(apply,stdout=log,stderr=subprocess.STDOUT).returncode
  if f.endswith('.ts'):
   for rel in ['stage1/cohere/formatfiles/'+n for n in ['main.ts','enumerate.ts','disk.ts','golang.ts']]+['stage1/cohere/gitignore/'+n for n in ['path.ts','glob.ts','gitignore.ts']]:
    tpath=work/rel;tpath.parent.mkdir(parents=True,exist_ok=True)
    if rel!=f:tpath.write_bytes(subprocess.check_output(['git','show','HEAD:'+rel]))
   cmd=['go','run','./.audit-u091',str(work/'stage1/cohere/formatfiles/main.ts'),str(work/'port')];env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u091/cache/validation/'+mid;v=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  else:
   replace={str(root/g):str(work/g if g==f else origin/g) for g in go_files};overlay=work/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}));cmd=['go','vet','-overlay='+str(overlay),'./stage1/cohere/formatfiles/'];v=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT).returncode
 rs.append(dict(id=mid,apply_command=apply,apply_exit=a,compile_command=cmd,compile_exit=v,wall_seconds=time.monotonic()-t));(p/'validation.json').write_text(json.dumps(rs,indent=2));print(mid,a,v,round(rs[-1]['wall_seconds'],3),flush=True)
