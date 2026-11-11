import pathlib,json,gzip,hashlib,collections
out=pathlib.Path(__file__).resolve().parents[1];p=out/'continuation'
def load(file):return json.loads(file.read_text()) if file.exists() else json.loads(gzip.decompress(pathlib.Path(str(file)+'.gz').read_bytes()))
def stream(file):
 if not file.exists():file=pathlib.Path(str(file)+'.gz')
 with (gzip.open(file,'rt') if file.suffix=='.gz' else file.open()) as src:return [json.loads(line) for line in src]
a=load(p/'e-analysis.json');sites=load(p/'e-site-comparison.json');base=pathlib.Path('/workspace/scratch/baobab-adapted');root=pathlib.Path('/workspace/scratch/baobab-road2-e')
assert len(sites)==34 and len({s['location'] for s in sites})==34
assert a['site_counts']=={'cleared at original view':22,'unchanged finding':12},a['site_counts']
assert len(a['files'])==10 and all(r['complete_stream'] and r['child_exit']==0 and not r['timed_out'] and r['bound_seconds']==1500 and r['max_concurrent_files']==2 for r in a['files'])
for r in a['files']:
 stem=r['file'].replace('/','__').removesuffix('.ts');rows=stream(p/'e/raw'/(stem+'.jsonl'))
 assert rows[0]['checker_rejected'] is False and rows[0]['diagnostics']==[] and rows[0]['latent_mode']=='full'
 assert rows[-1]['status']=='relation_report'
 assert [x['file'] for x in rows if 'file' in x]==[str(root/'src/compiler'/r['file'])]
assert hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-latent').read_bytes()).hexdigest()==load(out/'provenance.json')['binary_sha256']
assert load(out/'adapted-source-sha256.json')=={str(f.relative_to(base/'src/compiler')):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((base/'src/compiler').rglob('*.ts'))}
# Verify the exact two source edits, including the existing SourceFileLike JSDoc and all members.
old=(base/'src/compiler/types.ts').read_bytes();doc=b'/**\r\n * Subset of properties from SourceFile that are used in multiple utility functions\r\n */\r\nexport interface SourceFileLike {'
added=b'/** @internal */\r\nexport interface LineMapSource {\r\n    readonly text: string;\r\n    lineMap?: readonly number[] | undefined;\r\n}\r\n\r\n'
assert old.count(doc)==1
assert old.replace(doc,added+doc,1)==(root/'src/compiler/types.ts').read_bytes()
old=(base/'src/compiler/scanner.ts').read_bytes();expected=old.replace(b'    type LineAndCharacter,',b'    type LineAndCharacter, type LineMapSource,',1)
expected=expected.replace(b'export function getLineStarts(sourceFile: SourceFileLike)',b'export function getLineStarts(sourceFile: LineMapSource)',1)
before=b'export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter {'
after=b'/** @internal */\r\nexport function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;\r\nexport function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter; // eslint-disable-line @typescript-eslint/unified-signatures\r\nexport function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter {'
assert expected.count(before)==1 and expected.replace(before,after,1)==(root/'src/compiler/scanner.ts').read_bytes()
diffs=[]
for f in (base/'src').rglob('*'):
 if f.is_file():
  rel=f.relative_to(base);assert (root/rel).is_file()
  if f.read_bytes()!=(root/rel).read_bytes():diffs.append(str(rel))
assert set(diffs)=={'src/compiler/types.ts','src/compiler/scanner.ts'}
assert not [f for f in (root/'src').rglob('*') if f.is_file() and not (base/f.relative_to(root)).exists()]
assert len(a['routing'])==18
for r in a['routing']:
 assert r['native']['chosen_overload']==1 and r['native']['chosen_first_parameter']=='LineMapSource'
 assert r['native']['chosen_declaration'].endswith('/scanner.ts:549:1')
 s=r['stock']['signature_resolution'];assert s['chosen_overload']==1 and s['chosen']['first_parameter']=='LineMapSource' and s['chosen']['internal']
 assert [x['first_parameter'] for x in s['signatures']]==['LineMapSource','SourceFileLike']
