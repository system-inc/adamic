import pathlib,json,gzip,hashlib,subprocess
root=pathlib.Path('/workspace/or3-report');out=root/'stage3/census/overload-result-02d48211'
manifest=json.loads((out/'adapted-source-hashes.json').read_text());assert all(hashlib.sha256((pathlib.Path('/workspace/or3-adapted')/f).read_bytes()).hexdigest()==h for f,h in manifest.items())
streams={};allrows={}
for label,sha in [('parent','19ab1c58'),('head','02d48211')]:
 d=pathlib.Path('/workspace/or3-build-'+label);rs=[json.loads(s) for s in (d/'latent.jsonl').read_text().splitlines()];allrows[label]=rs
 assert rs[-1]['status']=='relation_report';assert sorted(r['file'] for r in rs if 'file' in r)==sorted(rs[0]['sources'])
 assert json.loads((d/'latent.time.json').read_text())['exit_code']==0
 for name in ['latent.jsonl','latent-times.jsonl']:
  raw=(d/name).read_bytes();p=out/(label+'-'+name);archive=p.with_name(p.name+'.gz') if len(raw)>1000000 else p
  actual=gzip.decompress(archive.read_bytes()) if archive.suffix=='.gz' else archive.read_bytes();assert actual==raw
  streams[label+'-'+name]={'sha256':hashlib.sha256(raw).hexdigest(),'bytes':len(raw),'archived_as':archive.name}
 p=subprocess.run(['git','diff','--exit-code',sha,'--','.'],cwd='/workspace/or3-'+label,capture_output=True,text=True);print(label+' source_diff_child_exit='+str(p.returncode));assert p.returncode==0,p.stdout+p.stderr
assert allrows['parent'][0]==allrows['head'][0]
a=json.loads((out/'unit-analysis.json').read_text());assert len(a['targets'])==43 and len(a['other_overload'])==67
cl=json.loads((out/'overloaded-value-classification.json').read_text())['results'];assert len(cl)==len(a['overloaded_value_units']);assert {r['unit']['where'] for r in cl}=={r['unit']['where'] for r in a['overloaded_value_units']}
assert all(r['overloaded_set']['declarations'] and r['finding']['reason']=='an overloaded function as a value' for r in cl)
head_units={u['where'] for row in allrows['head'] for u in row.get('units') or []}; paired_first={}
for row in allrows['head']:
 for f in row.get('findings') or []:
  if f.get('phase') in ('lowering','deferred') and f['unit'] in head_units:paired_first.setdefault(f['unit'],f)
assert {w for w,f in paired_first.items() if f['reason']=='an overloaded function as a value'}=={r['unit']['where'] for r in cl}
assert len((out/'43-units.md').read_text().splitlines())==45;assert len((out/'67-units.md').read_text().splitlines())==69
(out/'validation.json').write_text(json.dumps({'adapted_source_files_checked':len(manifest),'adapted_sources_unchanged':True,'compiler_sources_unchanged':True,'headers_identical':True,'completed_reach_files':len(allrows['head'][0]['sources']),'stream_integrity':streams,'original_cohort_sizes':[43,67],'all_overloaded_value_first_stops_classified':len(cl)},indent=2)+'\n')
print('all stream, source, coverage and classification checks passed')
