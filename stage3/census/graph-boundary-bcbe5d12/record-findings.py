"""Classify selected stopped-unit findings without replacing the IR census."""
import collections,gzip,hashlib,json,re,sys
from pathlib import Path
here=Path(__file__).resolve().parent;raw=here/'raw';scratch=Path('/workspace/scratch/jawqwv7')
def read(p):
 with gzip.open(p,'rt') as f:return [json.loads(l) for l in f]
def rel(p):return p.replace('/workspace/graph-boundary-adapted/src/compiler/','').replace('/workspace/graph-boundary-adapted/src/','')
chosen={};original_lookup={}
for p in [raw/'first-boundary.jsonl.gz',*sorted(raw.glob('full-*-boundary.jsonl.gz'))]:
 for r in read(p):chosen[r['file'],r['unit']]=(p.name.replace('-boundary','-latent'),r)
for p in [raw/'first-latent.jsonl.gz',*sorted(raw.glob('full-*-latent.jsonl.gz'))]:
 for row in read(p):
  for f in row.get('findings',[]):
   if f['phase']=='lowering':original_lookup.setdefault((p.name,row['file'],f['unit']),[]).append(f)
stopped={key:value for key,value in chosen.items() if value[1]['status']=='stopped'}
assert len(stopped)==3666
missing={key:value for key,value in stopped.items() if (value[0],*key) not in original_lookup}
assert len(missing)==904
plan=json.loads((here/'record-recovery-plan.json').read_text());recovered={};recovery_sources={};recovery_checks=[]
for job in plan:
 p=scratch/(job['name']+'-findings.jsonl');archived=raw/(p.name+'.gz')
 if not p.exists() and not archived.exists():continue
 recovery_rows=read(archived) if archived.exists() else [json.loads(line) for line in p.open()]
 for r in recovery_rows:
  key=(r['file'],r['unit']);assert key in missing
  original=missing[key][1]
  # Unit selection does not alter registration or independent-unit state. Require
  # identical emitted boundary metadata, counts and retained IR site stream.
  for field in ['status','full_mode','graph_types_count','graph_types_nonempty','sites','findings_added']:
   assert r[field]==original[field],(key,field)
  seen=set();fs=[]
  for f in r.get('recovered_findings') or []:
   assert f['phase']=='lowering' and f['unit']==key[1],(key,f)
   identity=json.dumps(f,sort_keys=True)
   if identity not in seen:seen.add(identity);fs.append(f)
  recovered[key]=fs;recovery_sources[key]=p.name+'.gz'
  recovery_checks.append({'file':key[0],'unit':key[1],'recovery_stream':p.name,'original_stream':missing[key][0],'boundary_metadata_and_sites_match':True,'findings_before_deduplication':len(r.get('recovered_findings') or [])})
all_findings=[];missing_now=[]
for key,(stream,row) in stopped.items():
 fs=original_lookup.get((stream,*key))
 if fs is None:fs=recovered.get(key)
 if fs is None:missing_now.append({'file':key[0],'unit':key[1],'original_stream':stream});continue
 assert fs,(key,'stopped without a finding')
 for i,f in enumerate(fs):all_findings.append({**f,'owner_file':key[0],'owner_unit':key[1],'finding_index':i,'source_stream':stream,'finding_payload_stream':recovery_sources.get(key,stream),'boundary_source_stream':stream.replace('-latent','-boundary'),'recovered':key in recovered})
where=sorted({f['where'] for f in all_findings if f['where']})
(scratch/'record-all-query-sites.json').write_text(json.dumps(where)+'\n')
if '--prepare' in sys.argv:
 print(json.dumps({'stopped_units':len(stopped),'original_payload_units':len(stopped)-len(missing),'recovered_units':len(recovered),'missing_units':len(missing_now),'findings':len(all_findings),'sites':len(where)}));sys.exit(0)
inspection_path=raw/'record-all-source-classification.json.gz'
inspection=json.loads(gzip.decompress(inspection_path.read_bytes())) if inspection_path.exists() else json.loads((scratch/'record-all-source-classification.json').read_text());
if '--source-inspection' in sys.argv:
 selected_inspection=Path(sys.argv[sys.argv.index('--source-inspection')+1]);inspection=json.loads(gzip.decompress(selected_inspection.read_bytes()) if selected_inspection.suffix=='.gz' else selected_inspection.read_text())
