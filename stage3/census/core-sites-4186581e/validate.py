import pathlib,json,hashlib,subprocess,os,collections,re
repo=pathlib.Path('/workspace/core-sites-main');d=repo/'stage3/census/core-sites-4186581e';s=pathlib.Path('/workspace/scratch/b5w3ycg')
r=json.loads((d/'rows.json').read_text());p=json.loads((d/'provenance.json').read_text());run=json.loads((d/'census-result.json').read_text())
raw=[json.loads(l) for l in (d/'raw.jsonl').read_text().splitlines()];core=next(x for x in raw if x.get('file','').endswith('/core.ts'))
assert len(raw)==2 and raw[0]['latent_mode']=='full' and raw[0]['diagnostics']==[] and not raw[0]['checker_rejected']
assert run['exit_code']==0 and not run['timed_out'] and run['process_group_cleaned']
assert (d/'raw.jsonl').read_bytes()==(s/'raw.jsonl').read_bytes()
assert hashlib.sha256((d/'raw.jsonl').read_bytes()).hexdigest()==p['raw_stream_sha256']
assert len(r['units'])==130 and len(r['sites'])==48 and len(r['first_walls'])==48
assert len(r['deferred_false'])==76 and sum(u['status']=='deferred_generic_instantiation' for u in r['units'])==78
assert len(r['requested_functions'])==28 and all(r['requested_functions'].values())
assert [u['where'] for u in r['units']]==[u['where'] for u in core['units'] if int(u['where'].rsplit(':',2)[1])>=1322]
assert len({u['id'] for u in r['units']})==130
assert all(f['kind'] in ['Refused','NotYet'] and f.get('emitter') for f in r['sites'])
assert not any(f['kind'] in ['panic','error','SkippedDependency'] for f in core['findings'])
for f in r['sites']:
 for e in f['emitter']:
  assert (repo/e['file']).read_text().splitlines()[e['line']-1]==e['source_line']
manifest=[]
for f in sorted(pathlib.Path('/workspace/core-sites-adapted/src').rglob('*')):
 if f.is_file():manifest.append({'path':str(f.relative_to('/workspace/core-sites-adapted')),'sha256':hashlib.sha256(f.read_bytes()).hexdigest(),'size':f.stat().st_size})
assert manifest==json.loads((d/'adapted-source-manifest.json').read_text())
assert hashlib.sha256((json.dumps(manifest,indent=2)+'\n').encode()).hexdigest()==p['source_manifest_sha256']
pid=json.loads((s/'census-pid.json').read_text())['pid']
try:os.kill(pid,0)
except ProcessLookupError:pass
else:raise RuntimeError('census process still exists')
assert not list(d.rglob('*.go'))
assert '\u2014' not in (d/'REPORT.md').read_text() and '\u2013' not in (d/'REPORT.md').read_text()
subprocess.run(['git','diff','--exit-code','HEAD','--','internal','stage3/census/latent','stage3/apply.py','stage3/apply.sh','stage3/adapt'],cwd=repo,check=True,stdout=subprocess.PIPE)
checks={'raw_stream_byte_identical':True,'raw_records':len(raw),'full_core_units':len(core['units']),'full_core_findings':len(core['findings']),'range_units':len(r['units']),'range_sites':len(r['sites']),'range_site_kinds':dict(collections.Counter(f['kind'] for f in r['sites'])),'range_site_phases':dict(collections.Counter(f['phase'] for f in r['sites'])),'first_walls':len(r['first_walls']),'deferred_generic_units':78,'deferred_instantiated_false':76,'deferred_instantiated_true':2,'requested_functions':28,'source_files_unchanged_after_census':len(manifest),'source_manifest_sha256_after_census':p['source_manifest_sha256'],'checker_diagnostics':0,'no_output_assertions':'completed before first metadata row; census exit 0','panic_or_unclassified_findings':0,'all_site_emitters_match_frozen_main_source':True,'census_pid':pid,'census_process_exists':False,'production_compiler_and_adapter_diff':'empty','new_go_files':0,'report_forbidden_dash_characters':0}
(d/'validation.json').write_text(json.dumps(checks,indent=2)+'\n');print(json.dumps(checks,indent=2))
