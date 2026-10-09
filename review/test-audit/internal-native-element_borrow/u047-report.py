exec(open('/tmp/u047-run.py').read().split("if __name__")[0])
import statistics, collections, shlex
menu=json.loads((out/'menu.json').read_text()); commands=[json.loads(l) for l in (out/'commands.jsonl').read_text().splitlines()]
def events(file):
 result=[]
 for l in file.read_text().splitlines():
  try:result.append(json.loads(l))
  except:pass
 return result
matrix={}; fails={}; errors={}; statuses={}
for m in menu:
 id=m['id'];f=out/(id+'.log')
 if not f.exists():continue
 ev=events(f)
 for af in out.glob(id+'-alone-*.log'):ev+=events(af)
 states={r:'unknown' for r in rows};err={}
 for e in ev:
  test=e.get('Test','').split('/')[0]
  if test not in rows:continue
  if e['Action'] in ['pass','fail','skip']:states[test]=e['Action'] if e.get('Test')==test or e['Action']=='fail' else states[test]
  if e.get('OutputType')=='error' and test not in err:err[test]=e['Output'].strip()
 matrix[id]=states;fails[id]=[r for r in rows if states[r]=='fail'];errors[id]=err
(out/'matrix.json').write_text(json.dumps(matrix,indent=2))
with (out/'matrix.tsv').open('w') as f:
 f.write('selector\t'+'\t'.join(rows)+'\n')
 for id,states in matrix.items():f.write(id+'\t'+'\t'.join(states[r] for r in rows)+'\n')
seconds={}
for r in rows:
 vals=[]
 for n in range(1,4):
  text=(out/f'timing-{r}-{n}.log').read_text(); match=re.search(r'^ok\s+\S+\s+([0-9.]+)s',text,re.M);assert match,(r,n);vals.append(float(match[1]))
 seconds[r]=statistics.median(vals)
