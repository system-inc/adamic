import pathlib,json,re,statistics,csv,datetime,shutil
out=pathlib.Path('review/test-audit/stage1-cohere-css-print'); runs=json.loads((out/'matrix-runs.json').read_text()); scope=json.loads((out/'scope.json').read_text()); byname={x['test']:x for x in scope}; meta=json.loads((out/'mutants.json').read_text()); origins=json.loads((out/'scratch-to-origin-lines.json').read_text())
def events(path):
 result=[]
 for l in pathlib.Path(path).read_text().splitlines():
  try: result.append(json.loads(l))
  except: pass
 return result
def binary_seconds(path):
 for e in reversed(events(path)):
  m=re.search(r'ok\s+github.com/system-inc/adamic/stage1/cohere/css\s+([0-9.]+)s',e.get('Output',''))
  if m:return float(m[1])
 return None
def mapped(s):
 for p,lines in origins.items():
  name=pathlib.Path(p).name
  s=re.sub(re.escape(name)+r':(\d+):',lambda m:name+':'+str(lines.get(m[1],int(m[1])))+':',s)
 return s.strip()
def failure(r,n):
 candidates=[]
 for e in events(r['log']):
  s=e.get('Output','')
  if e.get('Test')==n and re.search(r'\b\w+_test\.go:\d+:',s):
   if any(x in s for x in ['=== RUN','shared exact-output','build ','Go and Node','original Prettier proofs','six boolean-flag','own work:','setup including','planted disagreement mode/case']):continue
   candidates.append(s)
 return mapped(candidates[-1]).splitlines()[0][:650] if candidates else 'No diagnostic selected; see '+r['log']
def evidence(r,n):
 return 'ADAMIC_MUTANT='+r['id']+'; selector /tmp/u079/selector='+r['id']+'; '+r['command']+'; '+failure(r,n)+'; log '+pathlib.Path(r['log']).name
semantic=['TestClosedPrinterRegexGap','TestCSSPrinterThroughput','TestCSSPrinterBoundaryProofs','TestOptionalBooleanPrinterMatchesGo','TestCSSProfileSnapshotsAgree','TestSharedSliceAppendAgreesWithNode']
state={m:{n:None for n in semantic} for m in ['M1','M2','M3','M4']}; proof={}; probes={}
for r in runs:
 if r['id'] in state:
  for n,a in r['outcomes'].items():
   if n in semantic and a in ['pass','fail']:
    state[r['id']][n]=a
    if a=='fail':proof[(n,r['id'])]=r
 if r['id'].startswith('P'):
  for n,a in r['outcomes'].items():probes[(n,r['id'])]=a
assert all(v is not None for row in state.values() for v in row.values()),state
assert all(x['code']==0 for x in json.loads((out/'workspace-port-validations.json').read_text()))
assert all(x['code']==0 for x in json.loads((out/'standalone-validations.json').read_text()))
assert all(x['code']==0 for x in json.loads((out/'complete-native-family-timings.json').read_text()))
timings=json.loads((out/'timings.json').read_text()); seconds={}
for n in set(x['test'] for x in timings):
 s=[binary_seconds(x['log']) for x in timings if x['test']==n]; seconds[n]=statistics.median(s)
