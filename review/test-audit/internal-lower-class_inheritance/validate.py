import pathlib,json,subprocess,time,difflib
r=pathlib.Path('/workspace/adamic');o=r/'review/test-audit/internal-lower-class_inheritance';plans=json.loads((o/'plan.json').read_text());out=[]
for m in plans:
 if m['id'].startswith('P_'):
  m['new']=m['new'].split('\n')[0]+'\n if true { '+('return nil' if m['id']=='P_FIELDS' else 'return nil, nil')+' }'
  s=(r/m['file']).read_text();changed=s.replace(m['old'],m['new']);(o/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
 m['supplemental']=m['id']=='M11'
 check=subprocess.run(['git','apply','--check',str(o/(m['id']+'.diff'))],cwd=r,capture_output=True,text=True)
 if check.returncode:raise RuntimeError(check.stderr)
 subprocess.run(['git','apply',str(o/(m['id']+'.diff'))],cwd=r,check=True)
 start=time.monotonic()
 with (o/(m['id']+'-vet.log')).open('w') as f:v=subprocess.run(['timeout','90','go','vet','./'+str(pathlib.Path(m['file']).parent)+'/'],cwd=r,stdout=f,stderr=subprocess.STDOUT)
 subprocess.run(['git','restore','--source=HEAD','--',m['file']],cwd=r,check=True)
 out.append({'id':m['id'],'apply_check':check.returncode,'vet_code':v.returncode,'wall':time.monotonic()-start});(o/'validation.json').write_text(json.dumps(out,indent=2));print(out[-1],flush=True)
(o/'plan.json').write_text(json.dumps(plans,indent=2))