assert inspection['typescript_version']=='6.0.3'
assert all(w in inspection['sites'] for w in where)
assert not missing_now,missing_now
# Classification rules apply per occurrence, never to an entire mixed reason.
# Wrapper metadata has its own exact-reason group, but is not an extra cause.
bookkeeping={'Boundary'}
def explicit(reason):
 if re.search(r'\brecord\b|\bdictionary\b|\bindex[- ]signature\b|\bMapLike\b|\bRecord\s*<|dynamic[- ]shape object',reason,re.I):
  return 'Reason explicitly describes a record, dictionary or index-signature object.'
 if reason.startswith('a checked predicate target') and 'CompilerOptions' in reason:
  return 'Predicate target is CompilerOptions, whose types.ts declaration has a string index signature.'
 if reason.startswith('overload ') and 'convertOptionsFromJson result ' in reason:
  return 'Result types CompilerOptions, TypeAcquisition and WatchOptions all declare string index signatures in types.ts.'
 return None
for f in all_findings:
 evidence=inspection['sites'].get(f['where'],{}).get('matches',[])
 f['source_record_evidence']=evidence
 if f['kind'] in bookkeeping:
  f['record_shaped']=bool(evidence);f['classification_basis']='Continuation bookkeeping, not an independent failure cause; marked only when the source checker verifies record-shaped context.'
 else:
  basis=explicit(f['reason'])
  if basis:f['record_shaped']=True;f['classification_basis']=basis
  elif evidence:f['record_shaped']=True;f['classification_basis']='Source checker verifies '+ '; '.join(sorted({x['basis'] for x in evidence}))+'.'
  else:
   f['record_shaped']=False
   if 'process.recordreplay' in f['reason']:decision='process.recordreplay is a process operation name, not a record object.'
   elif f['reason']=='a Map of CompilerOptionsValue':decision='CompilerOptionsValue includes MapLike<string[]> in types.ts:7412, but this Map representation reason does not identify the failing union member; record involvement is uncertain.'
   elif f['reason'].startswith('a Map ') or f['reason'].startswith('a Map or Set '):decision='Map/Set representation alone is not a dictionary or record object; no verified record source context.'
   elif 'CompilerOptions' in f['reason']:decision='CompilerOptions is mentioned in another feature signature, but record-shaped causality is not established at this site.'
   elif f['reason'].startswith('reading '):decision='Generic read reason; no verified index-signature object operation or typed record literal at this site.'
   else:decision='Reason and source inspection do not establish a record-shaped failure.'
   f['classification_basis']='Unmarked: '+decision+' Uncertain cases remain unmarked.'
byreason=collections.defaultdict(list);byunit=collections.defaultdict(list)
for f in all_findings:byreason[f['reason']].append(f);byunit[f['owner_file'],f['owner_unit']].append(f)
def snippet(where):
 if not where:return None
 p,l,c=where.rsplit(':',2);a=Path(p).read_text().splitlines();line=int(l)
 return {'where':where,'source_start_line':max(1,line-1),'source_text':'\n'.join(a[max(0,line-2):min(len(a),line+3)])}
groups=[]
for reason,fs in sorted(byreason.items(),key=lambda kv:(-len(kv[1]),kv[0])):
 marked=[f for f in fs if f['record_shaped']];distinct=sorted({f['where'] for f in fs if f['where']});md=sorted({f['where'] for f in marked if f['where']})
 files=collections.Counter(f['where'].rsplit(':',2)[0] if f['where'] else f['owner_file'] for f in marked)
 decisions=sorted({f['classification_basis'] for f in fs})
 groups.append({'reason':reason,'count':len(fs),'distinct_locations':len(distinct),'findings_without_location':sum(not f['where'] for f in fs),'locations':distinct,'units':len({(f['owner_file'],f['owner_unit']) for f in fs}),'finding_kinds':dict(collections.Counter(f['kind'] for f in fs)),'record_shaped':bool(marked),'record_shaped_scope':'all' if len(marked)==len(fs) else 'some' if marked else 'unmarked','record_shaped_count':len(marked),'record_shaped_distinct_sites':len(md),'record_shaped_locations':md,'decision':decisions,'record_context_bookkeeping_count':sum(f['kind'] in bookkeeping and bool(f['source_record_evidence']) for f in fs),'top_files':[{'file':file,'count':n,'distinct_sites':len({f['where'] for f in marked if f['where'].rsplit(':',2)[0]==file})} for file,n in files.most_common(5)],'examples':[{**snippet(w),'classification_basis':next(f['classification_basis'] for f in marked if f['where']==w),'source_type_evidence':inspection['sites'].get(w,{}).get('matches',[])} for w in md[:2]]})
