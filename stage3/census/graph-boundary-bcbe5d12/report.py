"""Recount every raw observation; full-unit observations replace first-error ones."""
import collections,json,gzip
from pathlib import Path
here=Path(__file__).parent;raw=here/'raw';kinds=['views','record_literals','dynamic_reads','checked_object_calls']
def records(name):
 p=raw/(name+'.gz')
 return [json.loads(line) for line in gzip.decompress(p.read_bytes()).decode().splitlines()] if p.exists() else []
headers=records('first-latent.jsonl')
resolved_files=set(headers[0].get('files',[])) if headers else set()
first=records('first-boundary.jsonl'); full=[]
for p in sorted(raw.glob('full-*-boundary.jsonl.gz')):full+=records(p.name[:-3])
chosen={(r['file'],r['unit']):r for r in first}
assert len({(r['file'],r['unit']) for r in full})==len(full), 'overlapping full pieces'
for r in full:chosen[r['file'],r['unit']]=r
rows=list(chosen.values()); counts=collections.Counter();distinct={k:set() for k in kinds};byfile={};examples={}
# Coverage inventories come from the original guarded latent stream, not guesses
# based on missing observer events. Deferred and checker-skipped units stay gaps.
stock={};file_inventory=set()
for name in ['first-latent.jsonl',*[p.name[:-3] for p in sorted(raw.glob('full-*-latent.jsonl.gz'))]]:
 for record in records(name):
  if 'file' not in record:continue
  file_inventory.add(record['file'])
  for u in record['units']:stock[record['file'],u['where']]=u
for file in file_inventory | resolved_files:byfile[file]=collections.Counter()
for r in rows:
 file=r['file'];byfile.setdefault(file,collections.Counter());byfile[file]['units']+=1
 byfile[file]['graph']|=int(r['graph_types_nonempty'])
 for s in r['sites']:
  k=s['kind'];assert k in kinds
  counts[k]+=1;byfile[file][k]+=1;distinct[k].add(s['where'])
  examples.setdefault((file,k),{})[s['where']]=s
source=Path('/workspace/graph-boundary-adapted')
def rel(file):
 try:return str(Path(file).relative_to(source/'src/compiler'))
 except ValueError:
  try:return str(Path(file).relative_to(source/'src'))
  except ValueError:return file
