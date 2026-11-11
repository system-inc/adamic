import pathlib,json,collections,gzip
out=pathlib.Path(__file__).resolve().parents[1];f=out/'continuation'
def load(name):return json.loads((f/name).read_text())
def code(value):return '`'+str(value).replace('|','\\|').replace('\n',' ')+'`'
def quote(value):return '\n````json\n'+json.dumps(value,indent=2,ensure_ascii=True)+'\n````\n'
base=load('original/summary.json');summaries=load('variant-summary.json');comparison=load('site-comparison.json');public=load('public-comparison.json');calls=load('original/calls.json');changed={label:load(label+'/calls.json') for label in ['a','b']}
key=lambda c:(c['file'],c['line'],c['column'],c['callee'])
lookup={label:{key(c):c for c in rows} for label,rows in changed.items()}
assert all(set(lookup[label])=={key(c) for c in calls} for label in lookup)
coverage=load('original/src-caller-coverage.json');assert not coverage['uncovered_function_references']
retry=[]
for name in ['checker','emitter','parser','program','scanner','utilities']:retry.append(load('baseline/raw/'+name+'.status.json'))
lines=['## Follow-up: measured road 2 variants','', 'The follow-up uses the same main pin, upstream pin, frozen adapted input, and guarded binary as the initial report. The six old cut-offs were retried first, sequentially, with a 1500-second bound each. Only after all six attempts and their checkpoints completed were the variants applied or measured.','', '### Six longer baseline attempts','', '| File | Child exit | Seconds | Complete file result | Completed timing records | Last completed timed unit |','| --- | ---: | ---: | --- | ---: | --- |']
for s in retry:lines.append('| '+s['file']+' | '+str(s['child_exit'])+' | '+f'{s["wall_seconds"]:.2f}'+' | '+('yes' if s['child_exit']==0 else 'no')+' | '+str(s['completed_timing_units'])+' | '+(s['last_timing_units'][-1]['where'].split('/src/compiler/')[-1] if s['last_timing_units'] else 'none')+' |')
incomplete=[s['file'] for s in retry if s['child_exit']!=0]
lines+=['',('**The count remains a lower bound because these files still did not finish in 25 minutes: '+', '.join(incomplete)+'.** Their timing progress is recorded above and their partial streams are retained. A partial stream without an emitted file result does not mean zero sites.' if incomplete else 'All 79 baseline file streams now finish. The census remains observational at recorded internal body boundaries, but no file is omitted by a time cutoff.'),'', 'Successful longer streams supersede their first-pass streams in the primary count tables. Both sets of raw streams remain in the evidence. The complete function/body inventory is independent of whether a file census finishes.','', '### Exact variants and measured outcomes','', '(a) changes the first parameter of both helpers to `{ readonly text: string; lineMap?: readonly number[] }`.','', '(b) changes it to `{ readonly text: string; lineMap: readonly number[] | undefined }`.','', 'Each scratch tree was produced with the pinned apply script/product before replacing those two annotations. Full src-file hashes prove that scanner.ts is the only changed source file; byte comparisons prove that its only edits are the two parameter types. No SourceFile or services interface was changed. SourceFile.lineMap in this adapted tree is required and has the exact type `readonly number[] | undefined` (compiler/types.ts:4436). SourceFileLike.lineMap is optional.','', 'Adamic enables exactOptionalPropertyTypes (internal/load/load.go:63; load-options.go.txt records the options). Its (a) core check reports TS2379 because the adapted required SourceFile.lineMap can be undefined while (a) does not explicitly permit undefined when the optional property is present. The stock whole-src checks below retain the repository project defaults, where exactOptionalPropertyTypes is false. Thus a zero stock (a) error count does not imply that (a) passes Adamic. The core records and guarded-census diagnostics use the actual Adamic settings.','', 'The variants use separate single-file processes, at most one process for each variant concurrently. Each retains all 79 compiler roots and uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and an exact LATENT_ONLY_FILE. Each variant file has the same 1500-second safety bound. Checkpoints follow every completed pair of file attempts.','', '| Variant | Completed file streams | Baseline view locations | Unchanged refusal at same view | Replacement finding at same view | Checker diagnostic at same call | No finding there, unit attempted | Other / incomplete | Whole-src stock errors | Services TS2386 | Core first stop |','| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |']
for summary in summaries:
 rows=[r for r in comparison if r['variant']==summary['label']];unchanged=replaced=checked=clear=other=0
 for r in rows:
  old={(x['kind'],x['reason']) for x in r['baseline_findings']}
  if not r['complete_stream']:other+=1
  elif r['checker_diagnostics'] or r['go_checker_diagnostics']:checked+=1
  elif r['findings']:
   if any((x['kind'],x['reason']) not in old for x in r['findings']):replaced+=1
   else:unchanged+=1
  elif r['verdict']=='no finding at original call; unit attempted':clear+=1
  else:other+=1
 summary['measured_site_counts']=dict(unchanged=unchanged,replaced=replaced,checker=checked,no_finding=clear,other=other)
 stop=summary['first_stop'];where=stop.get('where','').split('/src/compiler/')[-1]
 label=stop['kind']+(' at '+where if where else '')
 lines.append(f'| ({summary["label"]}) | {summary["complete_streams"]}/{summary["census_files"]} | {len(rows)} | {unchanged} | {replaced} | {checked} | {clear} | {other} | {summary["whole_src_diagnostics"]} | {summary["services_TS2386"]} | {label} |')
