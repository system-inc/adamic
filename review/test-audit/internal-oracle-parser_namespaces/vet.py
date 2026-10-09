import pathlib,json,subprocess,re,time
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');tmp=pathlib.Path('/tmp/u067-vet');tmp.mkdir(exist_ok=True);orig=json.loads((p/'original-files.json').read_text());empty=tmp/'empty.go';empty.write_text('package lower\n');records=[]
for x in json.loads((p/'manifest.json').read_text()):
 check=subprocess.run(['git','apply','--cached','--check',str(p/'diffs'/(x['id']+'.diff'))],capture_output=True,text=True);mapping={str(pathlib.Path('internal/lower/audit_u067.go').resolve()):str(empty)}
 for f,s in orig.items():
  one=s.replace(x['old'],x['new'],1) if f==x['file'] else s
  if x['id'].startswith('P') and f==x['file']:
   for l in one.splitlines(True):
    m=re.match(r'\s*"([^\"]+)"\s*$',l)
    if m and m[1].split('/')[-1]+'.' not in one:one=one.replace(l,'',1)
  t=tmp/(x['id']+'-'+f.replace('/','_'));t.write_text(one);mapping[str(pathlib.Path(f).resolve())]=str(t)
 overlay=tmp/(x['id']+'.json');overlay.write_text(json.dumps({'Replace':mapping}));cmd=['timeout','90','go','vet','-overlay='+str(overlay),'./internal/oracle/' if x['id']=='W01' else './internal/lower/'];t=time.monotonic()
 with (p/(x['id']+'.vet.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(id=x['id'],apply_exit=check.returncode,vet_exit=r.returncode,seconds=time.monotonic()-t,command='git apply --cached --check '+str(p/'diffs'/(x['id']+'.diff'))+'; '+' '.join(cmd)));(p/'standalone-validation.json').write_text(json.dumps(records,indent=2));print(x['id'],check.returncode,r.returncode,flush=True)
