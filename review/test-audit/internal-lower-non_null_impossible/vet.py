import pathlib,json,subprocess,time,re
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-lower-non_null_impossible');tmp=pathlib.Path('/tmp/u040-vet');tmp.mkdir(exist_ok=True);original=json.loads((p/'original-files.json').read_text());items=json.loads((p/'manifest.json').read_text());empty=tmp/'empty.go';empty.write_text('package lower\n');records=[]
for x in items:
 check=subprocess.run(['git','apply','--cached','--check',str(p/'diffs'/(x['id']+'.diff'))],capture_output=True,text=True);mapping={'/workspace/adamic/internal/lower/audit_u040.go':str(empty)}
 for f,s in original.items():
  new=s.replace(x['old'],x['new'],1) if 'internal/lower/'+f==x['file'] else s
  if x['mode']=='probe' and 'internal/lower/'+f==x['file']:
   for line in new.splitlines(True):
    m=re.match(r'\s*"([^\"]+)"\s*$',line)
    if m and m[1].split('/')[-1]+'.' not in new:new=new.replace(line,'',1)
  target=tmp/(x['id']+'-'+f);target.write_text(new);mapping['/workspace/adamic/internal/lower/'+f]=str(target)
 overlay=tmp/(x['id']+'.json');overlay.write_text(json.dumps({'Replace':mapping}));cmd=['timeout','90','go','vet','-overlay='+str(overlay),'./internal/lower/'];t=time.monotonic()
 with (p/(x['id']+'.vet.log')).open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
 records.append(dict(id=x['id'],apply_check_exit=check.returncode,exit=r.returncode,seconds=time.monotonic()-t,command='git apply --cached --check diffs/'+x['id']+'.diff; '+' '.join(cmd)));(p/'standalone-validation.json').write_text(json.dumps(records,indent=2));print(x['id'],check.returncode,r.returncode,flush=True)
