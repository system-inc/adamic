from pathlib import Path
import hashlib,json,re
out=Path(__file__).resolve().parent;root=out.parents[2]
d=json.loads((out/'before.json').read_text());rows=[]
for r in d['rows']:
 p=root/r['path'];old=p.read_bytes();first,sep,rest=old.partition(b'\n');existing=first.decode(errors='replace') if first.startswith(b'// a-check:') else ''
 body=rest if existing else old
 if r['status']=='refused':header='// a-check: refused '+r['stderr'].split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
 elif r['status']=='type error':header='// a-check: type error TS'+re.search(r' error TS(\d+)',r['stderr'])[1]
 else:header=''
 new=(header.encode()+b'\n' if header else b'')+body
 if new!=old:p.write_bytes(new)
 rows.append({'path':r['path'],'status':r['status'],'old_header':existing,'header':header,'changed':new!=old,'body_sha256':hashlib.sha256(body).hexdigest(),'original_sha256':hashlib.sha256(old).hexdigest()})
(out/'headers.json').write_text(json.dumps(rows,indent=2)+'\n');print('headers changed',sum(r['changed'] for r in rows))
p=root/'internal/oracle/records_refusal_test.go';source=p.read_text();updated=re.sub(r'return \[3\]string\{":(\d+):',lambda m:'return [3]string{":'+str(int(m[1])+1)+':',source);assert source!=updated;p.write_text(updated)
# Both files are complete current-tree evidence scripts, never production compiler changes.
p=out/'scan.py';s=p.read_text();s=s.replace("label=sys.argv[1];diagnostics={};commands=[]","label=sys.argv[1];diagnostics={};commands=[]\nif len(sys.argv)>2:paths=json.loads(pathlib.Path(sys.argv[2]).read_text())");p.write_text(s)
