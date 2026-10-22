from pathlib import Path
import hashlib,json,os,time
root=Path('/tmp/adamic-gate/slice3-products/local');seen={};rows=[];digests={}
for listing in sorted(root.glob('*/listing.json')):
 if listing.parent.name.startswith('.'):continue
 for record in json.loads(listing.read_text()):
  if not record.get('Export'):continue
  p=listing.parent/record['Export'];size=p.stat().st_size;key=(record['ImportPath'],size)
  stat=p.stat();inode=(stat.st_dev,stat.st_ino)
  if inode not in digests:
   with p.open('rb') as f:digests[inode]=hashlib.file_digest(f,'sha256').hexdigest()
  digest=digests[inode];identity=(*key,digest)
  if identity not in seen:seen[identity]=p;continue
  original=seen[identity]
  if os.path.samefile(p,original):continue
  link=p.with_name(p.name+'.deduplicate-'+str(os.getpid()));os.link(original,link);os.replace(link,p);rows.append(dict(package=key[0],bytes=size,sha256=digest))
out=Path(__file__).resolve().parent
(out/('export-deduplication-'+str(time.time_ns())+'.json.txt')).write_text(json.dumps(rows,indent=2)+'\n')
print('Verified duplicate export bytes reclaimed:',sum(r['bytes'] for r in rows))
