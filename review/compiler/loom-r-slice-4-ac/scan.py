import concurrent.futures,hashlib,json,re,subprocess,time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-r-slice-4-ac';binary=out/'adamic'
paths=subprocess.check_output(['git','diff','--name-only','--diff-filter=AM','HEAD','--','*.a'],text=True).splitlines()
def sha(b):return hashlib.sha256(b).hexdigest()
def scan(path):
 start=time.monotonic();log=out/'diagnostics'/(path+'.stderr.log');log.parent.mkdir(parents=True,exist_ok=True)
 with log.open('w') as err,open('/dev/null','w') as stdout:r=subprocess.run([str(binary),'c',path],stdout=stdout,stderr=err,timeout=45)
 error=log.read_text();reason=None
 if r.returncode==0:status='checked'
 elif "can't lower" in error and ' yet' in error:status='NotYet';reason=error.strip()
 elif 'Adamic 0.1 refuses' in error:status='refused';reason=error.split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
 elif ' error TS' in error:status='type error';reason='TS'+re.search(r' error TS(\d+)',error).group(1)
 else:raise AssertionError((path,r.returncode,error))
 return dict(path=path,status=status,reason=reason,exit=r.returncode,seconds=time.monotonic()-start,diagnostics=str(log.relative_to(root)))
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as ex:rows=list(ex.map(scan,paths))
(out/'before.json').write_text(json.dumps(rows,indent=2)+'\n')
changed=[]
for row in rows:
 p=Path(row['path']);before=p.read_bytes();lines=before.splitlines(keepends=True);old=lines[0] if lines and lines[0].startswith(b'// a-check:') else b'';body=before[len(old):]
 header=('// a-check: '+row['status']+' '+row['reason']+'\n').encode() if row['status'] in ('refused','type error') else b''
 if header!=old:
  p.write_bytes(header+body);assert p.read_bytes()[len(header):]==body
  changed.append(dict(path=str(p),old_header=old.decode().strip(),header=header.decode().strip(),body_sha256=sha(body),before_sha256=sha(before),after_sha256=sha(p.read_bytes())))
(out/'headers.json').write_text(json.dumps(changed,indent=2)+'\n');print('headers',len(changed),flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as ex:rows=list(ex.map(scan,paths))
(out/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
from collections import Counter
print(dict(Counter(r['status'] for r in rows)),flush=True)
(out/'statuses.md').write_text('| File | Status | Reason/code |\n| --- | --- | --- |\n'+''.join('| '+r['path']+' | '+r['status']+' | '+(r['reason'] or '').replace('|','\\|').replace('\n','<br>')+' |\n' for r in rows))
