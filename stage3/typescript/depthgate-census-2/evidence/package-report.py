#!/usr/bin/env python3
"""Package the completed immutable runs and render their report; never modify compiler sources."""
from pathlib import Path
import json,collections,gzip,shutil,hashlib,re,statistics
root=Path('/workspace/depthgate2-base/stage3/typescript/depthgate-census-2')
e=root/'evidence';analysis=Path('/workspace/depthgate2-analysis');runs=Path('/workspace/depthgate2-runs')
summary=json.loads((analysis/'run-summary.json').read_text());sites=json.loads((analysis/'site-outcomes.json').read_text());comparisons=json.loads((analysis/'finding-comparisons.json').read_text());stability=json.loads((analysis/'repeat-stability.json').read_text())
assert len(summary)==12 and all(v['complete'] for v in summary.values())
assert all(v['both_complete'] and v['normalized_events_equal'] and v['site_unit_events_equal'] for v in stability.values())
assert len(sites)==40 and all(s['classification']!='awaiting complete paired repeats' for s in sites)
for p in analysis.iterdir():shutil.copy2(p,e/p.name)
for src,name in [('depthgate2-analyze.py','analyze-census.py'),('depthgate2-runner.py','run-censuses.py'),('depthgate2-package.py','package-report.py'),('depthgate2-validate.py','validate-evidence.py')]:shutil.copy2(Path('/workspace')/src,e/name)
for r in sorted(runs.iterdir()):
 out=e/'runs'/r.name;out.mkdir(parents=True,exist_ok=True)
 for name in ['started.json','status.json']:shutil.copy2(r/name,out/name)
 for name in ['latent.jsonl','stderr.log','stdout.log']:(out/(name+'.gz')).write_bytes(gzip.compress((r/name).read_bytes(),mtime=0))