seconds['TestProduct_CSSPrinterNative family']=statistics.median(binary_seconds(x['log']) for x in json.loads((out/'complete-native-family-timings.json').read_text()))
oracles={
semantic[0]:('Node and both backends must equal the self-written literal 2\\nOk\\n. parse(screen).kind is checked, but its tree contents are ignored; P2 passes.',['external-run','self']),
semantic[1]:('Go cohere and the pinned Prettier fork choose shared successful cases. Native, source Node, npm Prettier and fork summaries must match formatted counts and aggregate UTF-16 lengths. Equal-length wrong text could pass this checksum.','external-run'),
semantic[2]:('Go cohere directly formats four boundary cases; port source on Node must match complete answers. npm Prettier is run for separate logged boundary proofs.','external-run'),
semantic[3]:('Go cohere supplies complete six-case printer answers; Node source, native sanitizers and JavaScript backend must match exactly.','external-run'),
semantic[4]:('Go cohere supplies complete answers in default and narrow modes. Fresh source snapshot on Node and printer/profiled/counted native products must match stdout exactly.','external-run'),
semantic[5]:('Node, native and JavaScript backend each must print the self-written length literal 1152. It checks a length, not the appended bytes; M4 fails with ASan heap-buffer-overflow.',['external-run','self'])}
results=[]
for n in semantic:
 kills=[m for m in state if state[m][n]=='fail']; others=[x for x in semantic if x!=n and kills and all(state[m][x]=='fail' for m in kills)]; others.sort(key=lambda x:seconds[x]); subsumer=others[0] if others else None
 last=kills[-1]; r=proof[(n,last)]; own=['P2','P4'] if n==semantic[0] else ['P3'] if n==semantic[5] else ['P1']; seen={p:probes.get((n,p)) for p in own}
 row=dict(test=n,package='stage1/cohere/css',file=byname[n]['locations'][0],seconds=seconds[n],oracle=oracles[n][0],oracle_kind=oracles[n][1],kills=kills,unique_kills=[m for m in kills if sum(state[m][x]=='fail' for x in semantic)==1],last_proven_fail=last+' '+failure(r,n),verdict='subsumed' if subsumer else 'sacred',subsumed_by=[subsumer] if subsumer else [],mutants_in_matrix=4,probe_kills=[p for p,a in seen.items() if a=='fail'],subsumer_seconds=seconds[subsumer] if subsumer else None,vacuous=any(a=='pass' for a in seen.values()),bounded=True,matrix_rows=semantic,evidence=evidence(r,n),subsumption_mutants=len(kills))
 if n==semantic[0]: row.update(vacuous_entries={'mediaquery.parse':True,'adamic_regex_test':False},vacuous_subcases=['P2: media-query assertion accepts an empty Ok tree; all three backends still print Ok.'])
 results.append(row)
families={'TestProduct_CSSPrinterOracle family':['TestProduct_CSSPrinterParserOracle','TestProduct_CSSPrinterOracle'],'TestProduct_CSSPrinterNative family':['TestProduct_CSSPrinterSanitizedAndLowered','TestProduct_CSSPrinterSemicolonMutant','TestProduct_CSSPrinterIndentMutant','TestProduct_CSSPrinterWidthMutant','TestProduct_CSSPrinterDarwinLeaks']}
construction=[('TestProduct_CSSPrinterOracle family','S2','Successful Go oracle compilation/path preparation, with no executable-existence or semantic-answer assertion. S2 publishes paths without building any Go oracle binary.'),('TestProduct_CSSPrinterNative family','S1','Successful native product preparation, with no execution or artifact-existence assertion. S1 publishes paths while omitting native.Build; no native files are produced.'),('TestCSSProfileArtifacts','S3','Go cohere and the fork select shared sample inputs, but artifact assertions cover only calls/build success. S3 writes expected.txt under the wrong name and the test passes.')]
for n,ident,oracle in construction:
 members=families.get(n,[n]); r=next(x for x in runs if x['id']==ident); passline='; '.join(x+': PASS' for x,a in r['outcomes'].items() if a=='pass')
 row=dict(test=n,package='stage1/cohere/css',file=byname[members[0]]['locations'][0],seconds=seconds[n],oracle=oracle,oracle_kind=['external-run','self'] if n=='TestCSSProfileArtifacts' else 'self',kills=[],unique_kills=[],last_proven_fail=None,verdict='untrue',subsumed_by=[],mutants_in_matrix=0,probe_kills=[],subsumer_seconds=None,vacuous=None,bounded=True,matrix_rows=[n],evidence='ADAMIC_MUTANT='+ident+'; '+r['command']+'; '+passline+'; construction-artifact-proofs.json',construction_checks=[ident])
 if n in families:row['members']=members
 if n=='TestProduct_CSSPrinterNative family':row['skipped_members']=['TestProduct_CSSPrinterDarwinLeaks']
 results.append(row)
n='TestCSSPrinterShardingCatchesDisagreement'; r=next(x for x in runs if x['id']=='W1'); results.append(dict(test=n,package='stage1/cohere/css',file=byname[n]['locations'][0],seconds=seconds[n],oracle='Self-written disagreement/hash-owner expectation and duplicate/missing-case rejection. W1 disables stdout comparison; S4 disables union validation, and each independently makes the row fail.',oracle_kind='self',kills=[],unique_kills=[],last_proven_fail='W1 '+failure(r,n),verdict='witness',subsumed_by=[],mutants_in_matrix=0,probe_kills=[],subsumer_seconds=None,vacuous=None,bounded=True,matrix_rows=[n],evidence=evidence(r,n),witness_kills=['W1'],construction_kills=['S4']))
# Deliver in the brief's order, after grouping input-only product wrappers.
order=[semantic[0],semantic[1],semantic[2],semantic[3],*families,'TestCSSPrinterShardingCatchesDisagreement','TestCSSProfileArtifacts',semantic[4],semantic[5]]; results.sort(key=lambda x:order.index(x['test']))
(out/'results.json').write_text(json.dumps(results,indent=2))
with (out/'matrix.csv').open('w') as h:
 w=csv.writer(h); w.writerow(['mutant',*semantic]); w.writerows([m,*[state[m][n] for n in semantic]] for m in state)