native=load(p/'e/native-overload-resolution.json');assert native['compiler_roots']==79 and len(native['calls'])==20 and all(r['chosen_overload']==1 for r in native['calls'])
calls=load(p/'e/calls.json');assert len(calls)==93
assert sum(c['callee']=='getLineAndCharacterOfPosition' for c in calls)==79
assert all(c['signature_resolution']['chosen_overload']==1 for c in calls if c['callee']=='getLineAndCharacterOfPosition')
assert a['scanner_getLineAndCharacterOfPosition']['findings']==[]
assert len(a['scanner_getLineAndCharacterOfPosition']['units'])==3
assert all(u['status']=='attempted' and u['checker_diagnostics']==[] for u in a['scanner_getLineAndCharacterOfPosition']['units'])
assert {u['where'].rsplit(':',2)[1] for u in a['scanner_getLineAndCharacterOfPosition']['units']}=={'549','550','551'}
assert a['new_overload_findings']==[]
assert [(f['kind'],f['reason']) for f in a['scanner_getLineStarts']['findings']]==[('NotYet','an assignment value to a member'),('Boundary','failed statement; state rolled back; continued at next statement or declaration')]
assert a['cohere_summary']['checked']==77 and a['cohere_summary']['rules']==40 and not a['cohere_summary']['gaps']['ruleCrashes']
assert a['cohere_delta']['added']==[] and len(a['cohere_delta']['removed'])==24
assert all(f['rule']=='adamic/no-optional-widening' for f in a['cohere_delta']['removed'])
assert a['lowering_delta']['added']==[] and len(a['lowering_delta']['removed'])==22
assert a['stock']['whole_src_diagnostics']==0 and a['stock']['services_diagnostics']==0 and a['stock']['services_TS2386']==0
assert a['stock_exact']['whole_src_diagnostics']==416 and a['stock_exact']['services_TS2386']==0
assert a['native_checker_errors']==6851 and a['native_checker_delta']['added']==a['native_checker_delta']['removed']==[]
assert a['stock_exact_delta']['added']==a['stock_exact_delta']['removed']==[]
assert a['first_stop']['kind']=='Refused' and a['first_stop']['where'].endswith('/tracing.ts:232:34') and a['first_stop']['reason']=="a cast the runtime can't check"
assert a['api']['normalized_changes']==0 and a['api']['diagnostics']==[]
assert (p/'e/typescript-public.diff').read_bytes()==b'' and load(p/'e/public-diff.status.json')['child_exit']==0
original=p/'original/typescript.d.ts';old_bytes=original.read_bytes() if original.exists() else gzip.decompress(pathlib.Path(str(original)+'.gz').read_bytes())
new=p/'e/typescript.d.ts';new_bytes=new.read_bytes() if new.exists() else gzip.decompress(pathlib.Path(str(new)+'.gz').read_bytes())
assert old_bytes==new_bytes
assert a['lane']['errors']==[]
assert a['sdk']['checker']['accepted'] and a['sdk']['checker']['diagnostics']==[] and a['sdk']['cohere_findings']==[]
assert a['sdk']['guarded_delta']['added']==a['sdk']['guarded_delta']['removed']==[]
assert not any(f['kind']=='Refused' for f in a['sdk']['guarded_findings'])
for label in ['original-eslint','e']:
 lint=load(p/label/'eslint-scanner.json');assert sum(r['errorCount'] for r in lint['results'])==10
 assert not [m for r in lint['results'] for m in r['messages'] if m.get('ruleId')=='@typescript-eslint/unified-signatures' or m.get('fatal')]
assert not list(out.rglob('*.go'))
text=(out/'REPORT.md').read_text();assert text.count('## Fourth measurement: internal LineMapSource overload, variant (e)')==1 and '\u2013' not in text and '\u2014' not in text
result={'audit':'passed','variant':'e','complete_full_file_streams':10,'original_views':34,'cleared':22,'unchanged':12,'replacement':0,'checker_at_original_view':0,'native_internal_call_resolutions':20,'requested_internal_call_resolutions':18,'all_requested_calls_choose_internal_overload':True,'getLineAndCharacterOfPosition_overload_unit_findings':0,'new_overload_findings':0,'cohere_added':0,'cohere_removed':24,'guarded_added':0,'guarded_removed':22,'stock_whole_src_errors':0,'services_TS2386':0,'production_whole_src_errors':6851,'new_production_checker_errors':0,'public_api_byte_equal_to_original':True,'public_api_normalized_changes':0,'lane_api_unsanctioned_changes':0,'sdk_new_findings':0,'eslint_suppression_added':True}
(p/'e-audit-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