first=[];behind=[];none=[];causal_first=[];causal_behind=[];causal_none=[];bookkeeping_only=[]
for key,fs in byunit.items():
 causal=[f for f in fs if f['kind'] not in bookkeeping]
 assert causal and fs[0]['kind'] not in bookkeeping,(key,'first finding is bookkeeping')
 if fs[0]['record_shaped']:first.append(key)
 elif any(f['record_shaped'] for f in fs[1:]):behind.append(key)
 else:none.append(key)
 if causal[0]['record_shaped']:causal_first.append(key)
 elif any(f['record_shaped'] for f in causal[1:]):causal_behind.append(key)
 else:causal_none.append(key)
 if any(f['record_shaped'] for f in fs) and not any(f['record_shaped'] for f in causal):bookkeeping_only.append(key)
marked=[f for f in all_findings if f['record_shaped']]
summary={'stopped_units':len(stopped),'units_with_original_payloads':len(stopped)-len(missing),'missing_original_payloads':len(missing),'recovered_payloads':len(recovered),'unresolved_payloads':len(missing_now),'findings_including_bookkeeping':len(all_findings),'findings_without_source_location':sum(not f['where'] for f in all_findings),'causal_findings':sum(f['kind'] not in bookkeeping for f in all_findings),'exact_reason_groups':len(groups),'record_shaped_reason_groups':sum(g['record_shaped_count']>0 for g in groups),'record_shaped_causal_reason_groups':len({f['reason'] for f in all_findings if f['record_shaped'] and f['kind'] not in bookkeeping}),'record_shaped_findings':len(marked),'record_shaped_causal_findings':sum(f['kind'] not in bookkeeping for f in marked),'record_shaped_causal_distinct_sites':len({f['where'] for f in marked if f['kind'] not in bookkeeping and f['where']}),'record_shaped_distinct_sites':len({f['where'] for f in marked if f['where']}),'record_shaped_first_stop_units':len(first),'record_shaped_only_behind_another_stop_units':len(behind),'record_shaped_any_units':len({(f['owner_file'],f['owner_unit']) for f in marked}),'record_shaped_causal_units':len(causal_first)+len(causal_behind),'record_shaped_causal_first_stop_units':len(causal_first),'record_shaped_causal_only_behind_another_stop_units':len(causal_behind),'bookkeeping_only_record_context_units':len(bookkeeping_only),'no_verified_record_shaped_causal_units':len(causal_none),'no_verified_record_shaped_units':len(none),'bookkeeping_in_record_context':sum(f['kind'] in bookkeeping and bool(f['source_record_evidence']) for f in all_findings)}
assert len(first)+len(behind)+len(none)==3666
recovery_runs=[]
for job in plan:
 status_file=raw/(job['name']+'-status.json.gz')
 status=json.loads(gzip.decompress(status_file.read_bytes())) if status_file.exists() else json.loads((scratch/(job['name']+'-status.json')).read_text())
 recovery_runs.append({'name':job['name'],'mode':'full' if job['full_mode'] else 'first-error','target_payloads':len(job['units']),**status})
