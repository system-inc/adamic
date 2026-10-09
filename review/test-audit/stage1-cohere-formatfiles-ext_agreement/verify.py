from pathlib import Path
import subprocess,json,time,shutil
p=Path('review/test-audit/stage1-cohere-formatfiles-ext_agreement');root=Path('/workspace/u090-tmp/standalone');root.mkdir(exist_ok=True)
file='stage1/cohere/formatfiles/golang.ts';base=subprocess.check_output(['git','show','HEAD:'+file],text=True);plan=json.loads((p/'plan.json').read_text());status=[]
for mid in ['M01','M02','M03','P01']:
 d=root/mid;port=d/'formatfiles';gitignore=d/'gitignore';port.mkdir(parents=True,exist_ok=True);gitignore.mkdir(exist_ok=True)
 for name in ['golang.ts','disk.ts','enumerate.ts','main.ts']:
  source=subprocess.check_output(['git','show','HEAD:stage1/cohere/formatfiles/'+name],text=True)
  if name=='golang.ts':
   if mid=='P01':
    start=source.index('export function ext(');end=source.index('\n// filepath.Join',start);source=source[:start]+"export function ext(path: string): string {\n\treturn '';\n}\n"+source[end:]
   else:
    m=next(m for m in plan if m['id']==mid);source=source.replace(m['before'],m['after'])
  (port/name).write_text(source)
 for name in ['path.ts','glob.ts','gitignore.ts']:(gitignore/name).write_text(subprocess.check_output(['git','show','HEAD:stage1/cohere/gitignore/'+name],text=True))
 with (p/(mid+'-compile.log')).open('w') as f:
  apply=subprocess.run(['git','apply','--check','--cached',str(p/(mid+'.diff'))],stdout=f,stderr=subprocess.STDOUT)
  cmd=['timeout','90','/workspace/u090-tmp/adamic','build',str(port/'main.ts'),'-o',str(d/'port'),'--sanitize'];s=time.monotonic();r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 status.append(dict(id=mid,apply_status=apply.returncode,compile_status=r.returncode,native_rebuild_seconds=time.monotonic()-s,command=cmd))
 (p/'compile-status.json').write_text(json.dumps(status,indent=2))
 if r.returncode or apply.returncode:raise SystemExit(mid+' verification failed')
