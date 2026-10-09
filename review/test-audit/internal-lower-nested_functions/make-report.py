import pathlib,json,re,statistics,subprocess,difflib
p=pathlib.Path('review/test-audit/internal-lower-nested_functions');raw=json.loads((p/'rows.json').read_text());runs=json.loads((p/'runs.json').read_text());plan=json.loads((p/'plan.json').read_text())
def events(f):
 out=[]
 for line in (p/f).read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
matrix={}
for r in runs:
 es=events(r['log']);fails=sorted({e['Test'].split('/')[0] for e in es if e.get('Action')=='fail' and e.get('Test')});passed=sorted({e['Test'].split('/')[0] for e in es if e.get('Action')=='pass' and e.get('Test')})
 matrix.setdefault(r['id'],[]).append(dict(log=r['log'],command=r['command'],failed=fails,passed=passed,wall=r['wall'],status=r['status'],binary_seconds=next((e['Elapsed'] for e in reversed(es) if e.get('Action') in ['pass','fail'] and not e.get('Test') and 'Elapsed'in e),None)))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
timing=json.loads((p/'timings.json').read_text())+json.loads((p/'extra-timings.json').read_text());seconds={}
for name in {r['test'] for r in timing}:
 vals=[]
 for r in timing:
  if r['test']==name:vals.append(float(re.search(r'\t([\d.]+)s',pathlib.Path(r['log']).read_text())[1]))
 seconds[name]=statistics.median(vals)
(p/'medians.json').write_text(json.dumps(seconds,indent=2))
family=['TestNestedFunctionCycleIsRefused','TestNestedEnvironmentCycleIncludesDisjointSlots','TestNestedCallbackCycleIsRefused'];names=[raw[0],'TestNestedCycle family',raw[2],raw[4],raw[5],raw[6],raw[8],raw[9]]
line={x:int(n) for n,x in re.findall(r'(?m)^(\d+):func (Test\w+)',subprocess.check_output(['rg','-n','^func Test','internal/lower/nested_functions_test.go'],text=True))}
oracles={names[0]:'Self-written NotYet type and diagnostic substrings for block, generic-value and dynamic-this gaps.',names[1]:'Self-written Refused type and adamic/cycle-capable substring; all three cases use the same checker with different source inputs.',raw[2]:'Self-written IR invariant: exactly one allocation, two cells, complete layout and linked EnvironmentCell flags.',raw[4]:'Self-written IR invariant: a captured string parameter exists and is not Borrowed.',raw[5]:'Self-written proof predicate accepts a literal input and rejects the built-in SetProperty mutation; W01 alone decides witness verdict.',raw[6]:'Self-written lowering NotYet type and rebinding substring; TS2630 is bypassed, not used as the expected answer.',raw[8]:'Self-written NotYet type and exact bodyless diagnostic substrings on two direct entries.',raw[9]:'Checks only successful lowering, discards returned IR; PLower passes with nil output and nil error.'}
records=[]
for name in names:
 members=family if name==names[1] else [name]
 kills=[id for id in ['M01','M02','M03','M04'] if any(t in matrix[id][0]['failed'] for t in members)]
 unique=[id for id in kills if set(matrix[id][0]['failed'])<=set(members)]
 probes=['PLower'] if name in [names[0],names[1],raw[2],raw[4],raw[9]] else ['PClosedFrame'] if name==raw[5] else ['PDeclareModule','PStatements'] if name==raw[6] else ['PNestedDeclarations','PLowerBody']
 probe_kills=[];vacuous=False
 for id in probes:
  relevant=matrix[id];fail=any(any(t in m['failed'] for t in members) for m in relevant)
  if fail:probe_kills.append(id)
  else:vacuous=True
 verdict='sacred' if unique else 'subsumed' if kills else 'cannot-judge'
 by=[]
 if name==names[1]:by=['TestLiteralMethodCapturesCannotMakeCycles']
 elif name==raw[4]:by=[raw[2]]
 if name==raw[5]:verdict='witness';kills=[];unique=[]
 chosen='W01' if verdict=='witness' else kills[-1] if kills else None
 failure=None;evidence='No production mutant killed this row. See matrix.json; own-entry probes recorded separately.'
 if chosen:
  candidates=[e for e in events(chosen+'.log') if any(e.get('Test','')==t or e.get('Test','').startswith(t+'/') for t in members) and 'nested_functions_test.go:' in e.get('Output','')]
  failure=chosen+': '+candidates[-1]['Output'].strip();evidence='ADAMIC_MUTANT='+chosen+' ADAMIC_BUILD_CACHE_DIR=/tmp/u039/cache/'+chosen+' '+next(r['command'] for r in runs if r['id']==chosen)+' > '+chosen+'.log 2>&1; '+failure
 rec=dict(test=name,package='internal/lower',file='internal/lower/nested_functions_test.go:'+str(line[members[0]]),seconds=seconds[name],oracle=oracles[name],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=failure,verdict=verdict,subsumed_by=by,mutants_in_matrix=4,probe_kills=probe_kills,subsumer_seconds=seconds[by[0]] if by else None,vacuous=vacuous,bounded=False,matrix_rows=[],evidence=evidence)
 if len(members)>1:rec['members']=members
 if verdict=='cannot-judge':rec['reason']='Four compiler mutants exhausted the stated cap; no meaningful change to this row\'s specific guard or rest-parameter handling was exercised. No untrue verdict without that honest try.'
 if verdict=='subsumed':rec['subsumption_mutants']=len(kills)
 if name==raw[8]:rec['vacuous_subcases']={id:[] for id in probes}
 records.append(rec)
