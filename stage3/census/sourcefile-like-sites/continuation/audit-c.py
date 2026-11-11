import pathlib,json,gzip,hashlib,collections
out=pathlib.Path(__file__).resolve().parents[1];p=out/'continuation'
def load(path):return json.loads(path.read_text())
def stream(path):
 if not path.exists():path=pathlib.Path(str(path)+'.gz')
 with (gzip.open(path,'rt') if path.suffix=='.gz' else path.open()) as src:return [json.loads(line) for line in src]
a=load(p/'c-analysis.json');comparison=load(p/'c-site-comparison.json');base=load(out/'sites.json');locations={s['location'] for s in base}
assert len(locations)==34 and len(comparison)==68
assert {(r['variant'],r['location']) for r in comparison}=={(v,l) for v in ['c','c-public'] for l in locations}
source=pathlib.Path('/workspace/scratch/baobab-adapted')
assert hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-latent').read_bytes()).hexdigest()==load(out/'provenance.json')['binary_sha256']
manifest=load(out/'adapted-source-sha256.json')
assert manifest=={str(f.relative_to(source/'src/compiler')):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((source/'src/compiler').rglob('*.ts'))}
original_calls=load(p/'original/calls.json');keys=lambda cs:{(c['file'],c['line'],c['column'],c['callee']) for c in cs}
assert len(original_calls)==93
for label in ['c','c-public']:
 r=a['variants'][label];root=pathlib.Path('/workspace/scratch/baobab-road2-'+label)
 assert set(r['source_diff'])=={'src/compiler/types.ts','src/compiler/scanner.ts'}
 expected=(source/'src/compiler/scanner.ts').read_bytes().replace(b'    type LineAndCharacter,',b'    type LineAndCharacter, type LineMapSource,',1)
 for name in ['getLineStarts','getLineAndCharacterOfPosition']:
  before=('export function '+name+'(sourceFile: SourceFileLike').encode();assert expected.count(before)==1;expected=expected.replace(before,before.replace(b'SourceFileLike',b'LineMapSource'),1)
 assert expected==(root/'src/compiler/scanner.ts').read_bytes()
 assert len(r['files'])==10 and all(s['child_exit']==0 and s['complete_stream'] and not s['timed_out'] and s['bound_seconds']==1500 for s in r['files'])
 for s in r['files']:
  name=s['file'];stem=name.replace('/','__').removesuffix('.ts');rows=stream(p/label/'raw'/(stem+'.jsonl'))
  assert rows[0]['latent_mode']=='full' and rows[0]['checker_rejected'] is False and rows[0]['diagnostics']==[]
  assert rows[-1]['status']=='relation_report'
  assert [x['file'] for x in rows if 'file' in x]==[str(root/'src/compiler'/name)]
  assert s['max_concurrent_files']==2
 views=[v for v in comparison if v['variant']==label];helper=[v for v in views if v['callee'] in ['getLineStarts','getLineAndCharacterOfPosition']]
 assert len(helper)==22 and all(v['verdict']=='no finding at original call; unit attempted' and not v['findings'] and not v['checker_diagnostics'] and not v['go_checker_diagnostics'] for v in helper)
 assert sum(bool(v['findings']) for v in views)==12
 assert r['stock']['whole_src_diagnostics']==0 and r['stock']['calls']==93 and r['stock']['services_TS2386']==0
 assert r['stock_exact']['whole_src_diagnostics']==416 and r['stock_exact']['services_TS2386']==0
 assert r['adamic_whole_src_diagnostics']==6851
 assert r['adamic_whole_src_delta']['added']==[] and r['adamic_whole_src_delta']['removed']==[]
 assert r['stock_exact_delta']['added']==[] and r['stock_exact_delta']['removed']==[]
 assert r['cohere_summary']['rules']==40 and r['cohere_summary']['checked']==77 and not r['cohere_summary']['gaps']['ruleCrashes']
 assert r['cohere_delta']['added']==[] and len(r['cohere_delta']['removed'])==24
 assert all(f['rule']=='adamic/no-optional-widening' for f in r['cohere_delta']['removed'])
 assert r['core_first_stop']['kind']=='Refused' and r['core_first_stop']['where'].endswith('/tracing.ts:232:34') and r['core_first_stop']['reason']=="a cast the runtime can't check"
 assert [(f['kind'],f['reason']) for f in r['getLineStarts_findings']]==[('NotYet','an assignment value to a member'),('Boundary','failed statement; state rolled back; continued at next statement or declaration')]
 assert len(r['getLineStarts_units'])==1 and r['getLineStarts_units'][0]['status']=='attempted'
 assert keys(load(p/label/'calls.json'))==keys(original_calls)
 assert load(p/label/'src-caller-coverage.json')['uncovered_function_references']==[]
