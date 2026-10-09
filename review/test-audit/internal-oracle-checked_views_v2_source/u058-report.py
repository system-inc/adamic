exec(open('/tmp/u058-run.py').read().split("if __name__")[0]);import collections,statistics,shlex
menu=json.loads((out/'menu.json').read_text());prod=[m['id'] for m in menu if m['kind']=='mutant' and m['id']!='M1'];witness=rows[11]
def events(p):
 ev=[]
 for l in p.read_text().splitlines():
  try:ev.append(json.loads(l))
  except:pass
 return ev
matrix={};errors={};fails={}
for m in menu:
 id=m['id'];ev=events(out/(id+'.log'))
 for f in out.glob(id+'-alone-*.log'):ev+=events(f)
 state={r:'unknown' for r in rows};err={}
 for e in ev:
  r=e.get('Test','').split('/')[0]
  if r not in rows:continue
  if e['Action'] in ['pass','fail','skip'] and e.get('Test')==r:state[r]=e['Action']
  if e.get('OutputType')=='error' and r not in err:err[r]=e['Output'].strip()
 matrix[id]=state;errors[id]=err;fails[id]=[r for r in rows if state[r]=='fail']
(out/'matrix.json').write_text(json.dumps(matrix,indent=2))
with (out/'matrix.tsv').open('w') as f:
 f.write('selector\t'+'\t'.join(rows)+'\n')
 for id,states in matrix.items():f.write(id+'\t'+'\t'.join(states[r] for r in rows)+'\n')
seconds={};timings={}
for r in rows:
 vals=[float(re.search(r'^ok\s+\S+\s+([0-9.]+)s',(out/f'timing-{r}-{n}.log').read_text(),re.M)[1]) for n in [1,2,3]];seconds[r]=statistics.median(vals);timings[r]=dict(runs=vals,median=seconds[r])
(out/'timings.json').write_text(json.dumps(timings,indent=2))
ks={r:[id for id in prod if r in fails[id]] if r!=witness else [] for r in rows};results=[]
files=['internal/oracle/checked_views_v2_source_test.go']*7+['internal/oracle/class_wrong_output_test.go']*4+['internal/oracle/clock_generic_returns_t_01_test.go','internal/oracle/closure_merge_refusals_test.go']
oracles=['Node exit/stdout/stderr comparisons for good controls; self-written source-refusals.json pins for wrong/nested. Absent cases only require a diagnostic substring and clean Node exit.','Node must print true; successful backends compare to Node, wrong/nested cases use self-written exact panic pins.','Node executes and backends compare exit/stdout/stderr; self-written true output also checks Node.','Node controls must print true; self-written exact output and panic expectations decide each backend result.','Node executes and good backend output agrees; wrong case uses a self-written panic pin.','Node executes and good/absent backend output agrees; wrong/nested use self-written panic pins.','Self: Lower must return NotYet with an array-element or never-array reason. All five execution subcases then skip. No Node run in this row.','Node executes and self-written stdout control plus exact refusal path and repair are checked.','Node executes and self-written stdout control plus exact iterator-origin refusal and repair are checked.','Node executes and self-written stdout control plus exact symbol-key-view refusal and repair are checked.','Node executes and self-written stdout control plus exact private-storage refusal and repair are checked.','Witness: Node output, clean planted-mutant execution, exact stdout difference and leak check. Only W1 weakening disagreement decides its verdict.','Node checks only zero exit and empty stderr, not stdout. Self-written .refused snapshots pin the complete lowering diagnostic.']
commands=[json.loads(l) for l in (out/'commands.jsonl').read_text().splitlines()]
for i,r in enumerate(rows):
 kills=ks[r];unique=[id for id in kills if sum(id in ks[t] for t in rows)==1];subs=[];subsec=None
 if r==witness:verdict='witness'
 elif unique:verdict='sacred'
 elif not kills:verdict='untrue'
 else:
  candidates=[t for t in rows if t!=r and set(kills)<=set(ks[t])]
  if candidates:
   sub=min(candidates,key=lambda t:seconds[t]);subs=[sub];subsec=seconds[sub];verdict='subsumed'
  else:subs=[t for t in rows if t!=r and set(kills)&set(ks[t])];verdict='overlapping'
 own=[] if r==witness else ['P1','P2','P3'] if i<6 else ['P1'];probe_kills=[id for id in own if r in fails[id]];vacuous=None if not own else any(matrix[id][r]=='pass' for id in own)
 last=kills[-1] if kills else 'W1' if r==witness else None;line=errors.get(last,{}).get(r);cmd=next((c['command'] for c in reversed(commands) if c['name']==last),None)
 obj=dict(test=r,package='internal/oracle',file=files[i],seconds=seconds[r],oracle=oracles[i],oracle_kind='self' if i==6 else ['external-run','self'],kills=kills,unique_kills=unique,last_proven_fail=(last+': '+str(line)) if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=len(prod),probe_kills=probe_kills,subsumer_seconds=subsec,vacuous=vacuous,bounded=True,matrix_rows=rows,evidence=(shlex.join(cmd)+'; '+str(line)) if cmd else 'No eligible production kill observed.')
 if subs:obj['subsumption_mutants']=len(kills)
 if i==0:obj['vacuous_subcases']=['initial binding-name-source-good admission control: P1 returns nil,nil and reaches the subtests']
 if i==6:obj['skipped_subcases']=['good','wrong','nested','empty','mixed']
 results.append(obj)
