import pathlib,json,tarfile,shutil
out=pathlib.Path(__file__).resolve().parent
lane=pathlib.Path('/workspace/tracing-lane-run')
exit_record=json.loads((out/'lane.exit.json').read_text())
assert exit_record['exit']==0,exit_record
ledger=json.loads((lane/'ledger.json').read_text())
units=ledger['units'];assert len(units)==398
assert all(r['exit']==0 and r['status']=='pass' and r['seconds']<30 for r in units)
shutil.copyfile(lane/'ledger.json',out/'lane-ledger.json')
with tarfile.open(out/'lane-evidence.tar.gz','w:gz') as archive:
 for file in sorted(lane.rglob('*')):
  rel=file.relative_to(lane)
  if any(p in ('caches','adapted-tree') for p in rel.parts):continue
  if file.is_file() and not file.is_symlink():archive.add(file,arcname=str(rel),recursive=False)
counts={k:sum(r['counts'].get(k,0) for r in units) for k in ['pass','fail','skip']}
result={'implementation':'9ad919ace','base':'838809f7a','required_checks':{
'setup':0,'api_install':0,'fresh_apply':0,'normal_build_before':0,'normal_build_after':0,
'adamic_types':{'roots':79,'all_exit_zero':True},'javascript':{'files':10,'byte_identical':True},
'exports':{'string_keys':2256,'symbol_keys':0,'identical':True},
'node_witness':{'both_exit_zero':True,'types_bytes':378816,'legend_bytes':179,'byte_identical':True},
'full_census':{'before_and_after_checker_diagnostics':0,'units_each':23,'retired_variance_refusals_gone':True,
'now':'NotYet: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)'},
'core_first_stop':json.loads((out/'core-production-first-stop.json').read_text()),
'restoring_mutants':{'62':'340:14 Refused','63':'359:10 Refused'},
'lane':{'shards':398,'all_exit_zero':True,'all_pass':True,'counts':counts,'wall_seconds':ledger['wall_seconds'],'max_seconds':ledger['maximum']}},
'checkpoint_branch':'typescript/tracing-fs-views-checkpoint','evidence_branch':'typescript/tracing-fs-views-evidence'}
(out/'RESULTS.json').write_text(json.dumps(result,indent=2)+'\n')
p=out/'REPORT.md';s=p.read_text().replace('The complete 398-shard run is in progress. Final ledger and child exits will be\nadded when it finishes.',f"All 398 shards pass, all child exits are zero, and all walls are below 30\nseconds. Concurrency: {ledger['concurrency']}. Wall: {ledger['wall_seconds']:.3f} seconds. Maximum\nshard: {ledger['maximum']:.3f} seconds. Counts: {counts}.\nThe full ledger is lane-ledger.json; raw per-shard reports, observations and\nlogs are in lane-evidence.tar.gz. No cached tree or private build view is\nincluded in that archive.")
p.write_text(s)
print(json.dumps(result['required_checks']['lane'],indent=2))