# Public-marker-only difference is literal, not inferred from matching output.
a_bytes=pathlib.Path('/workspace/scratch/baobab-road2-c/src/compiler/types.ts').read_bytes();b_bytes=pathlib.Path('/workspace/scratch/baobab-road2-c-public/src/compiler/types.ts').read_bytes()
needle=b'/** @internal */\r\nexport interface LineMapSource';assert a_bytes.count(needle)==1 and a_bytes.replace(needle,b'/**              */\r\nexport interface LineMapSource',1)==b_bytes
assert load(p/'c/public-emit.log.status.json')['child_exit']==1
assert not (p/'c/typescript.d.ts').exists() and not (p/'c/typescript.d.ts.gz').exists()
assert load(p/'c/declaration-emit.json')['diagnostics']==[] and len(load(p/'c/declaration-emit.json')['emitted'])==278
for label in ['c-public','c-public-unstripped']:
 api=load(p/label/'api-analysis.json');assert api['diagnostics']==[]
 selected=api['selected_declarations'];assert any(d['name']=='LineMapSource' for d in selected) and not any(d['name']=='getLineStarts' for d in selected)
 assert next(d['text'] for d in selected if d['name']=='getLineAndCharacterOfPosition')=='function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;'
assert load(p/'original/lane-api-check.json')['errors']==[]
assert len(load(p/'c-unstripped/lane-api-check.json')['errors'])==4
assert len(load(p/'c-public/lane-api-check.json')['errors'])==7
assert [d['code'] for d in load(p/'c-unstripped/api-analysis.json')['diagnostics']]==[2304,2304]
assert len(load(p/'original/raw-declaration-analysis.json')['diagnostics'])==2 and len(load(p/'c/raw-declaration-analysis.json')['diagnostics'])==4
sdk=load(p/'c-sdk-analysis.json')
assert sdk['variants']['original']['cohere_findings']==[]
for label,entry in sdk['variants'].items():
 assert entry['checker']['accepted'] and entry['checker']['diagnostics']==[]
 related=[f for f in entry['guarded_findings'] if f['kind']=='Refused' and f['reason']=='optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields']
 if label=='original':assert not related
 else:
  assert len(related)==1
  assert len(entry['cohere_findings'])==1 and entry['cohere_findings'][0]['rule']=='adamic/no-optional-widening'
assert not list(out.rglob('*.go'))
text=(out/'REPORT.md').read_text();assert '\u2013' not in text and '\u2014' not in text
assert text.count('## Third measurement: named LineMapSource, variant (c)')==1
result={'audit':'passed','variants':2,'complete_full_streams':20,'original_views_per_variant':34,'helper_views_cleared_per_variant':22,'unchanged_other_views_per_variant':12,'cohere_new_findings_per_variant':0,'cohere_removed_optional_widening_per_variant':24,'production_whole_src_diagnostics_per_variant':6851,'new_production_checker_diagnostics_per_variant':0,'services_TS2386':0,'internal_strip_bundle_child_exit':1,'public_variant_closure_diagnostics':0,'public_variant_lane_api_errors':7,'public_sdk_new_optional_widening_refusals':1,'readonly_sdk_control_refused':True,'no_new_refusal_anywhere_outcome':'not met','lowering_deltas':{label:{key:len(a['variants'][label]['lowering_delta'][key]) for key in ['added','removed']} for label in ['c','c-public']}}
(p/'c-audit-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
