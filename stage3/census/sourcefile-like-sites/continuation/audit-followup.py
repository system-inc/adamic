import pathlib,json,gzip,hashlib,collections
out=pathlib.Path(__file__).resolve().parents[1];follow=out/'continuation'
def load(p):return json.loads(p.read_text())
def rows(p):
 if not p.exists():p=pathlib.Path(str(p)+'.gz')
 with (gzip.open(p,'rt') if p.suffix=='.gz' else p.open()) as src:return [json.loads(line) for line in src]
files=load(follow/'variant-files.json');comparison=load(follow/'site-comparison.json');baseline=load(out/'sites.json');locations={s['location'] for s in baseline}
assert len(comparison)==2*len(locations)
assert {(r['variant'],r['location']) for r in comparison}=={(v,l) for v in ['a','b'] for l in locations}
original=load(follow/'original/summary.json');assert original['calls']==93 and original['whole_src_diagnostics']==0
assert not load(follow/'original/non-direct-helper-references.json')
assert not load(follow/'original/src-caller-coverage.json')['uncovered_function_references']
keys=lambda cs:{(c['file'],c['line'],c['column'],c['callee']) for c in cs}
source_calls=load(follow/'original/calls.json');assert {c['signature_file'] for c in source_calls}=={'src/compiler/scanner.ts'}
counts={}
for label in ['a','b']:
 directory=follow/label;summary=load(directory/'summary.json');assert summary['services_TS2386']==0
 calls=load(directory/'calls.json');assert keys(calls)==keys(source_calls)
 assert {c['signature_file'] for c in calls}=={'src/compiler/scanner.ts'}
 diff=load(directory/'source-diff.json');assert diff['only_two_parameter_types'] and set(diff['changed_files'])=={'src/compiler/scanner.ts'}
 complete=0
 for name in files:
  stem=name.replace('/','__').removesuffix('.ts');s=load(directory/'raw'/(stem+'.status.json'));r=rows(directory/'raw'/(stem+'.jsonl'))
  assert r[0]['latent_mode']=='full'
  assert s['bound_seconds']==1500 and s['max_concurrent_files']==2
  if s['child_exit']==0:
   assert r[-1]['status']=='relation_report'
   assert [x['file'] for x in r if 'file' in x]==['/workspace/scratch/baobab-road2-'+label+'/src/compiler/'+name]
   complete+=1
  else:assert s['timed_out']
 counts[label]={'completed':complete,'attempted':len(files),'stock_errors':summary['whole_src_diagnostics'],'services_TS2386':0}
for p in load(follow/'public-comparison.json'):
 assert p['only_getLineAndCharacterOfPosition_parameter_differs'] and p['free_getLineStarts_declarations']==0
 assert not p['public_SourceFile_has_lineMap']
for name in ['checker','emitter','parser','program','scanner','utilities']:
 s=load(follow/'baseline/raw'/(name+'.status.json'));assert s['bound_seconds']==1500 and s['max_concurrent_files']==1
 assert s['child_exit']==0 or s['timed_out']
assert not list(out.rglob('*.go'))
result={'audit':'passed','baseline_observed_findings':len(baseline),'baseline_unique_locations':len(locations),'caller_count':len(source_calls),'variant_site_rows':len(comparison),'variants':counts}
(follow/'audit-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