result={'summary':summary,'recovery_runs':recovery_runs,'method':{'selection':'Identical (file,unit) selection as four-way census: latest full observation wins, otherwise first-error. Only phase=lowering findings of selected stopped units. Per-file refusal_scan findings do not stop unit lowering and are excluded.','order':'Preserved emission order within the selected unit; first stop is first non-Boundary finding, not earliest source line. Overall behind counts include later marked Boundary context; causal behind counts exclude Boundary entries.','deduplication':'Use each original latent stream as emitted; recovery uses the same all-fields exact deduplication as latentFullSelected. Repeat dependency failures across different owner units remain separate observations. Distinct sites deduplicate actual finding where, not unit fallback.','classification':'Explicit record/dictionary reasons or verified source index-signature object operations/literals. Generic reasons are classified occurrence by occurrence. A matching source read/write is context, not proof records are the only missing compiler feature. Primitive, array, tuple and Map/Set operations alone are excluded. Boundary groups are classified from source context and included in overall finding totals, but excluded from causal totals and first-stop ordering.','coverage':'All 3666 selected stopped units have finding payloads. This is findings coverage, not completed compiler lowering; rollback, skipped/deferred units and interrupted candidate traversal gaps remain. No census counts are replaced.'},'groups':groups,'units':{'first_stop':[{'file':f,'unit':u} for f,u in first],'only_behind_another_stop':[{'file':f,'unit':u} for f,u in behind],'causal_first_stop':[{'file':f,'unit':u} for f,u in causal_first],'causal_only_behind_another_stop':[{'file':f,'unit':u} for f,u in causal_behind],'only_bookkeeping_record_context':[{'file':f,'unit':u} for f,u in bookkeeping_only]},'recovery_validation':recovery_checks}
(here/'record-findings.json').write_text(json.dumps(result,indent=2)+'\n')
with gzip.GzipFile(filename=str(here/'record-findings-observations.jsonl.gz'),mode='wb',mtime=0) as out:
 for f in all_findings:out.write((json.dumps(f,separators=(',',':'))+'\n').encode())