(out/'timings.json').write_text(json.dumps({r:{'runs': [float(re.search(r'^ok\s+\S+\s+([0-9.]+)s',(out/f'timing-{r}-{n}.log').read_text(),re.M)[1]) for n in range(1,4)],'median':seconds[r]} for r in rows},indent=2))
filemap={r:'internal/native/element_borrow_test.go' for r in rows[:4]}
filemap.update({r:'internal/native/fields_test.go' for r in rows[4:8]});filemap.update({r:'internal/native/heap_test.go' for r in rows[8:11]});filemap.update({rows[11]:'internal/native/ieee754_test.go',rows[12]:'internal/native/library_array_test.go',rows[13]:'internal/native/library_map_set_iterator_test.go'})
oracles=[
 'Self: exactly five indexed declarations, owner=NULL and absence of releases in generated C. Count can admit a different set of five; no native behavior compared.',
 'Self: named functions must or must not borrow across throws, catches and message calls.',
 'Self: bounded harmless closure borrows; virtual, unknown closure and writing targets do not.',
 'Self: both named virtual and closure length-only targets must permit a borrow; docs are context, not an external authority.',
 'Node stdout executed and compared in release and sanitized binaries; self-written emitted-C lookup patterns also require optimization and conservative fallback.',
 'Self: runtime C shape declarations must be represented by the production field-offset proof. The two production representations agree; no external authority.',
 'Node stdout executed and compared in release and sanitized binaries; self-written generated-C regex must retain checked field lookup.',
 'Node executes the missing-property fixture and must print 2\\n2\\n; self-written NotYet type, location and diagnostic require refusal before C.',
 'AddressSanitizer executes deliberately freed runtime strings; requires heap-use-after-free or use-after-poison diagnostics plus failing exit, and a clean untouched control. Classified as a planted-failure witness; production M9 failures excluded from its kills.',
 'Self: release harness validates string bytes and compares churn/control RSS <=1.5. M10 inflates both, so the relative oracle passes despite lost spare reuse.',
 'Self: hand-written Linux and macOS getrusage-unit cases for a test-local normalization helper; no outside value checked.',
 'Node Math runs for all 21 entry functions; exact result bits compared, with all NaN payloads treated equal. Baseline 1,281,787 answers, zero mismatches.',
 'Self: generated C must end in newline for small and >256 KiB programs. Field rows also catch this through clang diagnostics.',
 'Self: RSS allowance 4096 KiB, active iterator counts through exit codes, successful numeric total and CPU bound 0.35 s. The count subcase checks only zero exit, so unrelated nonzero failures also fail it.'
]
kinds=['self']*14
for i in [4,6,7]:kinds[i]=['external-run','self']
kinds[8]='external-run';kinds[11]='external-run'
entry={rows[0]:['P1','P2'],rows[1]:['P1'],rows[2]:['P1'],rows[3]:['P1'],rows[4]:['P2'],rows[5]:['P3'],rows[6]:['P2'],rows[7]:['P4'],rows[8]:['P5'],rows[9]:['P6'],rows[10]:[],rows[11]:[m['id'] for m in menu if m['id'].startswith('P_adamic_math_')],rows[12]:['P2'],rows[13]:['P7']}
# Production failures of the planted-failure witness never establish worthiness.
prod=[m['id'] for m in menu if m['kind']=='mutant']; ordinary=[r for r in rows if r not in [rows[8],rows[10]]]
kills={r:[id for id in prod if r in fails.get(id,[])] if r in ordinary else [] for r in rows}
results=[]
for i,r in enumerate(rows):
 ks=kills[r]; unique=[id for id in ks if len([t for t in ordinary if id in kills[t]])==1];subs=[];median=None
 if i==8:verdict='witness'
 elif i==10:verdict='setup-check'
 elif unique:verdict='sacred'
 elif not ks:verdict='untrue'
 else:
  candidates=[t for t in ordinary if t!=r and set(ks)<=set(kills[t])]
  if candidates:
   sub=min(candidates,key=lambda t:seconds[t]);subs=[sub];median=seconds[sub];verdict='subsumed'
  else:verdict='overlapping';subs=[t for t in ordinary if t!=r and set(ks)&set(kills[t])]
 last=ks[-1] if ks else ('W1' if i==8 else 'S1' if i==10 else None)
 ev=errors.get(last,{}).get(r)
 if last=='W1':
  for e in events(out/'W1.log'):
   if e.get('OutputType')=='error':ev=e['Output'].strip();break
 if last=='S1':ev=errors['S1'].get(r)
 cmd=next((c['command'] for c in reversed(commands) if c['name']==last),None)
 tested=entry[r]; probes=[id for id in tested if r in fails.get(id,[])];vac=None if not tested else any(matrix.get(id,{}).get(r)=='pass' for id in tested)
 obj=dict(test=r,package='internal/native',file=filemap[r],seconds=seconds[r],oracle=oracles[i],oracle_kind=kinds[i],kills=ks,unique_kills=unique,last_proven_fail=(last+': '+str(ev)) if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=len(prod),probe_kills=probes,subsumer_seconds=median,vacuous=vac,bounded=True,matrix_rows=rows,evidence=(shlex.join(cmd)+'; '+str(ev)) if cmd else 'ADAMIC_MUTANT=M10 go test -json -count=1 -timeout 90s ./internal/native/ -run "$(cat scope.regex)"; TestSizeClassesShareTheirChunks passed despite RSS inflation. See survivor logs.')
 if i==8:obj['vacuous_subcases']=['malloc/none','slabs/none']
 if verdict=='subsumed':obj['subsumption_mutants']=len(ks)
 results.append(obj)
(out/'rows-report.json').write_text(json.dumps(results,indent=2));(out/'final-array.json').write_text(json.dumps(results,separators=(',',':')))
# Evidence attribution and scope limits are explicit.
summary=['u047 audited all 14 named rows at b902a0ccc09e97940571388a1634450da3383559; none moved or vanished.', 'Whole package cooked at 90.096 s; clean bounded slice passed in 11.280 s; nproc 5.', '13 production mutants and 28 entry probes; uniqueness and subsumption are bounded to these 14 rows.', 'Verdicts: '+', '.join(f'{n} {v}' for v,n in collections.Counter(o['verdict'] for o in results).items())+'.', 'Evidence: test-audit/internal-native-element_borrow, review/test-audit/internal-native-element_borrow/.']
report='\n'.join(summary)+'\n\n```json\n'+json.dumps(results,indent=2)+'\n```\n\n| ID | Origin file:line | Change | Failed rows |\n|---|---|---|---|\n'
for m in menu:
 if m['kind']!='mutant':continue
 report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['old'].replace('\n',' ').replace('|','\\|')+' -> '+m['new'].replace('\n',' ').replace('|','\\|')+' | '+', '.join(fails.get(m['id'],[]))+' |\n'
