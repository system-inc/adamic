import pathlib,json,statistics,re,gzip,shutil,hashlib
p=pathlib.Path('review/test-audit/internal-unicodeproperties-canonicalize');matrix=json.loads((p/'matrix.json').read_text());probes=json.loads((p/'probe-results.json').read_text());timings=json.loads((p/'timings.json').read_text());setup=json.loads((p/'setup-probes.json').read_text());plan=json.loads((p/'mutant-plan.json').read_text());source=pathlib.Path('internal/unicodeproperties/canonicalize_test.go').read_text()
family='TestCanonicalizeUnicodeNodeRange family';names=['TestCanonicalizeExamples','TestEquivalentsAreClosed','TestCanonicalizeLegacyNode',family,'TestUnicodeNodeBatchOrderAndLimit'];scope=names[:];members=['TestCanonicalizeUnicodeNodeRange'+str(i) for i in range(34)]
def binary(q):
 for e in reversed(q['events']):
  m=re.search(r'\bok\s+\S+\s+([0-9.]+)s',e.get('Output',''))
  if m:return float(m[1])
 return None
def mapped(test):return family if test.startswith('TestCanonicalizeUnicodeNodeRange') else test
seconds={name:statistics.median([binary(q) for q in timings if q['id'].startswith('timing-'+name+'-')]) for name in names if name!=family};seconds[family]=None
kills={name:[q['id'] for q in matrix if name in [mapped(t) for t in q['kills']]] for name in names};unique={name:[q['id'] for q in matrix if [mapped(t) for t in q['kills']]==[name]] for name in names}
def failing(q,name):
 for e in q['events']:
  if mapped(e.get('Test','').split('/')[0])==name and re.search(r'canonicalize_test.go:\d+:',e.get('Output','')) and 'complete scans' not in e.get('Output',''):
   return e['Output'].strip()
 return 'No failing assertion observed'
oracles=[('Unicode 17 CaseFolding C/S mappings and ECMA-262 22.2.2.7.3 Canonicalize. Checked A -> a against downloaded CaseFolding.txt: 0041; C; 0061. Legacy a -> A also checked by Node.','external-authority'),('Self-written idempotence, containment, sortedness and class-consistency invariants; not an independent semantic oracle.','self'),('Node toUpperCase values and case-insensitive RegExp scans over every UTF-16 code unit; exact equivalence members compared, with scan counts.','external-run'),('Node RegExp iu/iv scans compared against generated class tables; table input preparation does not call CanonicalizeUnicode or UnicodeEquivalents. Only the first 16-line shard was in the bounded matrix.','external-run'),('Self-written expected peak=2 and ordered callback-result indices for suite-owned unicodeNodeBatches.','self')]
report=[]
for i,name in enumerate(names):
 is_setup=i==4;qs=setup if is_setup else matrix;last=next((q for q in reversed(qs) if (q['exit']!=0 if is_setup else q['id'] in kills[name])),None)
 # Setup S03 empty output has different fail from S02 ordering. Production evidence uses last killed mutant.
 fail=failing(last,name) if last else None
 eligible={'TestCanonicalizeExamples':['P01','P02','P03','P04'],'TestEquivalentsAreClosed':['P01','P02','P03','P04'],'TestCanonicalizeLegacyNode':['P04'],family:[],names[4]:['S03']}[name]
 vacuity={q['id']:name not in [mapped(t) for t in q['kills']] for q in probes if q['id'] in eligible}
 if is_setup:vacuity={'S03':False}
 if is_setup:verdict='setup-check';subs=[];pk=['S03']
 else:
  verdict='sacred' if unique[name] else 'subsumed';subs=[] if unique[name] else ['TestEquivalentsAreClosed'];pk=[id for id,v in vacuity.items() if not v]
 line=(source[:source.index('func '+(members[0] if name==family else name)+'(')].count('\n')+1)
 q=dict(test=name,package='internal/unicodeproperties',file='internal/unicodeproperties/canonicalize_test.go:'+str(line),seconds=seconds[name],oracle=oracles[i][0],oracle_kind=oracles[i][1],kills=([] if is_setup else kills[name]),unique_kills=([] if is_setup else unique[name]),last_proven_fail=(last['id']+': '+fail if last else None),verdict=verdict,subsumed_by=subs,mutants_in_matrix=12,probe_kills=pk,subsumer_seconds=(seconds[subs[0]] if subs else None),vacuous=(any(vacuity.values()) if vacuity else None),entry_vacuity=vacuity,bounded=True,matrix_rows=scope,evidence=(last['command']+'; '+('ADAMIC_MUTANT='+last['id']+'; ' if not is_setup else '')+fail if last else None))
 if name==family:q['members']=members;q['timing_lower_bound_seconds']=90;q['timing_status']='All three full-family attempts timed out at 90 seconds; no completed median.'
 if i==1:q['vacuous_subcases']=['Entire row passes P01: CanonicalizeUnicode returns zero.','Entire row passes P03: CanonicalizeLegacy returns zero.']
 if is_setup:q['construction_kills']=['S01','S02'];q['oracle']+=' S02 rotates slots and fails the order assertion; S03 returns empty results and fails peak=2.'
 report.append(q)
