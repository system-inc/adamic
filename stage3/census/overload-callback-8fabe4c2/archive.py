import gzip,hashlib,json,pathlib,sys
out=pathlib.Path(__file__).parent
name=sys.argv[1]
result=[]
for p in sorted(out.glob(name+'*.jsonl')):
 raw=p.read_bytes(); dest=p.with_suffix(p.suffix+'.gz')
 with dest.open('wb') as f:
  with gzip.GzipFile(filename='',mode='wb',fileobj=f,mtime=0) as z:z.write(raw)
 result.append(dict(file=dest.name,bytes=len(raw),sha256=hashlib.sha256(raw).hexdigest(),compressed_bytes=dest.stat().st_size,compressed_sha256=hashlib.sha256(dest.read_bytes()).hexdigest()))
 p.unlink()
(out/(name+'-streams.json')).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
