import pathlib,gzip,hashlib,json,shutil
out=pathlib.Path('/workspace/never-call-report/stage3/census/never-call-392cb60b')
for p in list(out.glob('*.jsonl')):
 if p.stat().st_size>20_000_000:
  with p.open('rb') as a, (out/(p.name+'.gz')).open('wb') as f:
   with gzip.GzipFile(filename='',mode='wb',fileobj=f,mtime=0) as b:shutil.copyfileobj(a,b)
  p.unlink()
rows=[]
for p in sorted(out.iterdir()):
 if p.is_file() and p.name!='SHA256SUMS':
  rows.append(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name)
(out/'SHA256SUMS').write_text('\n'.join(rows)+'\n')
print('archive child exit code: 0')