coverage=collections.Counter(r['status'] for r in rows)
skips=collections.Counter(u['status'] for key,u in stock.items() if key not in chosen)
view_components=collections.Counter('ViewOrigins entries' if site.get('detail')=='ViewOrigins' else 'CheckedFields keys' for row in rows for site in row['sites'] if site['kind']=='views')
scope_counts={scope:{k:sum(sum(s['kind']==k for s in r['sites']) for r in rows if ('/src/compiler/' in r['file'])==compiler) for k in kinds} for scope,compiler in [('compiler_79',True),('tsc_entry_2',False)]}
summary=dict(view_components=dict(view_components),resolved_files=len(resolved_files),counts_by_scope=scope_counts,counts={k:counts[k] for k in kinds},distinct_fallback_locations={k:len(distinct[k]) for k in kinds},observed_unit_results=len(rows),observed_statuses=dict(coverage),unobserved_stock_statuses=dict(skips),first_error_observed=len(first),full_observed=len(full),files_in_stock_stream=len(file_inventory),files_in_table=len(byfile),files_observed=len({r['file'] for r in rows}),graph_types_nonempty_units=sum(r['graph_types_nonempty'] for r in rows))
(here/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
provenance=json.loads((here/'provenance.json').read_text())
lines=['# Graph adoption boundary census on bcbe5d12','',"Task jawqwv7. Report only. Counts cover lowered unit results only, including retained successful statements after full-mode rollback. They do not measure accepted whole programs.",'',f"Compiler/apply: `{provenance['base_sha']}`. Reference report tree: `30057eca`. TypeScript 6.0.3 source: `{provenance['typescript_actual']}`.",'',f"Box: `{provenance['box']}`; cgroup CPU quota `{provenance['cpu_quota']}`; memory limit `{provenance['memory_limit']}` bytes; {provenance['visible_cpus']} visible CPUs. Environment: `{provenance['environment']}`.",'','## Four-way split, lowered IR only','','| Kind | Observations | Distinct reported locations |','| --- | ---: | ---: |']
for k in kinds:lines.append(f'| {k} | {counts[k]} | {len(distinct[k])} |')
lines += ['', 'Views combine ' + str(view_components['ViewOrigins entries']) + ' ViewOrigins entries and ' + str(view_components['CheckedFields keys']) + ' CheckedFields keys, as requested; these are metadata observations, not that many distinct adapter expressions.', '', 'The 79 compiler-file subtotal is `' + str(scope_counts['compiler_79']) + '`. The two entry-file subtotal is `' + str(scope_counts['tsc_entry_2']) + '`. Both cover lowered unit results only.']
long=max(kinds,key=lambda k:counts[k])
lines += ['',f"Long pole by observed count: **{long}, {counts[long]}**. Coverage caveat: lowered IR only; failed statements, checker-skipped bodies, deferred generics and uncompleted units are absent. Inference: this ranking may differ after those gaps lower.",'','## Coverage and interpretation','',f"Observed {len(rows)} unit results: {dict(coverage)}. First-error pass observed {len(first)} results; full pieces observed {len(full)}. The summed table uses one result per unit, preferring full mode over first-error mode. Repeated sites in different units are repeated observations, not independent source sites.",'',f"Stock unit inventory has {len(stock)} units in {len(file_inventory)} emitted files out of {len(resolved_files)} resolved files. Units without an observer result, by stock status: `{dict(skips)}`. Missing stock file records after an interrupted pass are not evidence of zero units.",'',"In first-error mode a unit with a lowering finding contributes zero sites: its body is discarded. In full mode the stock hook rolls failed statements back and continues. A `stopped` full-mode observation can therefore contain retained successful statements; its failed statements contribute no sites. Panicked unit results contribute zero sites. `lowered` means no new lowering finding in that unit, not whole-program acceptance. File-level refusal-scan findings do not suppress independent-unit examination.",'',"Every listed expression and metadata entry lacks a source-position field at this base. All positions are explicitly labeled unit fallbacks in the raw stream. Distinct reported locations deduplicate `(kind, unit file:line:column)`, and are not a bound on distinct actual source sites. Several expressions can share a fallback, while dependency bodies repeated in different units receive different owner fallbacks. Exact source-site deduplication cannot be recovered from this IR. The observer does not use the boundary's last-seen graphViewSite or graphDynamicSite.",'',f"GraphTypes non-empty in {summary['graph_types_nonempty_units']} observed results. The latent hook never runs `findCycles`; GraphTypes is not computed per unit. Empty flags here do not establish acyclicity. No graph membership is inferred.",'','## Per file, lowered IR only','',"Attribution is to the unit's file, not an independent root. A zero with zero observed unit results is a coverage gap, not evidence of an empty file. Dependency function bodies or metadata can be included in that unit's result.",'','| Unit file | Views | Record literals | Dynamic reads | Checked object calls | Observed unit results | Any observed GraphTypes |','| --- | ---: | ---: | ---: | ---: | ---: | --- |']
for f in sorted(byfile,key=rel):
 c=byfile[f];lines.append('| '+rel(f)+' | '+' | '.join(str(c[k]) for k in kinds)+' | '+str(c['units'])+' | '+('yes' if c['graph'] else 'no, not computed')+' |')
for kind in ['record_literals','views']:
 lines += ['',f'## Top 10 files by {kind}, lowered IR only','','| Unit file | Count |','| --- | ---: |']
 ranked=sorted((f for f in byfile if byfile[f][kind]>0),key=lambda f:(-byfile[f][kind],rel(f)))[:10]
 if not ranked:lines.append('| no sites observed in covered unit results | 0 |')
 for f in ranked:lines.append(f'| {rel(f)} | {byfile[f][kind]} |')
 lines += ['', 'Examples below are unit fallback positions and adapted unit source text, not exact expression locations.']
 for f in ranked:
  samples=list(examples.get((f,kind),{}))[:3]
  for site in samples:
   path,line,column=site.rsplit(':',2);text=Path(path).read_text().splitlines();start=max(0,int(line)-1)
   lines += ['',f'`{rel(path)}:{line}:{column}` (unit fallback):','```ts','\n'.join(text[start:start+3]),'```']
lines += ['','## Pass exits and limits','','| Pass | Child exit | Seconds | Stopping reason |','| --- | ---: | ---: | --- |']
for p in sorted(raw.glob('*-status.json.gz')):
 s=json.loads(gzip.decompress(p.read_bytes()));lines.append(f"| {p.name[:-15]} | {s['exit_code']} | {str(round(s['seconds'],3)) if s['seconds'] is not None else 'not timed'} | {s['stopping_reason'] or ('completed' if s['exit_code']==0 else 'nonzero exit; cause not inferred')} |")
lines += ['','## Method and reproduction','',"Read `graph_types.go:460` and its callers at `graph_types.go:355` and `program_region.go:140`. Count ViewOrigins entries plus CheckedFields keys without deduplicating their overlap; walk Main and every function body with the boundary's own walk for record ObjectLiteral, both dynamic-read nodes, and checked ObjectCall. These are hypothetical boundary candidates before the rest of ordinary lowering, not demonstrated graph-program refusals.",'',"The entry resolves 81 implementation files: the 79 compiler files from the independent-root report plus `src/tsc/tsc.ts` and `src/tsc/_namespaces/ts.ts`. The sums cover this resolved entry program. `instrument.py` uses the meter entry overlay to keep the entry first and expand Files to its implementation dependencies. It removes only the meter driver's full-mode-only checker-rejection gate so the requested first-error latent pass can retain diagnostics and proceed. `instrument.py` extends the existing guarded latent overlay in scratch only. It retains successful top-level statement returns in the measurement result's Main, which the stock hook discards. Ordinary Load and Lower remain disabled and every real pass asserts those guards. `overlay/` contains all generated Go inputs as `.go.txt`; no production lower source is changed. The deterministic self-test exercises Main, a function body, both dynamic reads, both view sources, GraphTypes presence, and failed-unit exclusion.",'',"An initial entry build omitted the observer mapping and failed; the corrected build and entry fixture passed before the real census. Both attempts are retained. Raw boundary streams, original latent streams, timing streams, logs and each child's own exit status are compressed separately beside this report in `raw/`. `source-manifest.json`, `provenance.json` and `SHA256SUMS` retain source and binary identities. Piece selections and measurement environment variables are retained in raw JSON. Full pieces use Go's `GOMEMLIMIT=10GiB` soft memory limit and a 60-minute aggregate budget; each child has at most 1790 seconds. Each process group is killed and reaped by the supervisor on completion or stopping.",'','```sh','source /workspace/adamic-tools/env.sh','bash cloud/setup.sh > setup.log 2>&1','npm ci --prefix stage3/api > npm-ci.log 2>&1','export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source','bash stage3/apply.sh /workspace/graph-boundary-adapted > apply.log 2>&1','python3 stage3/census/latent/make_overlay.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay','python3 stage3/census/graph-boundary-bcbe5d12/instrument.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay','go build -overlay /workspace/scratch/jawqwv7/overlay/overlay.json -o /workspace/scratch/jawqwv7/census ./stage3/census/latent/tool','# run-pieces.py records the exact entry invocation and file selections','```','']
(here/'REPORT.md').write_text('\n'.join(lines))
print(json.dumps(summary))
