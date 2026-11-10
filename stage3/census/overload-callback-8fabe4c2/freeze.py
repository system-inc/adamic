import hashlib,json,pathlib,subprocess,shutil
out=pathlib.Path(__file__).parent;root=pathlib.Path('/workspace/overload-adapted')
manifest={p.relative_to(root).as_posix():hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'src').rglob('*')) if p.is_file()}
(out/'source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
shutil.copyfile(root/'patch-set.md',out/'patch-set.md')
for p in ['setup','source','npm','fetch']:
 src=pathlib.Path('/tmp/overload-'+p+'.log')
 if src.exists():shutil.copyfile(src,out/('initial-'+p+'.log.txt'))
records={}
for name in ['parent','head']:
 overlay=pathlib.Path('/workspace/overload-'+name+'-entry/overlay.json')
 data=json.loads(overlay.read_text())['Replace'];items=[]
 for src,target in sorted(data.items()):
  items.append(dict(source=src,overlay=target,sha256=hashlib.sha256(pathlib.Path(target).read_bytes()).hexdigest()))
 records[name]=items
 shutil.copyfile(overlay.parent/'entry_main.go',out/(name+'-entry-driver.go.txt'))
 binary=pathlib.Path('/workspace/overload-'+name+'-census');records[name+'_binary_sha256']=hashlib.sha256(binary.read_bytes()).hexdigest()
(out/'overlay-hashes.json').write_text(json.dumps(records,indent=2)+'\n')
print('frozen source files',len(manifest))
