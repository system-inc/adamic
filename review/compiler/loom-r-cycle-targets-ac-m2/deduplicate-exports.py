import pathlib,os,re,hashlib,json
root=pathlib.Path('/home/agent/.cache/adamic-build/local'); canonical={}; hashes={}; files=0; saved=0;links=0
for d in root.iterdir():
 if not re.fullmatch('[0-9a-f]{64}',d.name):continue
 exports=d/'exports'
 if not exports.is_dir():continue
 for f in exports.glob('*.a'):
  st=f.stat();key=(st.st_dev,st.st_ino)
  if key not in hashes:hashes[key]=hashlib.file_digest(f.open('rb'),'sha256').hexdigest()
  k=(st.st_size,st.st_mode,st.st_uid,st.st_gid,hashes[key]);files+=1
  if k not in canonical:canonical[k]=f;continue
  prior=canonical[k];ps=prior.stat()
  if ps.st_ino==st.st_ino:continue
  tmp=f.with_name(f.name+'.dedup')
  os.link(prior,tmp);os.replace(tmp,f);saved+=st.st_size;links+=1
out=pathlib.Path('review/compiler/loom-r-cycle-targets-ac-m2/export-deduplication.json');out.write_text(json.dumps(dict(files=files,links=links,duplicate_bytes_linked=saved),indent=2)+'\n')