report+='\nSurvivor M10: spare reuse disabled. Isolated clean churn/control 6280/5936 KiB becomes 49948/75816 KiB. Both runs pass. The relative gate fails to detect growth shared by its control. This is demonstrated changed behavior, not an equivalent candidate.\n\n'
report+='Brief ambiguities and costs:\n\n- The supplied historical SHA differs from current origin/main; the start rule wins. All 14 names still exist in the same six supplied files.\n- Full native package exceeds the 90 s budget. Only these 14 rows have an observed complete baseline and matrix. No full-package uniqueness is asserted. The whole-package baseline skipped 41 opt-in or missing-SDK rows outside this unit; exact names are in baseline-skips.json. No scoped row skips or needs an opt-in.\n- The suggested three mutants per row conflicts with the 20-mutant ceiling for 14 rows. I fixed 13 spread across the reached code before outcomes. The allocator row rests on one honest production attempt, M10, which it missed.\n- The at-most-four rebuild fallback concerns native cold builds. Four C production mutants share one environment-selected source and content-keyed runtime archive across selectors. Go selectors compile once; compiler product caches differ per selector. The runtime cache uses source bytes, compiler and flags, not ADAMIC_BUILD_CACHE_DIR, as library.go shows. No stale products are reused across differing runtime source.\n- The allocation safety row deliberately plants use-after-free to prove sanitizer visibility. I classify it as witness under the brief, with W1 disabling its instrumentation. Its M9 production failure remains in the matrix but never its kills or unique_kills. Disabling sanitizer in this special witness run is a permitted weakened-check harness edit, restored afterward.\n- The normalization helper is defined in heap_test.go. It is suite construction, so S1 changes its conversion divisor under the explicit setup-check exception. Production code cannot reach it.\n- Early-entry Go probe diffs need whole-body replacement and removal of newly unused imports to pass go vet. Their scratch selectors return at entry; saved standalone probes remove the unreachable body.\n- The empty Lower probe panics in rows using it only as preparation. Individually rerun rows lost to the abort; only the refusal row is judged by P4. Runtime-entry probes are separate for concat, allocation, map construction and every one of the 21 Math functions. The setup helper was not empty-probed, so its vacuous value is null.\n- The newline row shares its kill with strict clang builds in field rows. It is subsumed on one observed production mutant, not proposed for deletion.\n- The resource oracle for shared chunks admits a wrong result because the same mutant worsens its control. Its survivor witness is the clearest finding. Iterator count checks observe exit codes rather than the exact reason for failure.\n- C function reachability is a conservative source inventory, not measured C coverage. Native Go reached functions are backed by scope.cover. Full transitive lower and runtime coverage was not obtained. See reached-functions.txt.\n- I corrected an initially prepared direct test-binary launch before retaining results: Go tests require their package working directory. The retained matrix uses the required go test -json command. Repeated validation and the corrected build costs remain in commands.jsonl.\n- No external-authority claims are made from comments alone. Node and ASan actually ran; other expectations are self.\n\n'
report+='Timing and limits:\n\nWarm toolchain setup skipped, setup 0 s, npm ci 0.969 s. Every row ran alone three times; binary medians are in the array and raw runs in timings.json. Exact command wall times and exit statuses are in commands.jsonl. Whole-package binary budget 90.096 s; bounded coverage baseline command 32.297 s, binary 11.280 s. Runtime build time is included in test command time; no separate reliable per-runtime-archive timing was captured. Compilation, vet and clang validation wall times are itemized in commands.jsonl. No other packages tested, no repo-wide replay, no external-authority value cross-check, no exhaustive mutation coverage, no full native baseline completion.\n'
(out/'REPORT.md').write_text(report)
base=events(out/'baseline.log');(out/'baseline-skips.json').write_text(json.dumps([e['Test'] for e in base if e['Action']=='skip'],indent=2))
(out/'timing-totals.json').write_text(json.dumps({key:sum(c['seconds'] for c in commands if pred(c['name'])) for key,pred in {'isolated_timing_wall':lambda s:s.startswith('timing-'),'production_matrix_wall':lambda s:s in prod,'probe_wall_including_abort_reruns':lambda s:s.startswith('P'),'build_validation_wall':lambda s:s.startswith(('vet-','clang-','switch-','W1-vet'))}.items()},indent=2))
print('\n'.join(summary));print('Survivors:',[id for id in prod if not fails.get(id)]);print('Unknown:',[(id,r) for id,ss in matrix.items() for r,state in ss.items() if state=='unknown']);print((out/'timing-totals.json').read_text())
