from pathlib import Path
import subprocess,os,json,time,shutil
root=Path('/workspace/adamic');E=root/'review/test-audit/stage1-cohere-estree-syntax';menu=json.loads((E/'menu.json').read_text());results={}
for m in menu+[dict(id='P1')]:
 mid=m['id'];scratch=Path('/tmp/u089/replay')/mid;source=scratch/'stage1/cohere/estree';source.mkdir(parents=True,exist_ok=True); (scratch/'stage1/typescript').symlink_to(root/'stage1/typescript',target_is_directory=True)if not(scratch/'stage1/typescript').exists()else None
 for f in(root/'stage1/cohere/estree').glob('*.ts'):shutil.copyfile(f,source/f.name)
 applied=subprocess.run(['git','apply','--check',str(E/'diffs'/f'{mid}.diff')],cwd=scratch,capture_output=True,text=True)
 if applied.returncode==0:subprocess.run(['git','apply',str(E/'diffs'/f'{mid}.diff')],cwd=scratch,check=True)
 env=os.environ.copy();env.update(ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='4',ADAMIC_BUILD_CACHE_DIR='/tmp/u089/cache/replay-'+mid)
 cmd=['timeout','90','go','run','./cmd/adamic','build',str(source/'main.ts'),'-o',str(scratch/'port'),'--sanitize'];t=time.monotonic()
 with(E/(mid+'-native-build.log')).open('w')as out:
  out.write(applied.stdout+applied.stderr);out.flush();r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)if applied.returncode==0 else applied
 results[mid]=dict(exit=r.returncode,seconds=round(time.monotonic()-t,3),command=cmd);(E/'replay-checks.json').write_text(json.dumps(results,indent=2));print(mid,results[mid],flush=True)