(p/'report.json').write_text(json.dumps(report,indent=2)+'\n')
short={'TestCanonicalizeExamples':'E','TestEquivalentsAreClosed':'C','TestCanonicalizeLegacyNode':'L',family:'U','TestUnicodeNodeBatchOrderAndLimit':'B'}
table='| ID | origin/main location | Change | Failed rows |\n|---|---|---|---|\n'
for m,q in zip(plan,matrix):
 changed=m['new'].strip().replace('\n',' ');table+=f"| {m['id']} | {m['file']}:{m['line']} | {m['kind']}: `{changed}` | "+', '.join(short[mapped(t)] for t in q['kills'])+' |\n'
summary='Unit u076: all 14 requested names exist; grouped into five rows.\nBase 60397548dd8a9494a7d2aaa874b7648b8e625607; nproc 5; warm tools.\nWhole package and three full-family timings cooked at 90 seconds; bounded baselines passed.\nTwelve production mutants killed; closure accepts zero-answer canonicalization probes.\nVerdicts: examples/closure sacred, legacy/Unicode family subsumed, batch construction setup-check.\n'
notes='''Survivors: none among the twelve production mutants in the bounded matrix.

Scope, oracles, findings and brief costs:
- E=examples, C=closure, L=legacy Node, U=Unicode Node range family, B=batch order/limit. Both requested and additional range wrappers have identical checker calls with only a range index changed. U has all 34 members; only Range0/0000-0016-* was replayed in the matrix. Requested Range1 through Range9 remain members with unobserved mutation outcomes. No requested name moved or vanished.
- The brief's reference commit 8de93800f4 differs from fetched main 60397548dd8a9494a7d2aaa874b7648b8e625607. Main now has 34 ranges, not just the ten listed. Locations and standalone diffs refer to the actual starting commit.
- Baseline whole package timed out at 90.014 seconds without an observed test assertion failure. Three full-family timings also timed out. No median for a completed full family exists; null is intentional. The 90-second cap prevented completing this census. Repeating family attempts was solely to meet the three timing-run request, not to reuse the full package per mutant.
- Coverage/dispatch and planted-failure tests in shards_test.go assert additional construction and witness properties, so are separate rows rather than interchangeable wrappers of the range checker. They are outside the requested unit and excluded from the bounded mutation matrix. Their production-mutant outcomes are unknown. No package- or repo-wide uniqueness is claimed.
- The bounded matrix includes every direct caller of the four production functions in these tests: examples, closure and the full legacy scan. Unicode table reads additionally run the first complete 16-line range shard. All other package rows, other Unicode shards and unrun family members remain unknown. The family and legacy subsumption are small-set hints, not deletion advice.
- The code-under-test inventory is code-under-test.txt. Four production functions and six generated fold/class arrays are reached. No outside package tests ran. The independent Node scripts and assertions were never mutated. Go vet validated every standalone production and empty-answer diff. The scratch switch was compiled once and its clean bounded baseline passed before matrix runs.
- The Unicode differential family checks generated equivalence tables, not the CanonicalizeUnicode function. M01/M02/M03 break that function and the bounded family still passes. Its table mutation M11 is caught by closure as well. The examples have unique M04; closure has unique M10 within this matrix.
- Closure passes when CanonicalizeUnicode returns zero (P01), and also when CanonicalizeLegacy returns zero (P03). Each causes equivalence lookup to fall back to singleton classes, preserving the self-consistency checks while erasing case folding. This row is vacuous for those entries despite catching meaningful non-probe mutations. It rejects empty equivalence slices P02/P04. The JSON records per-entry vacuity because a single Boolean cannot distinguish these results.
- Examples call all four entries and reject all four empty answers. Legacy Node calls LegacyEquivalents directly, so P04 determines its vacuity. Its transitive CanonicalizeLegacy probe P03 is also observed but is not listed as its own entry probe. The Unicode table family has no production function entry; its vacuity is null. Returning no harness-generated scan lines would mutate preparation, not production, so no such probe was used.
- The batch-order row is a setup-check, not a production test. S01 dropped the worker-cap assignment; S02 rotates result slots; S03 returns empty batch results. All are separate construction evidence, never production kills. S01 reported peak workers 1, not excessive concurrency, so it does not independently prove the cap bound. S02 preserves the comparisons and proves the order check can fail. S03 is the setup entry's empty-answer probe.
- Unicode 17 CaseFolding.txt was downloaded independently and A -> a checked against its 0041; C; 0061 mapping. Node independently reported legacy a -> A and Kelvin matching k under iu. Closure and worker expectations are self, not external authority. Node scans compare exact members as well as scan counts.
- Generated table edits change values only; no generator, Node oracle, assertion or comparison was altered. M05 drops the entire copy loop, keeping declarations used and the package compilable. Switch-only scaffolding is not present in standalone diffs and carries no verdict.
- No opt-in or missing-tool skip event was observed in baseline or bounded logs. A timed-out baseline cannot establish that unrun tests would not skip. The final restored bounded run passed. No broad repository gate was run because the brief forbids other packages.

Timing and coverage limits:
Toolchain setup skipped; npm ci installed 3 packages in 0.481 s; nproc=5. Package baseline 90.014 binary seconds, over budget. Full family three attempts were capped at 90 seconds, recorded individually. Individual three-run binary seconds follow; they are completed package ok-line seconds, not parent durations that exclude parallel children.
'''
for name in names:
 samples=[binary(q) for q in timings if q['id'].startswith('timing-'+name+'-')];notes+=f'{name}: samples {samples}; median {seconds[name]}.\n'
notes+=f"Scratch compile: {json.loads((p/'build.json').read_text())['seconds']:.3f} s; validation total {sum(q['validation_seconds'] for q in plan):.3f} s.\n"
for q in matrix:notes+=f"{q['id']}: command wall {q['wall_seconds']:.3f} s; binary {binary(q)} s (failed commands have no ok line; elapsed events are in matrix.json).\n"
notes+='No native products were built, so no native cache/rebuild timing applies. The mutation switch changes only this Go package and its static data. Full Unicode duration and kills outside the bounded set were not covered.\n'
(p/'REPORT.md').write_text(summary+'\n```json\n'+json.dumps(report,indent=2)+'\n```\n\n'+table+'\n'+notes)
for s in ['/tmp/u076.py','/tmp/u076-finish.py','/tmp/u076-report.py']:shutil.copy(s,p/pathlib.Path(s).name)
for log in (p/'logs').glob('*.log'):
 with log.open('rb') as inp,gzip.open(str(log)+'.gz','wb') as out:shutil.copyfileobj(inp,out)
 log.unlink()
(p/'SHA256SUMS').write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+str(f.relative_to(p))+'\n' for f in sorted(p.rglob('*')) if f.is_file() and f.name!='SHA256SUMS'))
