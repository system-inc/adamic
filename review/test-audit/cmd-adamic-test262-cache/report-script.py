import pathlib,json,re,statistics,shutil
p=pathlib.Path('/workspace/adamic/review/test-audit/cmd-adamic-test262-cache');r=p.parents[2];rows=json.loads((p/'requested-rows.json').read_text());menu=json.loads((p/'menu.json').read_text())
def events(log):
 a=[]
 for l in (p/log).read_text().splitlines():
  try:a.append(json.loads(l))
  except:pass
 return a
def fails(log):return [x['Test'] for x in events(log) if x.get('Action')=='fail' and 'Test' in x and '/' not in x['Test']]
def secs(log):return float(re.findall(r'\bok\s+\S+\s+(\d+(?:\.\d+)?)s', (p/log).read_text())[-1])
mat={m['id']:fails(m['id']+'.log') for m in menu};(p/'matrix.json').write_text(json.dumps(mat,indent=2))
times={t:statistics.median([secs('timing-'+t+'-'+str(i)+'.log') for i in range(1,4)]) for t in rows}
for t in ['TestEditCacheSeparation','TestLoweringSourceEdit']:times[t]=statistics.median([secs(t+'-timing-'+str(i)+'.log') for i in range(3)])
(p/'median-seconds.json').write_text(json.dumps(times,indent=2))
subs={rows[0]:rows[2],rows[1]:rows[2],rows[4]:min(['TestEditCacheSeparation','TestLoweringSourceEdit',rows[10]],key=times.get),rows[5]:rows[3],rows[8]:rows[9]}
oracles=[('Node executes the program; literal word and cache invocation counts are handwritten.',['external-run','self']),('clang runs generated C; literal word and cache invocation counts are handwritten.','self'),('Handwritten key equality, dimension separation, and length-boundary assertions.','self'),('Handwritten exact stdout/stderr bytes, status, reuse count, and corruption rejection.','self'),('Handwritten invocation counts, bypass environment and transient-result rules.','self'),('Node/native agreement plus handwritten serial/parallel report equality. M17 extra attempt and M19 doubled pass counts preserve parity.',['external-run','self']),('Node/native agreement plus handwritten progress lines and report counts.',['external-run','self']),('Node/native comparison inside attempt; test requires pass verdict despite an unavailable cache.',['external-run','self']),('Node/native runs plus handwritten freshness/reason inequality. Node import failure is expected; native outcome is not pinned exactly.',['external-run','self']),('Handwritten literal word, key dimensions, and dependency detection; own compiler output.','self'),('Node/native agreement plus handwritten cache-hit and scratch independence checks.',['external-run','self']),('Exact output/status/class/reason comparison with our own separate adamic c subprocess. Shared compiler errors remain invisible.','self'),('Subprocess entry only; parent TestCompilerWorkerTimeout supplies assertions. Standalone timing measures an immediate return.','self'),('Handwritten 20 ms deadline, TimedOut/status assertions, worker closure and process reaping for known hanging helper.','self')]
pr=json.loads((p/'probe-rows.json').read_text());pf={k:fails(k+'.log') for k in pr};(p/'probe-matrix.json').write_text(json.dumps(pf,indent=2));out=[]
for i,t in enumerate(rows):
 kills=[m for m,fs in mat.items() if t in fs];unique=[m for m in kills if mat[m]==[t]];mid=(unique or kills or [None])[-1];f=next(f for f in ['compiler_test.go','cache_test.go'] if 'func '+t+'(' in (r/'cmd/adamic-test262'/f).read_text());txt=(r/'cmd/adamic-test262'/f).read_text();line=txt[:txt.index('func '+t+'(')].count('\n')+1;detail=None
 if mid:
  detail=next((x['Output'].strip() for x in events(mid+'.log') if x.get('Test')==t and x.get('Output','').lstrip().startswith(('cache_test.go:','compiler_test.go:'))),None)
  if detail is None:detail=next(x['Output'].strip() for x in events(mid+'.log') if x.get('Test')==t and x.get('Output','').startswith('--- FAIL'))
 probe=[k for k,fs in pf.items() if t in fs];own=[k for k,rs in pr.items() if t in rs];helper=i==12
 obj=dict(test=t,package='cmd/adamic-test262',file='cmd/adamic-test262/'+f+':'+str(line),seconds=times[t],oracle=oracles[i][0],oracle_kind=oracles[i][1],kills=[] if helper else kills,unique_kills=unique,last_proven_fail=(mid+': '+detail) if mid else None,verdict='helper' if helper else 'sacred' if unique else 'subsumed' if t in subs else 'untrue',subsumed_by=[subs[t]] if t in subs else [],mutants_in_matrix=20,probe_kills=probe,subsumer_seconds=times[subs[t]] if t in subs else None,vacuous=None if not own else any(t not in pf[k] for k in own),bounded=False,evidence=('ADAMIC_TEST262_MEASURE=1 ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR=/tmp/u012/cache/'+mid+' timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > '+mid+'.log 2>&1; '+detail) if mid else 'timing-TestCompilerHangHelper-{1,2,3}.log; parent timeout evidence M14.log and M15.log')
 if helper:obj['parent']='TestCompilerWorkerTimeout'
 if t=='TestCompilerCacheProgram':obj['vacuous_subcases']=['PDependent: unnamed before/after native-output loop and four key-dimension loops passed; final reference assertion failed (PDependent.log cache_test.go:322). These earlier checks do not call dependentProgram.']
 if t in subs:obj['subsumption_kills']=len(kills)
 out.append(obj)
