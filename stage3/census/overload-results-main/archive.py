import pathlib,json,gzip,hashlib,shutil
out=pathlib.Path(__file__).resolve().parent
manifest=json.loads((out/'adapted-source-hashes.json').read_text())
for rel,sha in manifest.items():assert hashlib.sha256((pathlib.Path('/workspace/orm-adapted')/rel).read_bytes()).hexdigest()==sha,rel
ledger={}
def retain(source,destination):
 source=pathlib.Path(source);destination=pathlib.Path(destination);destination.parent.mkdir(parents=True,exist_ok=True)
 data=source.read_bytes()
 if destination.suffix=='.gz':
  with destination.open('wb') as f:
   with gzip.GzipFile(fileobj=f,mode='wb',mtime=0) as z:z.write(data)
 else:destination.write_bytes(data)
 ledger[str(destination.relative_to(out))]=dict(bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
for label in ['main']:
 build=pathlib.Path('/workspace/orm-build-'+label)
 for i in range(8):
  d=json.loads((build/f'shard-{i}.time.json').read_text());assert d['exit_code']==0 and not d['timed_out'] and d['wall_seconds']<1800
 for p in sorted(build.iterdir()):
  if p.suffix in ('.log','.jsonl') or p.name.endswith('.time.json'):
   name=p.name+('.gz' if p.suffix in ('.jsonl','.log') else '')
   retain(p,out/label/name)
 meta=json.loads((build/'entry-overlay/overlay.json').read_text())
 retain(build/'entry-overlay/overlay.json',out/label/'entry-overlay.json')
 paths=[pathlib.Path(p) for p in meta['Replace'].values() if 'func latentRun(' in pathlib.Path(p).read_text()]
 assert len(paths)==1
 retain(paths[0],out/(label+'-sharded-lower.go.txt'))
 scanmeta=json.loads((build/'scan-overlay/overlay.json').read_text());retain(build/'scan-overlay/overlay.json',out/label/'scan-overlay.json')
 paths=[pathlib.Path(p) for p in scanmeta['Replace'].values() if 'func latentRun(' in pathlib.Path(p).read_text()];assert len(paths)==1
 retain(paths[0],out/(label+'-scan-sharded-lower.go.txt'))
for pattern in ['orm-main-submodules.log','orm-provenance.time.json','orm-merge*','orm-build-scan*','orm-fixture-scan*','orm-apply*','orm-setup*','orm-npm.log','orm-provenance.log','orm-pair-tests*','orm-fixture-main.log','orm-build-main.log','orm-resolve-preliminary*','orm-pair.log','orm-pair.time.json','orm-analysis.log','orm-analysis.time.json','orm-resolve.log','orm-resolve.time.json','orm-inventory-main*']:
 for p in pathlib.Path('/tmp').glob(pattern):
  if p.is_file():retain(p,out/'preparation'/(p.name+('.gz' if p.suffix=='.log' else '')))
(out/'archive-manifest.json').write_text(json.dumps(ledger,indent=2)+'\n');print('archived',len(ledger),'files; shared source hashes unchanged')