lines+=['','These columns report observations at the original views, not successful lowering. A replacement finding is a different kind or reason at the same call/argument span. Unchanged findings include unmodified SourceFileLike callees. A call with two baseline SourceFileLike arguments is compared by argument span so findings are not attributed twice. A missing finding is not counted as a clearance if the file stream is incomplete or its enclosing unit is excluded. Other blockers elsewhere remain in the complete raw streams.','', 'The original core.ts first stop and both variant first stops follow. The ordinary census driver emits a per-root record and then its one-root whole-program record; the first record is the requested per-root first stop. It never invokes a backend.']
for label in ['original','a','b']:
 p=f/label/'core-first-stop.jsonl';rows=[json.loads(line) for line in p.read_text().splitlines()];lines+=['','Core root '+label+':',quote(rows[0])]
lines+=['','### Every original view under each variant','', 'The baseline table above names every observed original SourceFileLike view; the checker cutoff remains unknown. The blocks below quote every distinct finding matching each original call/argument span, including reasons that no longer name SourceFileLike. Checker diagnostics at those calls are quoted too. Unit status and stream completeness distinguish an unreported view from an excluded or unfinished measurement. Finding identities are deduplicated by kind, where, reason, text, phase, and unit; the raw streams retain repeated observations.']
for label in ['a','b']:
 lines+=['','#### Variant ('+label+')']
 for r in [r for r in comparison if r['variant']==label]:
  lines+=['','**'+r['location']+'**, '+code(r['callee'])+': '+r['verdict']+'.']
  if r['findings']:
   for finding in r['findings']:lines.append(quote(finding))
  else:lines+=['','No census finding was emitted at this original view.']
  if r['checker_diagnostics']:
   for d in r['checker_diagnostics']:lines.append(quote(d))
  if r['go_checker_diagnostics']:
   lines+=['','The guarded census checker recorded:']
   for d in r['go_checker_diagnostics']:lines.append(quote(d))
  unit_summary=[{k:u[k] for k in ['where','name','status','checker_diagnostics'] if k in u} for u in r['unit_statuses']]
  lines+=['','Stream complete: '+str(r['complete_stream']).lower()+'. Enclosing unit records:'+quote(unit_summary)]
lines+=['','### Every src call and its argument type','', f'The stock TypeScript {"6.0.3"} checker loaded all {base["whole_src_roots"]} source roots from the 13 src solution projects. A separate syntax scan covered {coverage["all_source_files"]} src source/declaration files: {coverage["loaded_source_files"]} were loaded, and the remaining library declarations have no references to either helper. There are '+str(len(calls))+' resolved calls to the two scanner functions. The 61 same-named method calls have different resolved signatures and are listed separately in excluded-or-unresolved-calls.json. No helper reference remains outside the checked program.','', 'External dependencies came from a separate lockfile installation, with stock module resolution falling back to that installation. No node_modules directory was added to the frozen census tree. Whole-src checking uses strict source checks on the union of solution roots; composite, project-output, and isolated-declaration restrictions are disabled. Original whole-src diagnostics: '+str(base['whole_src_diagnostics'])+'. Full diagnostics and exact options are retained for each variant.','', 'The latent census loads the compiler-only 79-root program. This stock call audit intentionally loads services too, including its SourceFile augmentation. Declared type means the variable/parameter/property declaration, declared return type for a call expression, or explicit/class/contextual this type. Flow type is recorded separately because a Node or optional SourceFile binding may be narrowed at the call. Exact declarations and variant types are in each calls.json.','', '| Call location | Callee | Argument | Original declared type | Original flow type | (b) call errors |','| --- | --- | --- | --- | --- | --- |']
for c in sorted(calls,key=key):
 b=lookup['b'][key(c)];diags=b['diagnostics'];lines.append('| '+c['file']+':'+str(c['line'])+':'+str(c['column'])+' | '+code(c['callee'])+' | '+code(c['argument'])+' | '+code(c['declared']['type'])+' | '+code(c['flow_type'])+' | '+(', '.join('TS'+str(d['code']) for d in diags) or 'none')+' |')