(p/'rows.json').write_text(json.dumps(out,indent=2))
summary='\n'.join(['u012: 14 requested rows present at origin/main 8171b3173bdbfce1f7982d3c4f731279307ece37; none moved or vanished.','Clean whole-package baseline passed in 42.720 binary seconds; nproc=5; warm toolchain, setup skipped.','20 vetted production mutants ran against all 37 top-level package tests: 18 killed, 2 survived.','Verdicts: 8 sacred, 5 subsumed, 1 subprocess helper; all 13 audited rows failed their direct-entry probes.','Evidence: test-audit/cmd-adamic-test262-cache, review/test-audit/cmd-adamic-test262-cache/.'])
report=summary+'\n\n```json\n'+json.dumps(out,indent=2)+'\n```\n\n| ID | origin/main file:line | Change | Failed rows |\n| --- | --- | --- | --- |\n'
for m in menu:report+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['change']+' | '+', '.join(mat[m['id']])+' |\n'
report+='\nSurvivors:\n\n'
for mid,clean in [('M17','clean-limit'),('M20','clean-context')]:
 def observation(id):return next(l.strip() for l in (p/('witness-'+id+'.log')).read_text().splitlines() if 'u012_observation_test.go:' in l)
 report+='- '+mid+': '+observation(clean)+'; mutant '+observation(mid)+'. See witness-'+clean+'.log and witness-'+mid+'.log, temporary observation harness saved in survivor-witness.go.txt.\n'
report+='''
The brief cites 8de93800f4, but required a fresh origin/main checkout. All file lines and standalone diffs use 8171b3173bdbfce1f7982d3c4f731279307ece37. All requested names remain in the two cited files. No requested wrappers share a family checker; their assertions differ. The helper has no independent verdict and its immediate-return timing does not price the parent workload.

The toolchain was warm, but the initial list build exceeded the 90-second cap and was stopped. Repository cohere submodules and stage3/api npm dependencies still required installation. The retry list and whole-package clean baseline succeeded. ADAMIC_TEST262_MEASURE=1 enabled the measurement row. No baseline rows remained skipped. Setup and dependency wall timings were not captured separately; installation output is retained. Every later whole-package mutant run completed below 90 seconds.

The slice did not require a bounded matrix: every production mutant ran all 37 top-level package tests. Empty probes ran their direct entry callers, with results in probe-rows.json and probe-matrix.json. Probe failures are excluded from production kills and uniqueness. No agreement witnesses or construction-only checks occur in these requested rows. Corrupt-cache inputs exercise production cache validation. A separate compiler subprocess is our own oracle, so its classification is self, despite crossing a process boundary. No outside-authority values were claimed or checked.

The 20-mutant cap gives fewer than three mutants per row. Subsumption is only a hint over the reported caught mutants, not grounds to delete tests. The fastest measured bypass subsumer was selected from both out-of-slice candidates and the requested scratch-directory row. Parallel report parity accepts the M17 extra attempt and M19 doubled count; imported-input checks assert freshness without pinning every expected native outcome. M20 changes cache identity without changing execution output, but its before/after key observation proves a real change.

The initial matrix driver was interrupted after M06 completed to narrow probes to direct callers. M06's full package event is intact; its wall timing is unknown. Continuation did not narrow any production matrix. All standalone production diffs passed go vet individually. Production source was restored before final clean package run, go vet, and git diff --check. Probe diffs are separate P files and are not production mutants. Temporary observation tests were removed; their source is retained as text.

Coverage and reached-functions.txt document the code read before the fixed menu. This audit covers cache identity/storage/fallback/dependencies, worker transport/timeout, limits, and progress/count accounting. It does not settle repo-wide uniqueness, shared compiler correctness, every error path, or mutations in lower/native code. No other package's suite was run. Per-mutant native runner products used separate ADAMIC_BUILD_CACHE_DIR directories; there was no compiler-lowering mutation requiring a cold compiler rebuild. Standalone diffs contain no runtime selector. The switch source and orchestration scripts are retained only as text.
'''
mr=json.loads((p/'mutant-runs.json').read_text());cr=json.loads((p/'clean-runs.json').read_text());fr=json.loads((p/'finish-runs.json').read_text());binary=sum(next(x['Elapsed'] for x in events(m+'.log') if x.get('Action') in ['pass','fail'] and 'Test' not in x) for m in mat)
report+='\nTiming: production matrix binary total '+str(round(binary,3))+' seconds; clean timing/coverage orchestration '+str(round(sum(x['wall'] for x in cr),3))+' seconds; individual mutant vet total '+str(round(sum(x['wall'] for x in mr if '-vet.log' in x['log']),3))+' seconds. Switch test-binary build was 6.165 seconds. finish-runs.json and mutant-runs.json retain measured command wall times, including probes and survivor observations. The audit exceeded the soft 20-minute budget because it retained whole-package matrices and three runs of every row.\n'
(p/'REPORT.md').write_text(report)
for source in ['clean','mutants','continue','finish','report']:
 path=pathlib.Path('/tmp/u012-'+source+'.py')
 if path.exists():shutil.copyfile(path,p/(source+'-script.py'))
print(summary);print('REPORT.md',len(report))