(out/'families.json').write_text(json.dumps(families,indent=2))
(out/'survivors.json').write_text(json.dumps({'production_survivors':[],'equivalent_candidates':[],'outside_assigned_matrix':'unknown'},indent=2))
summary=['u079: 15 named functions exist at cf79ecec; none moved or vanished from the named files.','10 grouped rows: 6 subsumed, 3 untrue construction checks, 1 witness.','All four production mutants caught; no bounded unique kills or production survivors.','P2 passes the media-query entry; P4 fails the RegExp entry. Darwin product member skipped on Linux.','Evidence: test-audit/stage1-cohere-css-print, review/test-audit/stage1-cohere-css-print/.']
md='\n'.join(summary)+'\n\n```json\n'+json.dumps(results,indent=2)+'\n```\n\n'
md+='Production mutants, origin/main '+ 'cf79ecec3723604428ab91ebcb283400d05a1548'+'\n\n| ID | File:line | Change | Failed rows |\n|---|---|---|---|\n'
for m in meta:
 if m['id'] not in state:continue
 change={'M1':'default tabWidth: 2 → 3','M2':'fits loop: remaining >= 0 → remaining > 0','M3':'parse returns Error with audit early refusal before parsing','M4':'drop result->length = written;'}[m['id']]
 failed=[n for n in semantic if state[m['id']][n]=='fail']; md+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+change+' | '+', '.join(failed)+' |\n'
md+='\nProbes and construction checks are excluded from production kills and uniqueness.\n\n| ID | File:line | Operation | Observed result |\n|---|---|---|---|\n'
for m in meta:
 if m['id'] in state:continue
 rr=[r for r in runs if r['id']==m['id']]; observed={n:a for r in rr for n,a in r['outcomes'].items()}; md+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['category']+' | '+', '.join(n+' '+a for n,a in observed.items())+' |\n'
