import pathlib,json,csv,re,statistics
p=pathlib.Path('review/test-audit/internal-load-pin')
def events(m): return [json.loads(x) for x in (p/(m+'.log')).read_text().splitlines() if x.startswith('{')]
rows=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]
mat={}
for m in ['M1','M2','M3','M4','M5','M6','P1']:
 es=events(m); mat[m]={r:next((e['Action'] for e in reversed(es) if e.get('Test')==r and e['Action'] in ['pass','fail']), 'unknown') for r in rows}
for r in ['TestTypeScriptIsThePinnedCommit','TestRegExpCaptureTypes']:
 mat['P1'][r]=next(e['Action'] for e in reversed(events('P1-'+r)) if e.get('Test')==r and e['Action'] in ['pass','fail'])
with (p/'matrix.csv').open('w') as f:
 w=csv.writer(f); w.writerow(['test',*mat]); w.writerows([r,*[mat[m][r] for m in mat]] for r in rows)
(p/'matrix.json').write_text(json.dumps(mat,indent=2)+'\n')
def median(t): return statistics.median(float(re.search(r'\t([0-9.]+)s', (p/f'{t}-{i}.log').read_text())[1]) for i in range(1,4))
common=dict(package='internal/load',subsumed_by=[],mutants_in_matrix=6,subsumer_seconds=None,bounded=False,matrix_rows=rows)
results=[dict(common,test='TestTypeScriptIsThePinnedCommit',file='internal/load/pin_test.go:17',seconds=median('TestTypeScriptIsThePinnedCommit'),oracle='Git rev-parse HEAD versus self-recorded verified commit d92d9bfee114c80be2c375d72edae966176e3a4f. Git is run; the historical claim that the suite verified this pin was not independently checked.',oracle_kind='external-run',kills=[],unique_kills=[],last_proven_fail='S1 pin_test.go:24: TypeScript is at f901d0af4d5b5a9c445b2634aa4fe34ebff51ffd, and stage 0 was verified against d92d9bfee114c80be2c375d72edae966176e3a4f.',verdict='setup-check',probe_kills=[],vacuous=None,evidence='python3 review/test-audit/internal-load-pin/setup-probe.py; timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run ^TestTypeScriptIsThePinnedCommit$; S1.log pin_test.go:24: TypeScript is at f901d0af4d5b5a9c445b2634aa4fe34ebff51ffd'),dict(common,test='TestRegExpCaptureTypes',file='internal/load/regexp_test.go:5',seconds=median('TestRegExpCaptureTypes'),oracle='Self-written nonempty CheckError expectation. No outside authority is invoked or cited. M4 passes all four subcases despite a configuration error: observation of the split input changes TS18048 (value possibly undefined) into TS5052 (exactOptionalPropertyTypes requires strictNullChecks). Diagnostic identity is unchecked.',oracle_kind='self',kills=['M1','M2'],unique_kills=['M1','M2'],last_proven_fail='M2 regexp_test.go:11: Load: want a CheckError, got <nil>',verdict='sacred',probe_kills=['P1'],vacuous=False,vacuous_subcases=[],evidence='ADAMIC_MUTANT=M2 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M2.log 2>&1; regexp_test.go:11: Load: want a CheckError, got <nil>; isolated P1 also fails all four subcases.')]
(p/'results.json').write_text(json.dumps(results,indent=2)+'\n')
changes=[('M1','regexp_library.go:20','Drop exec-array rewrite'),('M2','regexp_library.go:21','Drop match-array rewrite'),('M3','regexp_library.go:22','Drop RegExp split rewrite'),('M4','load.go:59','Strict core.TSTrue -> core.TSFalse'),('M5','load.go:255','diagnostic.Code() -> 0'),('M6','load.go:278','column + 1 -> + 0')]
table='| ID | Origin file:line (internal/load/) | Change | Failed rows |\n|---|---|---|---|\n'
for m,f,c in changes: table+=f'| {m} | {f} | {c} | '+', '.join(r for r in rows if mat[m][r]=='fail')+' |\n'
report='''u025: both assigned rows exist, with no moves or vanished names.
Base: origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0.
Clean baseline: PASS, 1.090 test-binary seconds; no skips; nproc=5.
Verdicts: regexp sacred (M1 and M2 package-unique); pin setup-check (S1).
Evidence: test-audit/internal-load-pin, review/test-audit/internal-load-pin/.

'''+ '```json\n'+json.dumps(results,indent=2)+'\n```\n\n'+table+'''
P1 is a probe only: Load at load.go:80 returns nil,nil. The whole package aborts on a nil dereference in TestAdamicFilesImportEachOther. Rows without a completed result remain unknown in matrix.csv. Isolated assigned rows: pin passes, regexp fails all four subcases. S1 changes only the checkout construction to an identical-tree scratch commit and fails pin_test.go:24. It is setup evidence, not a production kill or standalone production diff.

Survivor M3: actual regexpLibraryFS.ReadFile output changes, witnessed by `go test -v -count=1 -timeout 90s ./internal/load/ -run ^TestAuditU025DeclarationWitness$`: explicit RegExp split undefined-union declaration is true before, false after. Saved survivor-original.log and survivor-M3.log. This is an unguarded declaration transformation; semantic equivalence remains unresolved because the separately rewritten Symbol.split overload remains. No additional survivor among M1 through M6.

Brief friction and execution limits:
- The supplied reference commit is older than fetched origin/main. Scope came from the actual list; neither assigned name moved or vanished.
- Warm env.sh did not imply initialized submodules. Initial listing failed with missing cohere/TypeScript/tsc/go.mod. npm ci succeeded before the baseline, then recursive submodule initialization took about 15.2 seconds.
- Initial cold test listing compilation ran roughly 95 seconds. I omitted its timeout and exceeded the 90-second step budget; no test binary approached 90 seconds. Subsequent full package runs fit without narrowing. The exact initial build wall time was not instrumented, so this is approximate.
- Pin reaches no production code. Treating repository submodule construction as setup is the applicable setup-check interpretation. Changing the recorded pin in the test would have violated the no-oracle-mutation rule, so it was left intact.
- Predeclared reached functions were conservative: coverage later showed Stat and Realpath uncalled; writeChain was entered but its body/recursion was unexecuted. All nonzero entries are recorded in coverage-functions.txt. The M5 line in the preplan was misnumbered; its correct origin/main line is 255.
- Empty Load caused a package panic. Isolated assigned-row reruns resolve only their results; other unfinished rows remain unknown.
- Supplemental temporary observation tests measured actual function output and diagnostics. They were not counted as mutants or kills and were removed. Scripts preserve the survivor witness; oracle logs preserve the diagnostic observation.

Costs: warm env sourcing/version check under one second; no cloud/setup.sh run. npm ci reported 782 ms. Submodules approximately 15.2 s. Initial checker compilation approximately 95 s. Isolated test binary samples: pin 0.005,0.006,0.005 s (median 0.005); regexp 0.063,0.073,0.098 s (median 0.073). Standalone vet checks total 3.739 s wall; switched binary build 4.659 s; six matrix commands 13.184 s wall, about 6.067 s in test binaries. Full per-command timings in timings.json. Compiler mutants change only Go loading, with no native product builds and no native build cache involved.

All six standalone production diffs and P1.diff apply to the starting origin/main and pass go vet ./internal/load/. The runner uses one mutation selector and excludes switch scaffolding from diffs. Production sources and submodule HEAD were restored. No other packages, repo-wide uniqueness, native products, or outside semantic authority were covered. No PR or main push. No subsumption verdict rests on this matrix.
'''
(p/'REPORT.md').write_text(report)
print(table)
