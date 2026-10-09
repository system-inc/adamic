import pathlib,subprocess,json,time
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-lower-census_overload_proof');tmp=pathlib.Path('/tmp/u027-vet');tmp.mkdir(exist_ok=True);items=json.loads((p/'manifest.json').read_text());files={x['file'] for x in items};original={f:subprocess.check_output(['git','show','7b18d057:'+f],text=True) for f in files};records=[]
empty=tmp/'empty.go';empty.write_text('package lower\n')
for x in items:
 if x['id'] not in ['M01','P01']:continue
 # The index is still the starting commit; validate diff application against it.
 check=subprocess.run(['git','apply','--cached','--check',str(p/'diffs'/(x['id']+'.diff'))],capture_output=True,text=True)
 mapping={'/workspace/adamic/internal/lower/audit_u027.go':str(empty)}
 for f,s in original.items():
  contents=s.replace(x['old'],x['new'],1) if f==x['file'] else s
  if f==x['file']:
   for unused in x.get('remove_imports',[]):contents=contents.replace(unused,'',1)
  target=tmp/(x['id']+'-'+pathlib.Path(f).name);target.write_text(contents);mapping['/workspace/adamic/'+f]=str(target)
 overlay=tmp/(x['id']+'.json');overlay.write_text(json.dumps({'Replace':mapping}));t=time.monotonic()
 cmd=['timeout','90','go','vet','-overlay='+str(overlay),'./internal/lower/']
 with (p/(x['id']+'.overlay-vet.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 records.append(dict(id=x['id'],apply_check_exit=check.returncode,exit=r.returncode,seconds=time.monotonic()-t,command='git apply --cached --check diffs/'+x['id']+'.diff; '+' '.join(cmd)));(p/'repaired-validation.json').write_text(json.dumps(records,indent=2));print(x['id'],check.returncode,r.returncode,round(records[-1]['seconds'],2),flush=True)
