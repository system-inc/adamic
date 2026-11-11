import pathlib,json,hashlib,sys
out=pathlib.Path(__file__).resolve().parent
root=pathlib.Path('/workspace/scratch/baobab-adapted/src/compiler')
files={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*.ts'))}
assert len(files)==79,len(files)
manifest=out/'adapted-source-sha256.json'
if manifest.exists():
 assert json.loads(manifest.read_text())==files,'adapted source changed'
 print('verified 79 adapted input hashes')
else:
 manifest.write_text(json.dumps(files,indent=2)+'\n')
 p=out/'provenance.json';x=json.loads(p.read_text());x['binary_sha256']=hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-latent').read_bytes()).hexdigest();x['adapted_root']=str(root);x['compiler_files']=len(files);p.write_text(json.dumps(x,indent=2)+'\n')
 print('sealed 79 adapted input hashes and binary hash')
