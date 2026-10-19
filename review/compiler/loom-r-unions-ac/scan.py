import concurrent.futures,hashlib,json,re,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-unions-ac';paths=subprocess.check_output(['git','diff','--name-only','--no-renames','--diff-filter=AM','dbdf977f5b','b2d049da7403461991e92861cf582dac0617acda','--','*.a'],text=True).splitlines();binary='/tmp/adamic-gate/unions-ac-adamic'
(e/'paths.json').write_text(json.dumps(paths,indent=2)+'\n')
def run(path):
 p=root/path;body=p.read_bytes();name=path.replace('/','-');cmd=[binary,'c',path];start=time.monotonic()
 with (e/(name+'.stdout.txt')).open('wb') as out,(e/(name+'.stderr.log')).open('wb') as err:
  try:code=subprocess.run(cmd,stdout=out,stderr=err,timeout=60).returncode
  except subprocess.TimeoutExpired:code=124
 text=(e/(name+'.stderr.log')).read_text();status='checked' if code==0 else 'NotYet' if "can't lower" in text and ' yet' in text else 'refused' if 'Adamic 0.1 refuses' in text else 'type error' if ' error TS' in text else 'failed'
 reason=text.split('Adamic 0.1 refuses ',1)[1].split(';',1)[0].strip() if status=='refused' else None
 match=re.search(r'error (TS\d+)',text);row=dict(path=path,status=status,reason=reason,code=match.group(1) if match else None,exit=code,seconds=round(time.monotonic()-start,3),diagnostics=text,body_sha256=hashlib.sha256(body).hexdigest(),header_added=None,command=cmd)
 if status in ['refused','type error']:
  header='// a-check: '+('refused '+reason if status=='refused' else 'type error '+row['code']);first=body.split(b'\n',1)[0]
  if first.startswith(b'// a-check:'):raise RuntimeError('existing unexpected header '+path)
  p.write_bytes(header.encode()+b'\n'+body);row['header_added']=header;assert p.read_bytes().split(b'\n',1)[1]==body
 return row
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=list(pool.map(run,paths))
(e/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n');print(__import__('collections').Counter(r['status'] for r in rows));assert all(r['status']!='failed' for r in rows)
