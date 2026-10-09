import subprocess,json,shutil,time
from pathlib import Path
root=Path('/workspace/adamic'); E=root/'review/test-audit/internal-unicodeproperties-alias'; scratch=Path('/tmp/u075-vet');scratch.mkdir(exist_ok=True);dest=scratch/'internal/unicodeproperties';dest.mkdir(parents=True,exist_ok=True)
for p in (root/'internal/unicodeproperties').glob('*.go'):
 if p.name!='audit_switch.go':shutil.copyfile(p,dest/p.name)
original={f:Path('/tmp/u075_'+f).read_text() for f in ('unicodeproperties.go','tables.go')}
(scratch/'go.mod').write_text('module github.com/system-inc/adamic\ngo 1.27\n')
result={}
for m in json.loads((E/'menu.json').read_text()):
 for f,s in original.items():(dest/f).write_text(s)
 mid=m['id'];t=time.monotonic();apply=subprocess.run(['git','apply','--check',str(E/'diffs'/f'{mid}.diff')],cwd=scratch,capture_output=True,text=True)
 if apply.returncode==0:subprocess.run(['git','apply',str(E/'diffs'/f'{mid}.diff')],cwd=scratch,check=True)
 with (E/(mid+'-vet.log')).open('w') as out:
  out.write(apply.stdout+apply.stderr);out.flush();r=subprocess.run(['go','vet','./internal/unicodeproperties/'],cwd=scratch,stdout=out,stderr=subprocess.STDOUT) if apply.returncode==0 else apply
 result[mid]=dict(exit=r.returncode,seconds=round(time.monotonic()-t,3));(E/'replay-checks.json').write_text(json.dumps(result,indent=2))