changes=[]
for label in ['a','b']:
 for c in calls:
  n=lookup[label][key(c)]
  if n['declared']['type']!=c['declared']['type']:changes.append(dict(variant=label,file=c['file'],line=c['line'],callee=c['callee'],original=c['declared']['type'],variant_declared=n['declared']['type']))
lines+=['','Declared argument types changed by the edited helper parameter itself (all other entries retain the original declared type):',quote(changes),'', '### Every failing call under (b)','', 'These are actual stock TypeScript diagnostics from the whole-src program, not a classification based only on the spelling SourceFileLike. This preserves flow narrowing and class/contextual this types. The whole diagnostic stream, including any diagnostics outside calls, is retained in b/whole-src-diagnostics.json.']
failing=[]
for c in sorted(changed['b'],key=key):
 if not c['diagnostics']:continue
 failing.append(c);lines+=['','**'+c['file']+':'+str(c['line'])+':'+str(c['column'])+'** '+code(c['callee'])+' argument '+code(c['argument'])+', declared '+code(c['declared']['type'])+', flow '+code(c['flow_type'])+'.']
 for d in c['diagnostics']:lines.append(quote(d))
lines+=['',str(len(failing))+' calls have diagnostics under (b). The corresponding whole-source and services summaries are:']
for summary in summaries:lines.append(quote({k:summary[k] for k in ['label','whole_src_roots','whole_src_diagnostics','calls','services_files','services_diagnostics','services_TS2386']}))
lines+=['','### Emitted public declarations with stripInternal','', 'All three public bundles were built from actual stock TypeScript declaration emit with stripInternal=true and emitDeclarationOnly=true, then the upstream declaration bundler was run on the emitted .d.ts entry. No JavaScript or backend output was produced. Emission is allowed despite semantic errors so that (b) can be compared; its source diagnostics above are not suppressed. The exact options, emit diagnostics, emitted-file hashes, public bundles, and textual diffs are retained.','', '| Variant | Free getLineStarts in public bundle | Only getLineAndCharacterOfPosition parameter differs | Public SourceFile has lineMap | Public SourceFile caller errors |','| --- | ---: | --- | --- | ---: |']
for p in public:lines.append('| '+p['label']+' | '+str(p['free_getLineStarts_declarations'])+' | '+str(p['only_getLineAndCharacterOfPosition_parameter_differs']).lower()+' | '+str(p['public_SourceFile_has_lineMap']).lower()+' | '+str(len(p['public_SourceFile_consumer_diagnostics']))+' |')
lines+=['','getLineStarts is @internal, so changing its parameter contributes no free-function declaration to public typescript.d.ts. The public API comparison normalizes only getLineAndCharacterOfPosition\'s first parameter back to SourceFileLike, then prints and compares the complete declaration AST. Equality confirms that this parameter is the sole public declaration difference, independent of formatting. The direct diffs and emitted signatures are retained.']
for label in ['original','a','b']:lines+=['','Public signature '+label+':',quote(load(label+'/public-signatures.json')['functions'])]
for p in public:
 if p['public_SourceFile_consumer_diagnostics']:lines+=['','The emitted-public-API caller probe for '+p['label']+' reports:',quote(p['public_SourceFile_consumer_diagnostics'])]
lines+=['','The public caller probe uses the emitted ts.SourceFile type, not the internal source interface. Its lineMap property can disappear under stripInternal even though the internal SourceFile member is required. This distinction is separate from the internal src caller audit.','', 'Follow-up evidence is under continuation/. Original first-pass raw files remain under raw/; longer main retries are under continuation/baseline/raw/, and variant streams are under continuation/a/raw/ and continuation/b/raw/. All started measurement children are waited for, with process-group termination at their bound. The production compiler and main adaptations are unchanged; the only TypeScript source edits are in the two scratch variants.','']
text='\n'.join(lines);assert '\u2013' not in text and '\u2014' not in text
(f/'REPORT-ADDITION.md').write_text(text);(f/'variant-summary.json').write_text(json.dumps(summaries,indent=2)+'\n');print('follow-up report generated with '+str(len(calls))+' calls and '+str(len(comparison))+' site/variant rows')