(p/'results.json').write_text(json.dumps(records,indent=2))
summary='\n'.join(['u039: all ten requested functions exist at f91994f019703ba25d2918cf529c0e0b0c05d93c; none moved or vanished.','Three cycle cases form one family, leaving eight audit rows.','Clean whole-package baseline passed in 21.589 s; nproc=5; warm toolchain setup skipped.','Four production mutants: two sacred rows, two subsumed rows; one proven witness; three cannot-judge rows.','Rest support is vacuous under PLower; every production mutant was killed; no production changes retained.'])
table='| ID | origin/main file:line | Change | Failed top-level tests |\n|---|---|---|---|\n'
for m in plan:table+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['menu']+': '+m['old'].replace('\n',' ').replace('|','\\|')+' -> '+(m['new'].replace('|','\\|') or '(removed)')+' | '+', '.join(matrix[m['id']][0]['failed'])+' |\n'
issues='''The supplied file reference is 8de93800f4, but fetching origin/main started this audit at f91994f019703ba25d2918cf529c0e0b0c05d93c. The test file is unchanged between those commits. All source lines and diffs refer to the actual starting commit.

The brief asks for about three mutants per row and also limits compiler mutants with separate rebuilds to four. I honored the four-mutant limit, which leaves the gap, rebinding and rest guards without an honest targeted attempt at their specific behavior. Those rows are cannot-judge, not untrue. The fixed plan was saved before test outcomes. The initial witness weakening `return called` failed vet because `safe` became unused. Before running tests it was replaced by `return safe || called`; both weakening and correction are recorded here.

The three nested cycle tests differ only in source input and use the same Refused/cycle-capable comparison. They are grouped as TestNestedCycle family, with members named in results.json. Their separate timings are retained, plus three runs of the grouped family. The family is subsumed by TestLiteralMethodCapturesCannotMakeCycles on one observed mutant; this is a small-matrix hint, not a deletion recommendation. CapturedParameters is subsumed by EnvironmentHasOneAllocationSite on one observed mutant.

The whole package fit the budget on all four production runs, so no narrowing was needed. The enabled package matrix is complete; repo-wide kills are unknown. The baseline skipped TestOriginalCycleLedger (external pristine corpus not configured), TestOptionalWideningCensus (external project inventory not configured), and one TestMixedUnionContractGraph subcase with an explicit compiler/views-v3 prerequisite. No requested test skipped. Those skipped checks remain unknown; no installable tool was missing for these ten rows.

PLower panicked in IR-inspection rows. All ten functions were rerun alone for that probe; only their observed results are recorded. Probe failures never count as production kills. ClosedFrameInput's PClosedFrame can fail during its production lowering precondition; only W01's mutable-graph assertion establishes the witness verdict. Two direct bodyless entries each reject their own empty probe; their unaffected sibling subcases pass, which is expected because they call different entries.

Warm setup took zero setup-script seconds. npm ci succeeded before baseline; its wall duration was not separately recorded. Clean test-binary compilation took {build:.3f} s. Per-mutant vet build validation is in validation.json. Production build-and-run wall times were {walls}; their binary run times were {bins}. Compilation/startup overhead is included in wall times and was not separately timed per mutant. All mutation, witness and probe runs total {run:.3f} wall seconds; 36 clean timing invocations total {timing:.3f} wall seconds. Baseline binary time was 21.589 s; clean coverage run was 0.215 s. Four mutants rather than twenty, the three unjudged guard paths, the skipped corpora, and repo-wide replay are not covered. No survivors exist in the four-mutant matrix.
'''.format(build=json.loads((p/'clean-build.json').read_text())['seconds'],walls=', '.join(f"{x}: {matrix[x][0]['wall']:.3f} s" for x in ['M01','M02','M03','M04']),bins=', '.join(f"{x}: {matrix[x][0]['binary_seconds']:.3f} s" for x in ['M01','M02','M03','M04']),run=sum(r['wall'] for r in runs),timing=sum(r['wall'] for r in timing))
(p/'report.md').write_text(summary+'\n\n```json\n'+json.dumps(records,indent=2)+'\n```\n\n'+table+'\nSurvivors: none. W01 is a witness weakening, not a production mutant.\n\n'+issues)
print(summary);print(json.dumps(seconds));print('run seconds',sum(r['wall'] for r in runs))
