from pathlib import Path
import subprocess,json,re,hashlib,time,collections,importlib.util,sys,gzip,difflib
root=Path.cwd();ev=root/'review/compiler/loom-r-views-fx-a-ac';cli='/tmp/views-ac-adamic'
paths=subprocess.check_output(['git','diff','--name-only','--diff-filter=AM','dbdf977f5b','6c733f96','--','*.a'],text=True).splitlines()
assert len(paths)==54,len(paths)
def body(data):return data.split(b'\n',1)[1] if data.startswith(b'// a-check:') else data
def digest(data):return hashlib.sha256(data).hexdigest()
def measure(path):
 start=time.monotonic();p=subprocess.run([cli,'c',path],capture_output=True,text=True,timeout=45)
 text=p.stderr
 if p.returncode==0:status,reason='checked',''
 elif "can't lower" in text and ' yet' in text:status,reason='NotYet',text.strip().splitlines()[0]
 elif 'Adamic 0.1 refuses' in text:status,reason='refused',text.split('Adamic 0.1 refuses ',1)[1].split(';',1)[0].strip()
 elif ' error TS' in text:status,reason='type error','TS'+re.search(r' error TS(\d+)',text).group(1)
 else:status,reason='failed',text.strip().splitlines()[0] if text.strip() else 'unclassified compiler failure'
 return dict(status=status,reason=reason,exit=p.returncode,stderr=text,stdout_bytes=len(p.stdout.encode()),seconds=round(time.monotonic()-start,4))
rows=[]
for path in paths:
 original=subprocess.check_output(['git','show','6c733f96:'+path]);current=(root/path).read_bytes()
 assert body(current)==body(original),(path,'program body changed')
 row=dict(path=path,body_sha256_before=digest(body(current)),original_sha256=digest(original),original_header=original.split(b'\n',1)[0].decode() if original.startswith(b'// a-check:') else '',observed_header_at_scan_start=current.split(b'\n',1)[0].decode() if current.startswith(b'// a-check:') else '')
 row['before']=measure(path)
 # A header can shift an internal location in the refusal reason. Reobserve after
 # insertion and update only its text until the single header line is stable.
 observation=row['before']
 for attempt in range(3):
  header=('// a-check: refused '+observation['reason']) if observation['status']=='refused' else ('// a-check: type error '+observation['reason']) if observation['status']=='type error' else ''
  desired=(header.encode()+b'\n' if header else b'')+body(original)
  if (root/path).read_bytes()!=desired:(root/path).write_bytes(desired)
  fresh=measure(path)
  assert fresh['status']==row['before']['status'],(path,'status changed',fresh)
  wanted=('// a-check: refused '+fresh['reason']) if fresh['status']=='refused' else ('// a-check: type error '+fresh['reason']) if fresh['status']=='type error' else ''
  if wanted==header:break
  observation=fresh
 else:raise RuntimeError((path,'unstable header'))
 row.update(status=fresh['status'],reason=fresh['reason'],header=header,after=fresh,header_changed=desired!=original,body_sha256_after=digest(body(desired)))
 assert row['body_sha256_after']==row['body_sha256_before'],path
 rows.append(row);(ev/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(path,row['status'],flush=True)
(ev/'paths.json').write_text(json.dumps(paths,indent=2)+'\n')
print('TOTAL',dict(collections.Counter(row['status'] for row in rows)),flush=True)