md+='\nSurvivors: none among M1 to M4 in the six-row bounded matrix. Kills outside that matrix are unknown. No equivalence claim is needed. S1, S2 and S3 are construction survivors, with witnesses in construction-artifact-proofs.json: zero native files, zero oracle binaries, and missing expected.txt with missing-expected.txt present.\n\n'
md+='''The brief, limits and time costs

The reference commit 8de93800f4 was stale. Fresh origin/main was cf79ecec3723604428ab91ebcb283400d05a1548. Every requested name still existed in its named file. The test list, scope locations and standalone apply checks are saved.

The full package reached its 90-second budget before completing the unit. A batch containing the opt-in throughput row also cooked, and a subsequent batch containing cold products cooked. All assigned runnable rows were then proved green alone. The sixteen original printer shards also cooked at 90 seconds. Their partial outcomes are preserved, but they do not establish family success or package uniqueness. The production matrix was narrowed to the six assigned semantic rows. Three construction rows and the witness have their own construction/check experiments. No other Adamic packages were run to seek uniqueness.

The throughput opt-in is unusually expensive for a measurement row: three internal rounds each run native, Adamic on Node, fork Prettier and npm Prettier, then Go throughput. ADAMIC_CSS_PRINTER_BENCH_ONCE=1 was enabled; median binary cost is 77.027 seconds, without the default ten repetitions. It is subsumed on this four-mutant set, rather than slow-worthy, because none of its kills is unique. The aggregate checksum catches M2's 605005 output units versus the expected 604909, but equal-length wrong text could pass.

Profile artifacts must be generated before snapshots are read. Enabling both in one concurrent batch risks reading an incomplete directory, so the clean artifact row ran first. Switched artifacts were freshly rebuilt once and reused only with the runtime selector, including native counter/profiler products. This avoids stale port products. Direct-build rows still call native.Build on each test invocation; their large main C translation unit is rebuilt despite the source selector. Those costs remain in the logs.

Input-only product declarations are families: the two Go oracle builders share cssOracleProduct, and five native builders share cssPrinterProduct. The Darwin wrapper's availability guard does not add a distinct assertion, so it belongs to the native family. The complete native family was retimed three times with the Darwin member included and skipped. Linux cannot install the macOS leaks tool. Its product recipe was not exercised. The three named mutant-product declarations merely build mutated products; they do not themselves witness a comparison failure.

The brief assumes one code entry per row. The closed-regex row reaches both mediaquery.parse and adamic_regex_test. P2's empty Ok tree is accepted by all three backends, while P4's false RegExp answer fails. vacuous is true for the accepted parse entry, with separate vacuous_entries preserving the nonvacuous RegExp entry. This aggregation is an explicit interpretation of an unspecified multi-entry case.

A timeout interrupted the first M1 fast batch. It was not counted as a kill. Every assigned fast row was rerun alone for M1, and those completed outcomes replace unknown entries. No other production or probe batch panicked. The broad caller baseline panic has no mutation verdict.

Cold detached validation worktrees unexpectedly rebuilt Go dependencies and hit the 120-second compilation backstop before a test binary produced output. Those failed validation attempts are retained. I stopped their children and validated the same standalone TypeScript patches sequentially in the warm starting workspace. All five native validations passed; every file was restored immediately. C patches pass the exact native runtime clang flags, and the five standalone witness/construction patches pass Go vet overlays. All thirteen standalone diffs apply to the starting origin/main. The switch and drivers are evidence only, and production sources are restored.

Setup used the warm Go 1.27.1 toolchain, so setup.sh was not run. nproc is 5. npm ci in stage3/api reported 471 ms; installing the required Prettier 3.9.6 oracle reported 822 ms. No other package-loaded node_modules directories were found. Assigned opt-ins were enabled: npm printer library, benchmark with once mode, profile artifact directory, and profile snapshots. The only assigned skipped function was TestProduct_CSSPrinterDarwinLeaks. Outside the assigned unit, the broad baseline also recorded TestCSSThroughput skipping its separate benchmark opt-in.

The 384-function port inventory comes from clean V8 traces over the entire printer corpus and both regression fixtures, before mutation. It includes module initializers and generated member initializers. Native runtime entry inventory is static, with untraced downstream C helpers marked conservative; native function coverage was not measured. This is a reachability limit, rather than a claim about native path coverage.

Subsumption rests on only one mutant for the closed-regex, boundary and shared-slice rows, two for optional booleans, and four for throughput/snapshots. Throughput and snapshots mutually subsume one another. These are bounded retention hints for the defender wave, not deletion recommendations. Three construction checks accept missing or misnamed artifacts; their construction survivors are separate from production survivors.

Timing and rebuild details are in timings.json, complete-native-family-timings.json, workspace-port-validations.json, builds.log and matrix-runs.json. Native product timing includes checking, lowering, C emission and clang, not clang alone. The switched artifact product built in 52.256 binary seconds; the original four-native-product cold family run took 32.312 seconds. Actual standalone native rebuild timings are reported below. Total elapsed exceeded the approximate 30-minute budget because of mandatory isolated timings, repeated cooked baselines, direct native rebuilds and the cold validation attempt. Package and repository uniqueness, macOS behavior, unsupported opt-in corpora outside this unit, and all omitted caller outcomes remain unmeasured.
'''
validations=json.loads((out/'workspace-port-validations.json').read_text()); md+='\n| Standalone | Binary seconds | Command wall seconds |\n|---|---:|---:|\n'
for r in validations:md+=f"| {r['id']} | {binary_seconds(r['log'])} | {r['wall']:.3f} |\n"
total_timing=sum(binary_seconds(x['log']) or 0 for x in timings); md+=f'\nOriginal isolated timing runs consumed {total_timing:.3f} binary seconds in total. Matrix commands consumed {sum(r["wall"] for r in runs):.3f} wall seconds. These totals are phase costs and must not be summed with overlapping validation attempts as elapsed time.\n'
md+='\nFull JSON, raw logs, standalone diffs, family membership, function inventory, selector source, compile validations and reproducible drivers are included in this directory. No PR was opened, and no production edits are committed.\n'
(out/'REPORT.md').write_text(md)
for p in ['/tmp/u079/validate_in_workspace.py','/tmp/u079/validate_ports.py','/tmp/u079/report.py']:
 shutil.copy2(p,out/pathlib.Path(p).name)
print('Generated',len(results),'grouped results; six-row matrix is complete.')
