"""Archive finished passes, recount raw pairs and publish a coverage-qualified report."""
import collections,gzip,hashlib,json,os,re,socket,subprocess,sys
from pathlib import Path
here=Path(__file__).resolve().parent;root=here.parents[2];scratch=Path('/workspace/scratch/vqzty8t');source=Path('/workspace/view-pairs-adapted');raw=here/'raw';raw.mkdir(exist_ok=True)
def git(*args):return subprocess.check_output(['git',*args],cwd=root,text=True).strip()
def read(path):
 if not path.exists():return []
 data=gzip.decompress(path.read_bytes()).decode() if path.suffix=='.gz' else path.read_text()
 return [json.loads(l) for l in data.splitlines() if l]
def rel(s):return s.replace(str(source/'src/compiler')+'/','').replace(str(source/'src')+'/','')
# Freeze completion markers once: a sibling finishing during this loop must
# enter the next snapshot as a whole, never as a mixture of early/late files.
finished_at_start=set()
for marker in scratch.glob('*-status.json'):
 name=marker.name.removesuffix('-status.json')
 if not re.fullmatch(r'first|full-\d+',name):continue
 try:
  status=json.loads(marker.read_text());status['exit_code']
 except (ValueError,KeyError):continue
 finished_at_start.add(name)
for p in scratch.iterdir():
 match=re.match(r'(first|full-\d+)(?:-|\.)',p.name)
 if match and match[1] not in finished_at_start:continue
 if p.is_file() and p.suffix in ('.log','.jsonl','.json'):(raw/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for p in (scratch/'overlay').rglob('*'):
 if not p.is_file():continue
 target=here/'overlay'/p.relative_to(scratch/'overlay');target=target.with_name(target.name+'.txt') if target.suffix=='.go' else target
 target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(p.read_bytes())
base=git('rev-parse','734aef5f');probe=git('rev-parse','86157bea')
provenance={'task':'vqzty8t','source_base':base,'probe_commit':probe,'probe_first_commit_parent':base,'probe_immediate_parent':git('rev-parse','86157bea^'),'adapted_tree':str(source),'typescript_pin':'050880ce59e30b356b686bd3144efe24f875ebc8','typescript_actual':subprocess.check_output(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],text=True).strip(),'box':socket.gethostname(),'cpu_quota':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'memory_limit':Path('/sys/fs/cgroup/memory.max').read_text().strip(),'binary_sha256':hashlib.sha256((scratch/'census').read_bytes()).hexdigest(),'original_probe_sha256':hashlib.sha256((here/'memory-probe-original.go.txt').read_bytes()).hexdigest(),'same_bcbe_apply':False,'branch':'typescript/view-pairs-census'}
comparison=here/'apply-comparison.json'
if comparison.exists():provenance['apply_comparison']=json.loads(comparison.read_text())
(here/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
manifest={str(p.relative_to(source)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((source/'src').rglob('*.ts'))}
(here/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
passes=[]
for p in raw.glob('*-status.json.gz'):
 name=p.name[:-15]
 if re.fullmatch(r'first|full-\d+',name):passes.append((name,json.loads(gzip.decompress(p.read_bytes()))))
passes.sort(key=lambda x:(x[0]!='first',x[0]))
chosen={};observations={};all_pairs=[];inventory={};file_inventory=set();full_files=set();headers=[];orphans=[]
for name,status in passes:
 boundaries=read(raw/(name+'-boundary.jsonl.gz'));unit_rows=read(raw/(name+'-units.jsonl.gz'));pairs=read(raw/(name+'-pairs.jsonl.gz'));latent=read(raw/(name+'-latent.jsonl.gz'))
 if latent:headers.append(latent[0])
 for item in latent:
  if 'units' in item:
   file_inventory.add(item['file'])
   if name!='first':full_files.add(item['file'])
   for u in item['units']:inventory[item['file'],u['where']]=u
 unit_map={(r['file'],r['unit']):r for r in unit_rows};byunit=collections.defaultdict(list)
 for index,r in enumerate(pairs):
  r={**r,'pass':name,'pass_index':index};byunit[r['file'],r['unit']].append(r);all_pairs.append(r)
 for key,items in byunit.items():
  if key not in unit_map:orphans.extend(items)
 for r in boundaries:
  key=(r['file'],r['unit']);chosen[key]=(name,r)
  u=unit_map.get(key);ps=byunit.get(key,[])
  if u:
   assert u['pairs_logged']==len(ps),(name,key)
   assert u['status']==r['status'] and u['full_mode']==r['full_mode']
   assert u['graph_types_count']==r['graph_types_count']
  observations[key]=(u,ps)
selected=[];incomplete=[];errors=[];unitdetails=[]
for key,(name,b) in chosen.items():
 u,ps=observations[key]
 if u is None:incomplete.append({'pass':name,'file':key[0],'unit':key[1]});continue
 if u['probe_error']:errors.append({'pass':name,**u})
 selected.extend(ps);unitdetails.append({'pass':name,**u})
assert not errors,errors
labels=sorted({r[key] for r in all_pairs for key in ['source','target'] if r[key]})
(scratch/'scalar-query-labels.json').write_text(json.dumps(labels)+'\n')
(scratch/'scalar-query-files.json').write_text(json.dumps([str(source/name) for name in manifest if '/compiler/' in '/'+name])+'\n')
scalar_run=subprocess.run(['node',str(here/'scalar-types.cjs'),str(scratch/'scalar-query-labels.json'),str(scratch/'scalar-query-files.json'),str(here/'scalar-types.json')],capture_output=True,text=True,cwd=root)
(scratch/'latest-scalar-classification.log').write_text(scalar_run.stdout+scalar_run.stderr)
print('scalar classification child_exit='+str(scalar_run.returncode),flush=True)
if scalar_run.returncode:raise RuntimeError(scalar_run.stdout+scalar_run.stderr)
scalar_doc=json.loads((here/'scalar-types.json').read_text());assert scalar_doc['typescript_version']=='6.0.3'
scalar_labels=set(scalar_doc['scalar_labels'])
# Nil sources stay a distinct category even when their target is an intersection.
def category(r):
 if r['source_nil']:return 'nil_source'
 if r['target_nil']:return 'nil_target_only'
 if r['object_pair'] and (r['target_intersection'] or (r['source'] not in scalar_labels and r['target'] not in scalar_labels)):
  related=r['intersection_related'] if r['source_intersection'] or r['target_intersection'] else r['related']
  return 'object_'+('related' if related else 'unrelated')+('_top' if r['depth']==0 else '_nested')
 return 'scalar_top' if r['depth']==0 else 'scalar_nested'
def split(rows):
 c=collections.Counter(category(r) for r in rows)
 return {k:c[k] for k in ['object_related_top','object_related_nested','object_unrelated_top','object_unrelated_nested','scalar_top','scalar_nested','nil_source','nil_target_only']}
counts=split(selected);raw_counts=split(all_pairs)
unrelated=[r for r in selected if category(r).startswith('object_unrelated')]
nils=[r for r in selected if r['source_nil']]
nil_refs=[r for r in nils if r.get('target_reference_fields')]
nil_objects=[r for r in nils if r.get('target_object_reference_fields')]
nil_unknown=[r for r in nils if not r.get('target_reference_fields') and (r.get('target_unknown_fields') or r['target_nil'])]
intersections=[r for r in selected if not r['source_nil'] and r['target_intersection']]
grouped=collections.defaultdict(list)
for r in unrelated:grouped[r['source'],r['target'],tuple(r.get('path') or [])].append(r)
pairgroups=[]
for (src,dst,path),rs in sorted(grouped.items()):
 units=collections.Counter((r['file'],r['unit']) for r in rs)
 pairgroups.append({'source':src,'target':dst,'path':list(path),'count':len(rs),'files':sorted({r['file'] for r in rs}),'units':[{'file':f,'unit':u,'count':n} for (f,u),n in sorted(units.items())]})
status_counts=collections.Counter(r['status'] for _,r in chosen.values());skips=collections.Counter(u['status'] for key,u in inventory.items() if key not in chosen)
plan_path=raw/'plan.json.gz';plan=json.loads(gzip.decompress(plan_path.read_bytes())) if plan_path.exists() else {'jobs':[]}
finished_names={name for name,_ in passes};full_finished=[j for j in plan['jobs'] if j['mode']=='full' and j['name'] in finished_names]
resolved_files=headers[0].get('files',[]) if headers else []
summary={'counts':counts,'raw_counts':raw_counts,'total_selected_pairs':len(selected),'total_raw_pairs':len(all_pairs),'unrelated_object_pairs':len(unrelated),'distinct_unrelated_text_pairs':len(pairgroups),'nil_source_views':len(nils),'nil_source_with_reference_field':len(nil_refs),'nil_source_with_object_reference_field':len(nil_objects),'nil_source_reference_status_unknown':len(nil_unknown),'intersection_targets':len(intersections),'intersection_split':split(intersections),'observed_units':len(chosen),'unit_statuses':dict(status_counts),'stock_skips_without_boundary':dict(skips),'deferred_units_in_emitted_inventory':skips['deferred_generic_instantiation'],'resolved_files':len(resolved_files),'full_files_attempted':len(full_finished),'full_files_complete_traversal':len(full_files),'stock_inventory_units':len(inventory),'probe_incomplete_units':len(incomplete),'orphan_pair_lines':len(orphans),'probe_errors':errors,'view_sources_in_selected_units':sum(u['view_sources'] for u in unitdetails),'deduplicated_pair_visits':sum(u['dedup_skipped'] for u in unitdetails),'over_depth_pair_visits':sum(u['depth_skipped'] for u in unitdetails),'nil_reference_definition':'Known lower.representation(fieldType) with ir.Type.IsReference(); includes strings and Weak. Object-reference companion excludes ir.String and ir.Weak. Fields come from cycleFinder.fields, excluding prototype methods, accessors and static prototype.','first_error_selected_stopped_units':sum(not b['full_mode'] and b['status']=='stopped' for _,b in chosen.values()),'cycle_finder_ran':False,'graph_types_nonempty_units':sum(r['graph_types_count']>0 for _,r in chosen.values())}
assert sum(counts.values())==len(selected) and sum(raw_counts.values())==len(all_pairs)
(here/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
(here/'unrelated-object-pairs.json').write_text(json.dumps(pairgroups,indent=2)+'\n')
(here/'nil-source-reference-views.json').write_text(json.dumps(nil_refs,indent=2)+'\n')
(here/'nil-source-reference-unknown.json').write_text(json.dumps(nil_unknown,indent=2)+'\n')
with gzip.GzipFile(filename=str(here/'selected-pairs.jsonl.gz'),mode='wb',mtime=0) as out:
 for r in selected:out.write((json.dumps(r,separators=(',',':'))+'\n').encode())
with gzip.GzipFile(filename=str(here/'selected-units.jsonl.gz'),mode='wb',mtime=0) as out:
 for r in unitdetails:out.write((json.dumps(r,separators=(',',':'))+'\n').encode())
(here/'probe-incomplete-units.json').write_text(json.dumps(incomplete,indent=2)+'\n')
def esc(s):return str(s).replace('&','&amp;').replace('<','&lt;').replace('>','&gt;').replace('|','\\|').replace('\n','<br>')
lines=['# TypeScript view-pairs census','','Task vqzty8t. Measurement only. Counts describe observed latent unit lowerings, not accepted whole programs.','',f"Box `{provenance['box']}`, CPU quota `{provenance['cpu_quota']}`, memory limit `{provenance['memory_limit']}` bytes. Source/compiler base `{base}`. Probe `{probe}`. TypeScript 6.0.3 pin `{provenance['typescript_pin']}`.",'','## Base and probe','',"The base is 734aef5f, the main parent before Memory's first throwaway probe commit 8bedbc18. The named tip 86157bea fixes that probe's program-name conversion. Its immediate parent already contains the probe, so it is not used as a production base. bcbe5d12 has different adaptation inputs; this run builds the keyed adapted product at 734aef5f rather than reusing bcbe's adapted bytes. apply-comparison.json records both input keys and the changed adapted source files. Prepared Go/cohere and pinned parser dependencies were reused.",'',"Memory's original probe is preserved as memory-probe-original.go.txt and copied into scratch. The scratch context version calls logViewPairs() at every unit end beside the boundary observer, using only &cycleFinder{l:l, where:map[*checker.Type]*ast.Node{}}. findCycles never runs. Original pair traversal, per-unit checker-pointer seen set, first encountered path and depth bound 6 are preserved. The original program field names the entry; file and unit fields supply actual owner attribution. No probe is added to production lower source.",'',"Scalar enum/boolean unions are classified as scalar fields using the adapted source declarations, rather than treated as objects merely because TypeFlagsUnion is set. The raw object_pair flag records broad probe eligibility; the final split applies scalar-types.json to that candidate set. Annotations add full nested paths, nil-source target types and reference fields, and deriveRelated results for intersection records. Object classification accepts object/union shapes and intersections, with intersection targets counted as objects even when the original probe labels them target-not-object. Source intersections also qualify when the target is object/union shaped. The original skipped flags remain in raw rows. Nil-source records remain a separate bucket even when the target is an intersection. Intersections do not add a second recursive traversal, so their fields skipped by the original probe remain unwalked.",'','## Counts','',f"**Unrelated object pair observations: {len(unrelated)}; distinct textual (source, target, nested path) pairs: {len(pairgroups)}. Nil-source views with a known runtime reference field: {len(nil_refs)}.** The object-capable reference companion is {len(nil_objects)}. Nil-source views with no proven reference field but unknown fields or missing target: {len(nil_unknown)}.",'',f"Selected-unit pair observations: **{len(selected)}**. Raw lines across all passes, including superseded first-error and retry observations: **{len(all_pairs)}**. Selected units held {summary['view_sources_in_selected_units']} viewSources entries; pointer deduplication suppressed {summary['deduplicated_pair_visits']} repeat visits and the depth bound suppressed {summary['over_depth_pair_visits']} visits.",'','| Split | Selected unit observations | All raw lines |','| --- | ---: | ---: |']
for k,n in counts.items():lines.append(f'| {k} | {n} | {raw_counts[k]} |')
lines+=['',f"Intersection targets are a subset of the object counts, not an added bucket: {len(intersections)}, split as `{summary['intersection_split']}`. Scalar/non-object rows include source-proven primitive, enum and scalar-union fields that the probe's broad union flag test admits; scalar-types.json preserves that source-based classification. Intersection targets retain object status even when scalar-like; top-level non-object rows are shown separately from nested scalar fields.",'',"Reference field means a cycleFinder.fields field whose lowering representation is known and ir.Type.IsReference() is true. This runtime definition includes strings and Weak. The companion object-capable count excludes strings and Weak. Unknown representations remain explicitly unknown, not silently declared scalar. The nil-source count is a count of views; unlike non-nil pairs, the original probe does not deduplicate nil sources.",'','## Coverage','',f"Observed **{len(chosen)} unit results**: `{dict(status_counts)}`. Deferred standalone units in emitted inventories: **{skips['deferred_generic_instantiation']}**. Unobserved stock units by status: `{dict(skips)}`. Full-mode files attempted: **{len(full_finished)}/{len(resolved_files) or 81}**; emitted complete full-mode file inventories: {len(full_files)}. Stock inventory has {len(inventory)} units. Probe-incomplete selected units: {len(incomplete)}; pair lines without a completed probe unit row: {len(orphans)}.",'',"One result per (unit file, unit position) is selected: full mode supersedes first-error mode, and later full pieces supersede earlier ones. Repeated dependency pairs in different units remain repeated observations. TypeToString text can be truncated or reused for distinct checker identities; the distinct-pair list groups displayed source/target strings plus the full first-encountered nested path, not globally shared checker identities.",'',"First-error stopped units expose the attempted lowering state before the stop, even when their IR body would be discarded. Full mode rolls failed statements back and retains successful statements. The probe sees the retained state at the end of that unit. Neither mode counts rolled-back state, skipped/deferred bodies, units not reached before a time limit, paths hidden behind a statement's first failure, pair paths suppressed by the probe's per-unit seen set, or recursion past depth 6. Target fields absent from the source type are skipped by the original probe, and non-object records do not recurse. A completed file traversal still has those gaps. Missing buffered latent inventories are not evidence that a file had no units.",'',f"The latent hook does not compute GraphTypes or establish graph membership. GraphTypes was non-empty in {summary['graph_types_nonempty_units']} observed results. No acyclicity is inferred. Probe errors: {len(errors)}.",'','## Every distinct unrelated object pair','',"All counts and owner positions are also in [unrelated-object-pairs.json](unrelated-object-pairs.json). Paths are the nested field path at its first encounter in each unit. Files below are owner-unit files, not independent roots.",'','| Source type | Target type | Nested path | Count | Files and units |','| --- | --- | --- | ---: | --- |']
for g in pairgroups:lines.append('| '+esc(g['source'])+' | '+esc(g['target'])+' | '+esc('.'.join(g['path']) or '(top level)')+' | '+str(g['count'])+' | '+esc('; '.join(rel(u['unit'])+' x'+str(u['count']) for u in g['units']))+' |')
if not pairgroups:lines.append('| none observed | | | 0 | |')
lines+=['','## Every nil-source view with a runtime reference field','',"Each observation is retained, including repeats in one unit, in [nil-source-reference-views.json](nil-source-reference-views.json). Unknown-only cases are in [nil-source-reference-unknown.json](nil-source-reference-unknown.json).",'','| Unit | Target type | Runtime reference fields | Object-capable fields |','| --- | --- | --- | --- |']
for r in nil_refs:lines.append('| '+esc(rel(r['unit']))+' | '+esc(r['target'])+' | '+esc(', '.join(r['target_reference_fields']))+' | '+esc(', '.join(r.get('target_object_reference_fields',[])))+' |')
if not nil_refs:lines.append('| none observed | | | |')
lines+=['','## Per-file coverage','','| Owner file | Lowered | Stopped | Selected pairs | Full mode |','| --- | ---: | ---: | ---: | --- |']
for file in sorted(set(resolved_files)|{f for f,u in chosen}):
 bs=[r for (f,u),(_,r) in chosen.items() if f==file];ps=[r for r in selected if r['file']==file];full_job=next((j for j in full_finished if j['file']==file),None)
 state='not yet attempted'
 if full_job:
  st=dict(passes)[full_job['name']];state='complete traversal' if st['exit_code']==0 else 'bounded/nonzero exit'
 lines.append('| '+esc(rel(file))+' | '+str(sum(r['status']=='lowered' for r in bs))+' | '+str(sum(r['status']=='stopped' for r in bs))+' | '+str(len(ps))+' | '+state+' |')
lines+=['','## Pass exits','','Each child is capped at 1490 seconds with five seconds for forced cleanup, below 25 minutes. Priority and files larger than 300000 bytes run alone. At most two smaller files run concurrently. Completed pieces are committed and pushed as they finish. Each process group is reaped and cleaned by supervise.py.','', '| Pass | Child exit | Seconds | Reason |','| --- | ---: | ---: | --- |']
for name,st in passes:lines.append('| '+name+' | '+str(st['exit_code'])+' | '+str(round(st['seconds'],3))+' | '+esc(st['stopping_reason'] or ('completed' if st['exit_code']==0 else 'nonzero exit'))+' |')
lines+=['','## Evidence and replay','',"Raw streams, pass environments, timings, logs and actual child exits are in raw/. selected-pairs.jsonl.gz and selected-units.jsonl.gz preserve the selected population. overlay/ archives scratch Go inputs as .go.txt only. source-manifest.json hashes the adapted TypeScript bytes. provenance.json records pins, box and binary identity. SHA256SUMS covers evidence. report.py can recount finished archived/scratch streams without lowering. run-census.py resumes finished pieces without rerunning them.",'','```sh','source /workspace/adamic-tools/env.sh','export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source','bash stage3/apply.sh /workspace/view-pairs-adapted','python3 stage3/census/latent/make_overlay.py /workspace/view-pairs /workspace/scratch/vqzty8t/overlay','python3 stage3/census/view-pairs-734aef5f/instrument.py /workspace/view-pairs /workspace/scratch/vqzty8t/overlay','python3 stage3/census/view-pairs-734aef5f/prepare-probe.py','go build -overlay /workspace/scratch/vqzty8t/overlay/overlay.json -o /workspace/scratch/vqzty8t/census ./stage3/census/latent/tool','python3 stage3/census/view-pairs-734aef5f/run-census.py','```','']
(here/'REPORT.md').write_text('\n'.join(lines))
# Independent reconciliation against exact selected raw rows and emitted unit counts.
assert sum(g['count'] for g in pairgroups)==len(unrelated)
assert all(r['source_nil'] and r['target'] and r['target_reference_fields'] for r in nil_refs)
assert len(selected)==sum(u['pairs_logged'] for u in unitdetails)
assert not (root/'internal/lower/zz_view_pairs_probe.go').exists()
changed=git('diff','--name-only',base).splitlines();assert all(p.startswith('stage3/census/view-pairs-734aef5f/') for p in changed)
validation={'raw_recount_passed':True,'unit_pair_counts_passed':True,'unrelated_group_counts_passed':True,'nil_reference_rows_passed':True,'production_probe_absent':True,'scope_passed':True,'probe_errors':errors}
(here/'validation.json').write_text(json.dumps(validation,indent=2)+'\n')
audit=subprocess.run([sys.executable,str(here/'audit.py')],capture_output=True,text=True,cwd=root)
(scratch/'latest-independent-audit.log').write_text(audit.stdout+audit.stderr)
print('independent audit child_exit='+str(audit.returncode),flush=True)
if audit.returncode:raise RuntimeError(audit.stdout+audit.stderr)
files={str(p.relative_to(here)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(here.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'}
(here/'SHA256SUMS').write_text(''.join(f'{sha}  {name}\n' for name,sha in files.items()))
print(json.dumps(summary))