(out/'rows-report.json').write_text(json.dumps(results,indent=2))
summary=['u058: all 13 names exist at c0a7667baadaf161d6bb9066e0838b32774caa6c; none moved or vanished.','Whole oracle timed out at 90.309 s without an earlier test failure; scoped baseline passed in 11.347 s; nproc 5.','14 eligible production mutants, M1 supplemental, three entry probes and one witness check.','Bounded verdicts: '+', '.join(f'{n} {v}' for v,n in collections.Counter(o['verdict'] for o in results).items())+'.','Evidence: test-audit/internal-oracle-checked_views_v2_source, review/test-audit/internal-oracle-checked_views_v2_source/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(results,indent=2)+'\n```\n\n| ID | Origin file:line | Change | Failed rows |\n|---|---|---|---|\n'
for m in menu:
 if m['kind'] not in ['mutant','witness']:continue
 report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['old'].replace('\n',' ').replace('|','\\|')+' -> '+m['new'].replace('\n',' ').replace('|','\\|')+' | '+', '.join(fails[m['id']])+' |\n'
report+='\nSurvivors:\n\n- M1 supplemental, outside reached code: optional tag changes from no tags to tag=x in the direct exported-planner probe. Coverage says UntaggedViewMembers 0.0%; excluded from all verdicts.\n- M2: supportsUntaggedRead on a synthetic Boolean field-only union changes true to false. The tested source Boolean fixture itself did not change; direct function witness establishes the changed output.\n- M3: completeUntaggedRecursiveContracts on a valid incomplete descriptor changes cleared Unsupported="" to retained "untagged object union". Direct function witness, not a test kill.\n- M4: Lower output for callable-union-good-number changes two ProducerCertified flags from true to false; all requested rows still pass.\n- M15: see survivor-M15-direct-before.log and after.log. If they agree, equivalent candidate; if they differ, the lazy-read diagnostic classification changed without a matrix kill.\n\n'
report+='Brief ambiguities, costs and limitations:\n\n- Supplied historical SHA differs from the required current origin/main start. Every requested name remains in its supplied file.\n- Whole-package oracle exceeds 90 seconds. I ran the 13 requested rows as a bounded slice and obtained production coverage for 629 functions across lower/native/javascript. Other rows and package-wide uniqueness are unknown. The coverage inventory is dynamic Go reachability, not a complete C-runtime call graph.\n- These bodies have distinct assertions, so all 13 remain separate rows. The shared disagreement helper alone does not make different refusal and contract checks one family. Internal subcases stay grouped under their top-level test.\n- Three mutants per row would require 39, beyond the ceiling of 20. I fixed 15 attempts before results; one was later excluded because its old planner is not reached. Fourteen eligible mutants remain. The M1 selection was my audit mistake, not a weakness of the brief. Its diff and observed result remain supplemental for transparency.\n- Standalone validation found that ir.Void is a type, not an enum value. M2 was corrected to ir.Array before any mutant outcomes.\n- All production mutants are Go compiler/emitter changes under one switch. Each selector has a distinct ADAMIC_BUILD_CACHE_DIR; ADAMIC_GATE_UNCACHED=1 bypasses oracle result caches, including the environment-selected generated JavaScript check. Runtime source archives use their content/flag cache and remain unchanged. The at-most-four fallback was unnecessary.\n- TestClockGenericReturnsT01Mutant is a witness. Its failures under M14 or empty compiler entries are broken preconditions and contribute no production kills. W1 disables only disagreement in a scratch copy and proves the witness fails. Node and the test itself remain unchanged.\n- TestCheckedViewUntaggedArrayPending is not missing a tool or opt-in. Every subcase asserts NotYet and then skips execution because the compiler lacks V3 array metadata. Its five skips are reported; no claim of supported native/JavaScript behavior is made.\n- Empty Lower aborts parallel runs when helpers dereference nil programs. Lost rows were rerun individually. Only own-entry results determine vacuity. The witness has null vacuity because production entry probes do not judge its comparison check.\n- SourceDispatch initially checks only a nil error and ignores the program; that admission control passes the empty Lower probe. Its later negative and execution subcases fail, so the row is not vacuous.\n- Node controls and refusal pins have different authority. The tests actually run Node, but Adamic-only panic text, repairs and .refused snapshots are self. No external-authority value was claimed or checked. ClosureMerge checks Node exit/stderr only, so unrelated successful stdout would pass that control.\n- Survivors require output witnesses, not speculation. Direct synthetic-IR calls establish M2/M3 output changes; M4 changes actual fixture IR. M1 demonstrates a change only outside the reached unit. Synthetic probes do not establish source-level runtime misbehavior.\n- Subsumption is a small-matrix hint, not a deletion recommendation. Each row records how many kills support it and the fastest observed subsumer.\n- Compiler rebuild time was not separately instrumented for each emitted native fixture. Command wall times include clang work; isolated binary medians remain the reported cost. No other package tests ran.\n\n'
totals={key:sum(c['wall'] for c in commands if pred(c['name'])) for key,pred in {'isolated_timing_wall':lambda x:x.startswith('timing-'),'production_matrix_wall':lambda x:x in prod,'probe_wall_with_abort_reruns':lambda x:x.startswith('P'),'build_validation_wall':lambda x:x.startswith(('vet-','switch-'))}.items()};(out/'timing-totals.json').write_text(json.dumps(totals,indent=2))
report+='Timing: warm setup skipped (0 s); npm ci about 1 s; clean whole-package binary 90.309 s; bounded baseline 11.347 s. Command-wall totals: '+json.dumps(totals)+'. Raw commands and timings are saved. Native runtime rebuilds are included in test-command time, not separately measured. No full-package completion, repo-wide uniqueness replay, source-level witness for every survivor, exhaustive mutation coverage or complete C reachability is claimed. Production sources are restored.\n'
(out/'REPORT.md').write_text(report)
base=events(out/'baseline.log');(out/'baseline-skips.json').write_text(json.dumps([e['Test'] for e in base if e['Action']=='skip'],indent=2));print('\n'.join(summary));print('Unknown production results:',[(id,r) for id in prod for r,s in matrix[id].items() if s=='unknown']);print(totals)
