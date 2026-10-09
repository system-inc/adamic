import pathlib,json,re,statistics,subprocess
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-oracle-typeof_dispatch';groups=json.loads((p/'row-members.json').read_text());tests=json.loads((p/'requested-tests.json').read_text());witness=json.loads((p/'witness-tests.json').read_text());prod=json.loads((p/'production-tests.json').read_text());base=(p/'base.txt').read_text().strip();runs=json.loads((p/'audit-runs.json').read_text());clean=json.loads((p/'clean-runs.json').read_text());menu=json.loads((p/'menu.json').read_text())
def ev(log):
 es=[]
 for l in (p/log).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return es
def failed(log):return [e['Test'] for e in ev(log) if e.get('Action')=='fail' and e.get('Test') in tests]
def line(log,ts):
 lines=[x.get('Output','').strip() for x in ev(log) if x.get('Test','').split('/')[0] in ts and x.get('Action')=='output']; candidates=[s for s in lines if '_test.go:' in s and 'gate cache' not in s and not any(k in s for k in ['caught by','caught:','caught planted','removal caught']) and any(k in s for k in ['want','surviv','lost','not caught','missing','escaped','accepted','differ','checked'])];return next(iter(candidates),next((s for s in lines if 'FAIL:' in s),'no failure'))
def cmd(log):
 x=next(x for x in runs if x['log']==log);return ' '.join(k+'='+v for k,v in x['environment'].items())+' '+' '.join(repr(v) if '(' in v or '|' in v else v for v in x['command'])+' > '+log+' 2>&1'
raw={m['id']:failed(m['id']+'.log') for m in menu};matrix={id:[row for row,ts in groups.items() if any(t in fs and t in prod for t in ts)] for id,fs in raw.items()};(p/'matrix.json').write_text(json.dumps(dict(base=base,rows=list(groups),test_functions=tests,raw_failures=raw,production_row_failures=matrix,witness_production_failures_excluded=True,outside_rows='unknown'),indent=2))
seconds={row:statistics.median(float(re.search(r'([0-9.]+)s',(p/('timing-'+row.replace(' ','_')+'-'+str(i)+'.log')).read_text()).group(1)) for i in range(1,4)) for row in groups};record=[]
for row,ts in groups.items():
 files=[]
 for t in ts:
  matches=[]
  for f in ['typeof_dispatch_test.go','typeof_null_test.go','unknown_test.go','view_fields_test.go','wasi_shards_test.go']:
   text=subprocess.check_output(['git','show',base+':internal/oracle/'+f],cwd=r,text=True)
   if 'func '+t+'(' in text:files.append('internal/oracle/'+f+':'+str(text[:text.index('func '+t+'(')].count('\n')+1));break
 w=any(t in witness for t in ts);setup=ts==['TestWASIShardUnion'];kills=[id for id,fs in matrix.items() if row in fs];unique=[id for id in kills if len(matrix[id])==1];subs=[]
 if w:verdict='witness' if all(t in failed('W1.log') for t in ts) else 'untrue';log='W1.log';id='W1'
 elif setup:verdict='setup-check' if ts[0] in failed('S1.log') else 'untrue';log='S1.log';id='S1'
 else:
  subs=[other for other in groups if other!=row and kills and all(other in matrix[k] for k in kills)];subs.sort(key=lambda t:seconds[t]);verdict='sacred' if unique else 'subsumed' if subs else 'overlapping' if kills else 'untrue';subs=[] if unique else subs[:1];id=kills[-1] if kills else None;log=id+'.log' if id else 'M1.log'
 oracle='Node source execution; full stdout, stderr and exit-code disagreement, with clean sanitizer/leak preconditions';kind='external-run'
 if row=='TestRequiredViewFieldPrimitive':oracle='Node runs three positive source values; negative expected exits and exact panic text are handwritten. Built-in mutants must disagree with those self pins.';kind=['external-run','self']
 if row=='TestRequiredViewFieldOperandOnce':oracle='Node source pins full output 7/1 and backend agreement. M2 fails on wrong field lookup, not duplicate evaluation; failure label overstates its cause.'
 if row=='TestViewFieldReadiness family':oracle='Handwritten eager non-null panic prefix/suffix and empty stdout; inserted-check count >0; native/JavaScript agree. Source Node only must not exit 70, so other source failures could pass that health pin.';kind=['external-run','self']
 if row=='TestDefaultTaggedSourceViews':oracle='Node runs positive source cases. Negative field/literal panic texts and migrated eager non-null checks are handwritten self pins. Migrated source Node only must not exit 70.';kind=['external-run','self']
 if setup:oracle='Self: fixture registry union, identity uniqueness and shard construction; construction loop removed in S1.';kind='self'
 if row=='TestWASIShardPlantedDisagreement':oracle='Self: synthetic agreed/disagreed stdout, planted fixture identity and exactly one rejection.';kind='self'
 if row=='TestWASIShardPlantedFixture':oracle='Node source compared with compiled WASI fixture; parent expects nonzero child, owning shard and stdout-difference text. Full byte comparison is exercised.';kind=['external-run','self']
 probes=[]
 for x in json.loads((p/'probes.json').read_text()):
  if all(t in x['tests'] for t in ts):
   # A panic can omit top-level fail. The isolated command exit still demonstrates failure of its selected row.
   rr=[next(a for a in runs if a['log']==x['id']+'-'+t+'.log') for t in ts]
   if any(a['exit']!=0 for a in rr):probes.append(x['id'])
 if setup:
  failure=line(log,ts).replace('wasi_shards_test.go:73:', 'wasi_shards_test.go:76:')+' [origin line; raw S1.log line 73]'
 else:failure=line(log,ts)
 o=dict(test=row,package='internal/oracle',file=files[0],seconds=seconds[row],oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=unique,last_proven_fail=id+': '+failure if id else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=4,probe_kills=probes,subsumer_seconds=seconds[subs[0]] if verdict=='subsumed' else None,vacuous=None if w else False,bounded=True,matrix_rows=list(groups),evidence='Apply '+str(id)+'.diff at base, or use saved selector for production/probes; '+cmd(log)+'; '+failure,mutants_supporting_subsumption=len(kills) if verdict=='subsumed' else None,unique_scope='bounded matrix; package uniqueness unknown')
 if len(ts)>1:o['members']=[dict(test=t,file=f) for t,f in zip(ts,files)]
 record.append(o)
