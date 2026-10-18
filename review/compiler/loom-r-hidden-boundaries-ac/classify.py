import concurrent.futures, subprocess, pathlib, json, re, hashlib, collections
root=pathlib.Path.cwd(); evidence=root/'review/compiler/loom-r-hidden-boundaries-ac'
paths=subprocess.check_output(['git','diff','--name-only','6685711^','6685711','--','*.a'],text=True).splitlines()
def observe(path):
    body=(root/path).read_bytes() # Read whole file before editing.
    original=subprocess.check_output(['git','show','6685711:'+path])
    if body.startswith(b'// a-check:'): body=body.split(b'\n',1)[1]
    if original.startswith(b'// a-check:'): original=original.split(b'\n',1)[1]
    assert body==original, path
    p=subprocess.run(['/workspace/scratch/hidden-ac-compiler','c',path],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,text=True,timeout=45)
    diag=p.stderr
    if p.returncode==0: status='checked'; reason=''
    elif 'error TS' in diag: status='type error'; reason=re.search(r'error (TS\d+)',diag)[1]
    elif 'Adamic 0.1 refuses ' in diag: status='refused'; reason=diag.split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
    elif "stage 0 can't lower " in diag: status='NotYet'; reason=diag.splitlines()[0]
    else: raise RuntimeError((path,p.returncode,diag))
    header='// a-check: '+status+' '+reason if status in ('refused','type error') else None
    row=dict(path=path,status=status,reason=reason,exit=p.returncode,diagnostic=diag,body_sha256=hashlib.sha256(body).hexdigest(),header=header)
    if header:
        (root/path).write_bytes(header.encode()+b'\n'+body)
    return row
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool: rows=list(pool.map(observe,paths))
(evidence/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
print(collections.Counter(r['status'] for r in rows))
for r in rows: print(' '.join(v for v in (r['path'],r['status'],r['reason']) if v))
