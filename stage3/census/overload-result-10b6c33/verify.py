import pathlib,json,hashlib,gzip,subprocess
root=pathlib.Path('/workspace/overload-head');out=root/'stage3/census/overload-result-10b6c33'
manifest=json.loads((out/'adapted-source-hashes.json').read_text())
changed=[f for f,h in manifest.items() if hashlib.sha256((pathlib.Path('/workspace/overload-adapted')/f).read_bytes()).hexdigest()!=h]
assert not changed,changed
streams={}
for label in ['parent','head']:
 d=pathlib.Path('/workspace/overload-build-'+label)
 rows=[json.loads(s) for s in (d/'latent.jsonl').read_text().splitlines()]
 assert rows[-1]['status']=='relation_report'
 assert sorted(r['file'] for r in rows if 'file' in r)==sorted(rows[0]['sources'])
 for name in ['latent.jsonl','latent-times.jsonl']:
  raw=(d/name).read_bytes();p=out/(label+'-'+name)
  archive=p.with_name(p.name+'.gz') if len(raw)>1000000 else p
  actual=gzip.decompress(archive.read_bytes()) if archive.suffix=='.gz' else archive.read_bytes()
  assert raw==actual,archive
  streams[label+'-'+name]={'decompressed_sha256':hashlib.sha256(raw).hexdigest(),'bytes':len(raw),'archived_as':archive.name}
 stats=json.loads((d/'latent.time.json').read_text());assert stats['exit_code']==0
for label,sha in [('parent','efe88776'),('head','10b6c33')]:
 cmd=['git','diff','--exit-code',sha,'--','.']
 if label=='head':cmd+=[':(exclude)stage3/census/overload-result-10b6c33']
 p=subprocess.run(cmd,cwd='/workspace/overload-'+label,capture_output=True,text=True)
 print(label+' source_diff_child_exit='+str(p.returncode));assert p.returncode==0,p.stdout+p.stderr
(out/'validation.json').write_text(json.dumps({'adapted_source_files_checked':len(manifest),'adapted_source_changes':changed,'stream_integrity':streams,'source_diffs_clean':True},indent=2)+'\n')
print('source and stream integrity validated')