prep=e/'preparation';prep.mkdir(exist_ok=True)
for p in sorted(Path('/workspace').glob('depthgate2-*.log')):
 (prep/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
compiler='/workspace/depthgate2-adapted/src/compiler/'
def short(s):return s.replace(compiler,'').replace('/workspace/depthgate2-adapted/','')
def cell(s):return str(s).replace('|','\\|').replace('\n','<br>')
def code(s):return '`'+cell(short(s))+'`'
def mdtable(headers,rows):return '\n'.join(['| '+' | '.join(headers)+' |','| '+' | '.join('---' for _ in headers)+' |']+['| '+' | '.join(map(cell,row))+' |' for row in rows])
def unique(fs):return list({(f['kind'],f['phase'],f['where'],f['reason'],f['text']):f for f in fs}.values())
counts=collections.Counter(s['classification'] for s in sites)
oldI=[s for s in sites if s['probe_54210cd4_outcome']=='identity bound'];oldcounts=collections.Counter(s['classification'] for s in oldI)
p=[]
p.append('# Depth-gate submission census — #rt5evf5\n')
p.append('## Result\n')
p.append(f"All **12 runs completed normally** within 30 minutes. Both repeats give the same site and owning-unit findings. Of the **40 historical sites**, head **lowers {counts['lowers at mapped site']}**, **stops later at {counts['later stop in owning attempt']}**, and **retains the real Map check at {counts['real Map check']}**. There are **0 identity-bound sites** and **0 cap sites**, including outside the target population.\n")
p.append(f"Of the probe head `54210cd4`’s **34 identity-bound sites**, **{oldcounts['lowers at mapped site']} now lower** and **{oldcounts['later stop in owning attempt']} reach later refusals**. All 34 lose their old gate refusal, but the four later cases are not counted as successful lowering.\n")
p.append('**Growth outside the 40 exists:** `utilitiesPublic.ts:2559:13` newly reaches `a value of type JSDocArray | undefined`; the existing filter-overload refusal at `core.ts:261:1` and checked-field-write refusal at `binder.ts:1031:17` gain one owning context each. Two boundary messages change and two additional rollback boundaries are reached. These changes reproduce in both repeats. The other two measured files have only removals outside the 40.\n')
u={label:[summary[f'{label}-utilities-r{r}']['status'] for r in [1,2]] for label in ['base','head']}
bm=statistics.mean(x['wall_seconds'] for x in u['base']);hm=statistics.mean(x['wall_seconds'] for x in u['head'])
p.append(f"For **utilities.ts**, base wall times are **{u['base'][0]['wall_seconds']:.3f} / {u['base'][1]['wall_seconds']:.3f} seconds**; head is **{u['head'][0]['wall_seconds']:.3f} / {u['head'][1]['wall_seconds']:.3f} seconds**. The two-run mean changes **{(hm/bm-1)*100:+.1f}%**. The timing table retains every wall/RSS observation.\n")
p.append('## Inputs and reproducible method\n')
p.append('Base is `c56b022995e1e960bad71626fd2074bc46aee485`; head is `26d208af47f81143ff1db2280a5b43292deb96ba`. Each canonical entry tool was generated and built from its own SHA in an isolated worktree, using that SHA’s `make_overlay.py` and `entry_overlay.py`. Head uses main’s `typeNesting` (limit 5) for the growing-generic path bound; the beyond-depth-4 walk and 4096-pair cap are measured as supplied. No lowering source was patched. The report branch starts at base and does not merge the candidate.\n')
p.append('Both checkouts initialized their pinned submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, checker TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. One shared adapted tree, `/workspace/depthgate2-adapted`, was built successfully with **base’s `stage3/apply.sh`**, using TypeScript **v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`**, as the meter fetches it. Local verified clones supplied those commits.\n')
p.append('Every run uses `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_TAIL=1`, and an exact absolute `LATENT_ONLY_FILE`, through the entry tool rooted at shared `src/tsc/tsc.ts`. **Neither side sets `ADAMIC_PROBE_DEPTHGATE`**; inherited `ADAMIC_PROBE_*` flags are removed. The canonical `LATENT_TIMES=/dev/stderr` records per-unit timings; its pipe is captured whole. There are no DEPTHGATE or DEPTHWALK lines. Every stream has a full-mode/CommonJSBundle header, zero checker diagnostics, the same 81 source files, and its normal `relation_report` trailer. The no-output guard succeeds in all 12 runs.\n')
p.append('All runs are fresh child processes on the same machine, **sequential**, with no concurrent census or tool build. Per file the order is **base repeat 1, head repeat 1, head repeat 2, base repeat 2 (ABBA)**; file order is utilitiesPublic, transformer utilities, utilities. The runner measures end-to-end child wall time with a monotonic clock and child peak RSS with Linux `wait4.ru_maxrss` (KiB). Host, CPU/memory limits, timestamps, binary and selected-source hashes are in each status. The runner sends SIGTERM at 1,800 seconds and SIGKILL five seconds later if needed; **no timeout or kill was required**. Small report/progress reads and analysis occurred during the sequence; the machine was not benchmark-isolated.\n')
p.append('Evidence: [runner](evidence/run-censuses.py), [commands](evidence/commands.txt), [pins](evidence/build-pins.txt), [all run statuses and counts](evidence/run-summary.json), [source manifest](evidence/adapted-sources.sha256), [repeat stability](evidence/repeat-stability.json), and [raw run directories](evidence/runs/). Each run directory retains the entire stream, stderr and stdout, compressed losslessly.\n')
p.append('### Adapted tree compared with the earlier run\n')
p.append('**Not byte-identical.** Comparing all directory names, symlink targets and regular-file bytes (excluding `.git`, timestamps and permission bits) against `/workspace/depthgate-adapted` finds exactly **three changed paths among 81,506 regular files**: `patch-set.md`, `src/compiler/parser.ts`, and `src/compiler/types.ts`. Base’s additional adaptations annotate `lastQueriedPosition: number` and change Node/NodeArray heritage from ReadonlyTextRange to TextRange; both are type-only source edits. The three selected files are byte-identical to their earlier versions. These are pre-existing base adaptations, not changes made by this report. The prior-probe outcome comparison is therefore across these stated input differences as well as the candidate change. [Exact old/new hashes and tree manifests](evidence/tree-comparison.json).\n')
p.append('## The 40 at head\n')
p.append('The population is unchanged: **29 utilitiesPublic + 7 utilities + 4 transformer utilities = 40**. The broader scout’s 35 first-error sites plus eight extras made 43; `parser.ts:472:12`, `parser.ts:496:17` and `factory/utilities.ts:1614:9` are outside these three files. All 40 uniquely map by source anchors in their enclosing functions, with unchanged coordinates. Base has **38 fixed-depth-gate sites + 2 real Map-check sites**, confirmed in both repeats. [Exact targets](evidence/targets.json).\n')
p.append('Coordinates below are relative to `src/compiler/`. “Lowers” means the mapped comparison/statement no longer stops in this full census; it does not assert successful emission of the whole program. A new later refusal in the same owning attempt is conservatively classified **later**, even when the original comparison itself clears. “Real Map check” labels `container_fit`’s object-into-`ir.Map` guard path, not a new claim about runtime construction.\n')
p.append('Full reasons used below:\n\n* **G:** `a generic comparison deeper than 4 positions that can reach a Map` (fixed-depth at base; identity-bound at old probe head).\n* **M:** `a Map or Set interface implemented by a value that isn\'t a Map or Set`.\n* **L1:** `overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined` (Refused, lowering; `core.ts:261:1`).\n* **L2:** `writing a checked object field that may hold undefined or null, whose checked write can\'t change the representation its slot holds` (NotYet, lowering; `binder.ts:1031:17`).\n* **L3:** `a value of type JSDocArray | undefined` (NotYet, lowering; `utilitiesPublic.ts:2559:13`).\n')
rows=[]
for s in sites:
 r=s['repeats'][0];cl=s['classification'];fs=r['head_findings'] or r.get('new_owning_unit_stops',[])
 if cl=='lowers at mapped site':
  label='lowers';detail='No stop at mapped site.'
  rem=r.get('remaining_owning_unit_findings',[])
  if rem:detail+=' Existing owning-unit refusals remain; full records in site-outcomes.json.'
  else:detail+=' No remaining owning-unit findings.'
 elif cl=='real Map check':label='real Map check';detail='M (unchanged)'
 else:
  label='later'
  detail='; '.join(('L1' if f['where'].endswith('core.ts:261:1') else 'L2' if f['where'].endswith('binder.ts:1031:17') else 'L3')+' at '+code(f['where']) for f in unique(fs))
 owners=sorted({f['unit'] for f in r['base_findings']})
 old={'identity bound':'identity bound','real Map check':'real Map check','lowers at mapped site':'lowers'}[s['probe_54210cd4_outcome']]
 rows.append([s['id'],code(s['where']),'<br>'.join(code(x) for x in owners),s['function'],'M' if s['id'] in ['S34','S35'] else 'G',old,label,detail])
p.append(mdtable(['ID','Site','Measured owning unit(s)','Function','Base','54210cd4','26d208af (both repeats)','Reason / scope'],rows)+'\n')
p.append('**Which former 34 lift:** '+', '.join(s['id'] for s in oldI if s['classification']=='lowers at mapped site')+'. **S02, S25, S26 and S28 reach later stops** (L1, L2, L2 and L3 respectively). S25/S26 share one owning attempt and one later checked-write finding; they are two target sites, not two distinct later diagnoses. The four sites that already lowered at 54210cd4 (S01, S29, S30, S38) still lower. S34/S35 retain M under both `createNameResolver` and `resolveNameHelper`. S36 now lowers at its mapped site while those pre-existing Map refusals remain in its owning units.\n')
p.append('### Cap sites, listed separately\n\n**None.** No target or non-target finding has `a fit walk past 4096 distinct pairs` in any of the 12 runs. Head also has no finding using G; there is no observed remaining identity-bound site. This census does not measure how close a successful walk gets to the cap.\n')
p.append('## Findings outside the 40\n')
p.append('Pairing is the same as `typescript/depthgate-census` `7276d4ef`: key **(kind, phase, reason, message)** after stripping `:line:column` positions in reason/message. Equal keys pair first at the same site and owning unit, then within unit, then across remaining occurrences. Unmatched same-site/phase/unit records are **changed**; remaining base records are **gone**; remaining head records are **reached** when their point lies in a base rollback-boundary span in the same unit, otherwise **new**. Both raw event multiplicity and distinct `(kind, phase, site, reason, message)` findings are retained. No target-site events enter this outside-target comparison, but their rollback boundaries do.\n')
rows=[]
for name in ['utilitiesPublic','utilities','transformers-utilities']:
 for view in ['outside_target_events','outside_target_distinct_findings']:
  cc=[collections.Counter(x['grade'] for x in comparisons[f'{name}-r{r}'][view]['changes']) for r in [1,2]]
  assert cc[0]==cc[1]
  rows.append([name,'owning-context events' if view.endswith('events') else 'distinct source findings',*(cc[0][g] for g in ['new','gone','changed','reached']),'same in repeats 1 and 2'])
p.append(mdtable(['Measured file','Counting view','New','Gone','Changed','Reached','Repeat check'],rows)+'\n')
p.append('Counts are **per paired comparison**, not the sum of duplicate confirmations. All positive outside-target changes are in utilitiesPublic. The two **new raw events** are **additional owning contexts for existing distinct diagnoses**, so the distinct-finding table has zero new diagnoses. Growth is still present and is named below.\n')
rows=[]
for x in comparisons['utilitiesPublic-r1']['outside_target_events']['changes']:
 if x['grade']=='gone':continue
 f=x['head'];detail=f['reason']
 if f['kind']=='Boundary':
  detail+='; message '+('changes from G to L2' if f['where'].endswith(':1339:5') else 'changes from G to L3' if f['where'].endswith(':2559:5') else 'newly reached rollback boundary')
 elif f['where'].endswith('binder.ts:1031:17'):detail+=' (existing source diagnosis; additional owner)'
 elif f['where'].endswith('core.ts:261:1'):detail+=' (existing source diagnosis; additional owner)'
 rows.append([x['grade'],f['kind'],f['phase'],code(f['where']),code(f['unit']),detail])
p.append(mdtable(['Grade','Kind','Phase','Site','Owner','Full reason / change'],rows)+'\n')
p.append('Non-boundary removals outside the 40 are `reading last` at `utilitiesPublic.ts:2639:14` and `reading tag` at `utilitiesPublic.ts:1211:9`, `:1230:10` and `:1234:12`. The other gone records are rollback boundaries. Utilities has eight gone boundary events (five distinct sites); transformer utilities has four gone boundaries. The [complete outside-target appendix](evidence/outside-target-changes.md) lists **every grade with full before/after messages and owners**, including removals; the [machine-readable pairing](evidence/finding-comparisons.json) includes all events, matched counts, both repeats, and reached-span proofs.\n')
p.append('## Wall time and peak memory\n')
rows=[]
for name in ['utilitiesPublic','utilities','transformers-utilities']:
 for label in ['base','head']:
  for repeat in [1,2]:
   s=summary[f'{label}-{name}-r{repeat}']['status'];rows.append([name,label,repeat,f"{s['wall_seconds']:.3f}",f"{s['peak_rss_kib']:,}",f"{s['peak_rss_kib']/1024:.1f}",s['started_utc'],s['exit_code']])
p.append(mdtable(['File','Side','Repeat','Wall seconds','Peak RSS KiB','Peak RSS MiB','Started UTC','Exit'],rows)+'\n')
rows=[]
for name in ['utilitiesPublic','utilities','transformers-utilities']:
 bs=[summary[f'base-{name}-r{r}']['status'] for r in [1,2]];hs=[summary[f'head-{name}-r{r}']['status'] for r in [1,2]]
 b=statistics.mean(x['wall_seconds'] for x in bs);h=statistics.mean(x['wall_seconds'] for x in hs)
 rows.append([name,f'{b:.3f}',f'{h:.3f}',f'{(h/b-1)*100:+.1f}%',f"{min(x['peak_rss_kib'] for x in bs)/1024:.1f}–{max(x['peak_rss_kib'] for x in bs)/1024:.1f}",f"{min(x['peak_rss_kib'] for x in hs)/1024:.1f}–{max(x['peak_rss_kib'] for x in hs)/1024:.1f}"])
p.append(mdtable(['File','Base mean seconds','Head mean seconds','Mean wall change','Base RSS range MiB','Head RSS range MiB'],rows)+'\n')
separated=min(x['wall_seconds'] for x in u['head'])>max(x['wall_seconds'] for x in u['base'])
p.append('There are only two observations per side. Public and transformer timings overlap across sides. '+('Both utilities head observations exceed both base observations, so the utilities increase exceeds the observed within-side noise.' if separated else 'The utilities ranges overlap; its observed increase should be read with that noise.')+' The earlier **595.429 → 949.973 seconds** measured different base/head tools with probe logging enabled at head and the earlier adapted tree. It supplies context, not a controlled logging-only experiment. This rerun compares the submission head against its own base, using the same current tree and no probe logging.\n')
p.append('## Audit scope\n\nOnly this report and its evidence are committed. No adaptation, ledger, `internal/`, checker or probe implementation change is made or landed. [Validation](evidence/validation.txt) records completion, source/binary pins, repeat agreement, archive integrity and protected-path checks; [SHA256SUMS](evidence/SHA256SUMS) covers every other evidence file. The full messages and unit/pass records are retained in the lossless streams and [site outcomes](evidence/site-outcomes.json).\n')
(root/'REPORT.md').write_text('\n'.join(p).rstrip()+'\n')
a=['# All outside-target changes\n','Both repeats agree exactly by site/unit. Entries below show repeat 1 once; repeat 2 and both counting views remain in finding-comparisons.json. Positions in literal messages are retained here, although stripped for pairing.\n']
for name in ['utilitiesPublic','utilities','transformers-utilities']:
 a.append('## '+name+'\n')
 for i,x in enumerate(comparisons[name+'-r1']['outside_target_events']['changes'],1):
  f=x['head'] or x['base'];a.append(f"### {i}. {x['grade']} — {f['kind']} at {short(f['where'])}\n")
  for label in ['base','head']:
   q=x[label]
   if not q:a.append(label+': absent\n');continue
   a.append(f"{label}: {q['kind']}, {q['phase']}; owner `{short(q['unit'])}`\n\nFull reason:\n\n```text\n{q['reason']}\n```\n\nFull message:\n\n```text\n{q['text']}\n```\n")
  if x.get('reach_proof'):a.append('Reached proof: `'+json.dumps(x['reach_proof'])+'`\n')
(e/'outside-target-changes.md').write_text('\n'.join(a).rstrip()+'\n')
print('Packaged',len(summary),'runs;',dict(counts),'; former identities',dict(oldcounts))
