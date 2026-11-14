"""Final independent evidence integrity audit; no compiler execution or mutation."""
import gzip,hashlib,json,subprocess
from pathlib import Path
OUT=Path(__file__).resolve().parent;BASE=Path('/workspace/scratch/c4v7-census');NS=OUT.parent/'next-stops-c4v7'
read=lambda p:json.loads(p.read_text())
hash=lambda data:hashlib.sha256(data).hexdigest()
for label in ('main','head'):
 manifest=read(OUT/f'source-manifest-{label}.json')
 assert len(manifest)==747
 current_names={str(p.relative_to(BASE/f'adapted-{label}')) for p in (BASE/f'adapted-{label}/src').rglob('*') if p.is_file()}
 assert current_names==set(manifest)
 for name,expected in manifest.items():assert hash((BASE/f'adapted-{label}'/name).read_bytes())==expected,(label,name)
 assert hash((BASE/f'scan-{label}').read_bytes())==read(OUT/f'build-{label}.json')['binary_sha256']
 assert 'lower.LatentScanOnly(context.Background(), program,' in (OUT/f'driver-{label}.go.txt').read_text()
 for i in range(1,80):
  folder=OUT/'part2'/label;status=read(folder/f'{i:02}-status.json')
  assert status['exit_code']==0 and not status['timeout'] and status['seconds']<1510 and status['process_group_cleaned']
  for suffix in ['.jsonl','.log']:
   raw=gzip.decompress((folder/f'{i:02}{suffix}.gz').read_bytes());assert hash(raw)==status[suffix+'_sha256']
   if suffix=='.jsonl':rows=[json.loads(x) for x in raw.splitlines()]
  assert rows[0]['checker_rejected'] is False and rows[0]['diagnostics']==[]
  records=[x for x in rows if x.get('file')];assert len(records)==1 and records[0]['units']==[]
  assert rows[-1]['status']=='relation_report' and rows[-1]['findings']==[]
  assert not rows[-1]['relation_overflow']['after_load'] and not rows[-1]['relation_overflow']['end']
  assert all(f['phase']=='refusal_scan' for f in records[0]['findings'])
pair=read(OUT/'pairing.json');assert pair['before']-pair['gone']+pair['new']==pair['after']
for label in ['gone','new','changed']:assert sum(c[label] for c in pair['by_reason'].values())==pair[label]
new=read(OUT/'new-findings-attributed.json');report=(OUT/'REPORT.md').read_text()
assert len(new)==pair['new']==12
for f in new:
 assert f['raw_text'] in report and '### ' in report and f['where'] in report
 assert f['rule'] or f.get('registered_rule')
 assert f['enclosing_function'] and not f['generic_bodies']
for root in read(NS/'CommonJSBundle/roots.json'):
 name=root.split('/src/compiler/')[-1];assert '| `'+name+'` |' in report,name
assert pair['selected_191']['present_before']==pair['selected_191']['gone']==191
assert read(OUT/'scout-site-mapping.json')['scout_sys_sha256']==read(OUT/'scout-site-mapping.json')['reconstructed_scout_sys_sha256']
final=read(NS/'final-validation.json');assert final['observations']==316 and final['duplicates_agree']
changed=subprocess.check_output(['git','diff','1d48ecb5a5321717daf901579e49a23291995589','--name-only'],text=True).splitlines()+subprocess.check_output(['git','ls-files','--others','--exclude-standard'],text=True).splitlines()
assert all(n.startswith(('stage3/census/c4v7-census/','stage3/census/next-stops-c4v7/')) for n in changed)
assert not any(n.endswith('.go') for n in changed)
ignored=subprocess.check_output(['git','ls-files','--others','--ignored','--exclude-standard',str(OUT),str(NS)],text=True).splitlines()
assert not ignored,('unpackaged ignored evidence',ignored)
result=dict(source_files_per_commit=747,unchanged_source_hashes_verified=True,scan_binaries_verified=True,raw_scan_processes_verified=158,zero_lowering_units=True,all_raw_hashes_verified=True,all_headers_checker_clean=True,relation_reports_clean=True,all_new_diagnostic_texts_quoted=12,all_79_root_table_rows_present=True,scout_mapping_verified=True,ordinary_observations=316,compiler_and_adaptations_unchanged=True)
(OUT/'audit.json').write_text(json.dumps(result,indent=2)+'\n');print('final evidence audit passed')