heading='## Record-shaped findings among stopped units'
report=(here/'REPORT.md').read_text().split('\n'+heading)[0].rstrip()
lines=['',heading,'',f"All {summary['stopped_units']:,} stopped unit results are covered: {summary['units_with_original_payloads']:,} already had finding payloads; {summary['recovered_payloads']:,} payloads missing from interrupted passes were recovered. The inherited latent driver buffers file records until pass completion, while the boundary observer streams each unit. Available finding payloads were not rerun; missing-payload recovery does not replace any four-way IR census row. Each recovery uses only the originally observed units, the same mode, unchanged registration and lowering, and an observer that flushes findings per unit. Every recovered boundary status, findings-added count, GraphTypes flag and retained IR site list matches its original row. Recovery children are capped at 1490 seconds and their own exits are preserved.",'',f"**Record-shaped totals: {summary['record_shaped_findings']:,} finding observations, {summary['record_shaped_distinct_sites']:,} distinct finding locations, in {summary['record_shaped_reason_groups']} exact-reason groups and {summary['record_shaped_any_units']:,} stopped units. Of these, {summary['record_shaped_causal_findings']:,} findings at {summary['record_shaped_causal_distinct_sites']:,} locations are causal failures; the rest are bookkeeping in record-shaped context. {summary['record_shaped_first_stop_units']:,} units have a record-shaped first stop; {summary['record_shaped_only_behind_another_stop_units']:,} have one only behind another stop.** Excluding bookkeeping, the causal split is {summary['record_shaped_causal_first_stop_units']} first-stop units and {summary['record_shaped_causal_only_behind_another_stop_units']} only-behind units. {summary['bookkeeping_only_record_context_units']} additional units have record-shaped context only in bookkeeping, with no verified record-shaped cause. {summary['no_verified_record_shaped_causal_units']:,} units have no verified record-shaped causal finding.",'',f"Grouped {summary['findings_including_bookkeeping']:,} emitted findings into {summary['exact_reason_groups']} exact reason texts, including {summary['findings_including_bookkeeping']-summary['causal_findings']:,} Boundary bookkeeping entries. These wrappers describe rollback or continuation and do not add an independent cause; {summary['bookkeeping_in_record_context']} have source record context. They are included in overall record-shaped finding totals when the source context qualifies, but excluded from causal totals and first-stop ordering. Refusal-scan findings are excluded because they are file-level diagnostics, not the findings that stopped these unit results.",'',f"Findings without a source location: {summary['findings_without_source_location']}. They remain in reason counts and contribute no distinct site; classification there can use explicit reason evidence only.",'','Selection and ordering follow the existing census: one observation per (owner file, unit), latest full mode preferred, otherwise first error. First means first causal finding in emission order, not the smallest source line. The overall only-behind count includes marked bookkeeping that follows a non-record first stop; the separate causal only-behind count excludes it. Repeated dependency failures in different units count separately. Distinct sites use the finding’s actual file:line:column, unlike the four-way IR table’s unit fallbacks. Top files below use the finding location’s file, not its owner unit; both are retained in the observation stream.','',"Classification uses exact reason text and source inspection with pinned TypeScript 6.0.3. Record/dictionary/index-signature reasons are marked. Generic reasons are marked only at sites where the source checker verifies an index-signature object literal (including a member initializer, without crossing a function boundary) or a string-keyed operation on such an object. CompilerOptions, WatchOptions and TypeAcquisition declare string index signatures in types.ts. A mixed reason does not make all its occurrences record-shaped. Source context establishes a record-shaped operation, not that record support alone would fix it. Arrays, tuples, primitives, Maps and Sets do not qualify by indexing alone. Uncertain cases remain unmarked; each group's decision is in the table and JSON.",'','Coverage: all stopped-unit payloads, not all source programs or failed expressions. Failed statements roll back; later statements may be visited but expressions behind an internal first failure can remain hidden. Deferred generics, checker skips and interrupted unit traversal remain the same gaps described above. These findings do not imply GraphTypes membership or native lowering success. These totals describe failed-lowering evidence; the four-way table continues to count retained IR only.','',"Machine-readable exact groups, decisions, locations, top files and examples: [record-findings.json](record-findings.json). Per-observation classification and provenance: [record-findings-observations.jsonl.gz](record-findings-observations.jsonl.gz). No reason text is normalized; Markdown escapes only table syntax.",'','### Every exact finding reason','','| Exact reason | Findings | Distinct locations | Record-shaped findings | Record sites | Decision |','| --- | ---: | ---: | ---: | ---: | --- |']
def escape(s):return s.replace('&','&amp;').replace('<','&lt;').replace('>','&gt;').replace('|','\\|').replace('\n','<br>')
for g in groups:lines.append('| '+escape(g['reason'])+' | '+str(g['count'])+' | '+str(g['distinct_locations'])+' | '+str(g['record_shaped_count'])+' | '+str(g['record_shaped_distinct_sites'])+' | '+escape(g['record_shaped_scope']+': '+' '.join(g['decision']))+' |')
lines+=['','### Missing-stream recovery','','The initial first-error recovery ignored LATENT_ONLY_FILE in its inherited driver and scanned unrelated files. It was explicitly stopped, then retried with a scratch file filter; unit registration and lowering were unchanged. All previously completed payloads were reused.','', '| Recovery piece | Mode | Target payloads | Child exit | Seconds | Stopping reason |','| --- | --- | ---: | ---: | ---: | --- |']
for run in recovery_runs:lines.append('| '+run['name']+' | '+run['mode']+' | '+str(run['target_payloads'])+' | '+str(run['exit_code'])+' | '+str(round(run['seconds'],3))+' | '+escape(run['stopping_reason'] or 'completed')+' |')
lines+=['','### Record-shaped groups, top files and examples']
for g in groups:
 if not g['record_shaped_count']:continue
 lines+=['','#### '+escape(g['reason']),'',f"{g['record_shaped_count']} record-shaped findings at {g['record_shaped_distinct_sites']} distinct sites, out of {g['count']} findings in this reason group.",'','Decision: '+' '.join(g['decision']),'','Top finding files: '+ '; '.join(rel(f['file'])+': '+str(f['count'])+' findings, '+str(f['distinct_sites'])+' sites' for f in g['top_files'])+'.']
 for ex in g['examples']:lines+=['','`'+rel(ex['where'])+'` (source starts at line '+str(ex['source_start_line'])+'):','```ts',ex['source_text'],'```']
 if len(g['examples'])==1:lines+=['','Only one distinct site exists in this group; a second distinct example is unavailable.']
 if not g['examples']:lines+=['','No source position was recorded for this group; source examples are unavailable.']
lines+=['','Reproduce source classification and this section without lowering:','```sh','python3 stage3/census/graph-boundary-bcbe5d12/record-findings.py --prepare','node stage3/census/graph-boundary-bcbe5d12/classify-record-sites.cjs /workspace/scratch/jawqwv7/record-all-query-sites.json /workspace/scratch/jawqwv7/record-all-source-classification.json','python3 stage3/census/graph-boundary-bcbe5d12/record-findings.py --source-inspection /workspace/scratch/jawqwv7/record-all-source-classification.json','```','']
(here/'REPORT.md').write_text(report+'\n'+'\n'.join(lines))
print(json.dumps(summary))