(p/'rows.json').write_text(json.dumps(record,indent=2))
probe_matrix={x['id']:{t:next(a['exit'] for a in runs if a['log']==x['id']+'-'+t+'.log') for t in x['tests']} for x in json.loads((p/'probes.json').read_text())};(p/'probe-matrix.json').write_text(json.dumps(probe_matrix,indent=2))
with (p/'diff-apply-check.log').open('w') as f:
 for d in sorted(p.glob('*.diff')):
  q=subprocess.run(['git','apply','--check',str(d.relative_to(r))],cwd=r,stdout=f,stderr=subprocess.STDOUT);assert q.returncode==0;f.write(d.name+' applies to '+base+'\n')
summary='''u071: all 13 requested tests exist at origin/main %s; two form one readiness family.
Clean whole-package baseline cooked at 90.036 seconds without observed test failures; bounded baseline passed.
Four production mutants were caught; M3 is unique only within the bounded matrix, with no survivors.
Twelve grouped rows: eight witnesses, one setup-check, one bounded sacred and two subsumed.
Tools warm; WASI SDK installed and enabled; evidence saved on the requested audit branch.
'''%base
mutant_table='| ID | Origin file:line | Change | Raw top-level failures |\n|---|---|---|---|\n'
for m in menu:mutant_table+='| '+m['id']+' | '+m['file']+':'+','.join(map(str,m['lines']))+' | '+m['change']+' | '+(', '.join(raw[m['id']]) or 'none')+' |\n'
notes='''
Survivors: none. All four fixed-menu production mutants changed exercised behavior and were caught. W1 disables disagreement at internal/oracle/oracle_test.go:716; all eight witnesses fail. S1 drops the entire distribution loop at internal/oracle/wasi_shards_test.go:30; TestWASIShardUnion fails with missing shard-000. Its raw log line 73 maps to origin line 76 after the three-line loop deletion. Probe diffs and logs are separate and never count toward production kills or uniqueness.

Brief ambiguities, corrections and costs:
- The brief cites files at 8de93800f4; current origin/main is the recorded base. All 13 names remained in their named files; none moved or vanished.
- The two readiness wrappers call the same assertMigratedNonNullCheck with different fixture/expression inputs and no additional assertions. They are one family, timed together three times. It has two members, not two independent uniqueness rows.
- A family verdict must follow grouped kills. M4 catches one readiness member and not the inherited-static member; that still counts as one family kill. No member subsumes its own family.
- RequiredViewFieldPrimitive combines normal controls with built-in check-removal mutants. The whole top-level row is a witness. Its M2 production failure is a broken precondition and is excluded from production kills.
- Constructor, string, null, slot-presence and unknown rows are also witnesses. Their existing mutants remain unchanged. Their verdict rests solely on the allowed weakened comparison, not compiler-mutant failures.
- The WASI union is suite construction. S1 removes the whole loop, keeping standalone Go vet valid without unused ordinal or row variables. PShards is a separate empty-construction probe.
- The at-most-four compiler rebuilding instruction conflicts with a target of three mutants per row. Four spread mutations were selected from reached functions before checking failures; no claim extends beyond this menu.
- The package run exceeded its 90-second test budget. All matrices ran the 13 requested test functions, grouped into 12 rows. Kills outside this slice and package/repository uniqueness are unknown. Sacred here is explicitly bounded.
- No clean top-level or subcase skip was observed in the requested slice. WASI was enabled with an installed SDK, including the real planted-fixture child. The interrupted whole-package baseline is not a complete skip inventory.
- The names NarrowedFieldUsesSharedReadiness and ViewFieldInheritedStaticReadiness no longer describe the executed boundary precisely: their current sources are .ts and stop at an eager non-null initializer. The oracle is partly handwritten check text and metadata, not a Node prediction of Adamic's extra checks.
- Migrated source health checks reject only Node exit 70. Other source failure codes would pass that check. Native expectations do pin stdout and diagnostic prefix/suffix; backend-to-backend agreement by itself is a self oracle.
- OperandOnce compares complete output with Node, including the counter, but M2 catches a wrong lookup name, not repeated operand evaluation. Its fixed failure label says more than the particular failing observation establishes.
- M1 is diagnostic text. It demonstrates exact-message enforcement, not changed runtime control flow. M4 does change control flow and observed stdout/exit. Subsumption rests on two mutants for the readiness family and one for OperandOnce; it is a hint, not a deletion verdict.
- No witness production entry was empty-probed because the brief directs witness judgment through the comparison it guards. Every production entry used by the three production rows was probed; individual family members were run separately to avoid one Lower nil panic hiding the other.
- Lower's empty answer leads to Go panics in callers. Those isolated failures are recorded as probe kills only; they do not prove useful production discrimination. PC/PJS also fail the selected rows. No positive subcase accepting an empty answer was observed.
- Go coverage lists every reached CUT function before mutation selection. Runtime C functions were not instrumented, so that inventory does not claim exact C reachability.
- An initial timing command used unavailable /usr/bin/time and did not run installation/baseline. It was corrected to Bash time before the actual baseline. Both actual installation durations are in wasi-install.log.
- Warm Go/Node/clang tools did not include the WASI SDK. Only the optional SDK was installed, not a fresh cloud setup. stage3/api npm ci ran before baseline; requested tests use repository Node runners rather than another node_modules directory.
- Native caches were distinct per mutant. Matrix wall durations combine product rebuilding and execution; exclusive native rebuild duration was not isolated. Oracle caches retain keyed source observations. Production source/IR and emitted products are regenerated and rechecked.
- No full repository gate or other package test suite was run. No oracle or fixture was mutated in the production menu. W1/S1/PShards are explicitly allowed witness/construction edits, kept separate. No pull request or main push is part of this audit.
'''
isolated=sum(x['wall'] for x in clean if x['log'].startswith('timing-'));audit_total=sum(x['wall'] for x in runs);switch=next(x['wall'] for x in runs if x['log']=='switch-build.log');npm=(p/'npm-ci.log').read_text().strip();timing='\nTiming: setup.sh skipped (warm tools), nproc 5. SDK download 1.824 s, extraction 2.509 s. npm: '+npm+'. Whole-package baseline binary 90.036 s, wall 92.838 s; bounded baseline wall '+str(round(clean[0]['wall'],3))+' s. Coverage wall '+str(round(clean[1]['wall'],3))+' s. The 36 isolated timing commands totaled '+str(round(isolated,3))+' s wall; medians of own binary lines are in the JSON. Selector binary build '+str(round(switch,3))+' s. Cold per-mutant matrix/rebuild command walls: '+', '.join(m['id']+' '+str(round(next(x['wall'] for x in runs if x['log']==m['id']+'.log'),3))+' s' for m in menu)+'. Vet, witnesses, construction check, selector build, matrices, probes and final verification together took '+str(round(audit_total,3))+' s command wall. Final bounded baseline passed in '+str(round(next(x['wall'] for x in runs if x['log']=='final-baseline.log'),3))+' s wall. Human/tool overhead not included. Full timing records: clean-runs.json and audit-runs.json.\n'
(p/'REPORT.md').write_text(summary+'\n```json\n'+json.dumps(record,indent=2)+'\n```\n\n'+mutant_table+notes+timing)
for f in ['/tmp/u071-clean.py','/tmp/u071-audit.py','/tmp/u071-report.py']:(p/pathlib.Path(f).name).write_text(pathlib.Path(f).read_text())
# Preserve the exact requested scope and actual observations. No failures from truncated binaries are inferred.
assert all(set(tests)=={x['Test'] for x in ev(m['id']+'.log') if x.get('Action') in ['pass','fail','skip'] and x.get('Test') in tests} for m in menu)
assert all(x.get('Action')!='skip' for x in ev('bounded-baseline.log'));assert not failed('final-baseline.log')
text=(p/'CODE-AND-ORACLE.md').read_text().replace('The three-member? No: this family contains exactly those two top-level tests. ','');(p/'CODE-AND-ORACLE.md').write_text(text)
print(json.dumps(dict(seconds=seconds,raw=raw,grouped=matrix,isolated_wall=isolated,audit_wall=audit_total,report_rows=len(record))))
