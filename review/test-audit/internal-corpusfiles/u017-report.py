import pathlib,json,re,statistics,shutil
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/internal-corpusfiles';rows=[x for x in (e/'list.log').read_text().splitlines() if x.startswith('Test')];meta=json.loads((e/'menu.json').read_text()); mids=[x['id'] for x in meta];matrix={};events={}
for mid in mids+['PRepository','PUpstream','PSelect']:
 events[mid]=[json.loads(l) for l in (e/(mid+'.log')).read_text().splitlines() if l.startswith('{')]
 matrix[mid]=[x['Test'] for x in events[mid] if x.get('Action')=='fail' and x.get('Test') in rows]
seconds={row:statistics.median(float(re.search(r'\s([\d.]+)s',(e/f'timing-{row}-{n}.log').read_text()).group(1)) for n in [1,2,3]) for row in rows}
proofs=['M08','M10','M09','M11','M15'];lines=[61,112,121,147,162];out=[]
for row,mid,line in zip(rows,proofs,lines):
 kills=[m for m in mids if row in matrix[m]];unique=[m for m in kills if len(matrix[m])==1]
 output=next(x['Output'].strip() for x in events[mid] if x.get('Test','').startswith(row) and 'files_test.go:' in x.get('Output','') and any(s in x['Output'] for s in ['planted corpus','selection:','no tracked']))
 oracle='Git runs establish actual tracked, dirty, ignored, pinned and sparse states; self-written counts, order, equality and diagnostic substring assertions decide correctness. Diagnostic substrings are weaker than full text.'
 if row==rows[-1]:oracle='Runs Git via Upstream and requires selection to succeed. Counts are only logged, never asserted; nil output passes PUpstream. This is a success-status oracle with no positive result assertion.'
 obj=dict(test=row,package='internal/corpusfiles',file=f'internal/corpusfiles/files_test.go:{line}',seconds=seconds[row],oracle=oracle,oracle_kind=['external-run','self'] if row!=rows[-1] else 'external-run',kills=kills,unique_kills=unique,last_proven_fail=mid+' '+output,verdict='sacred' if unique else 'subsumed',subsumed_by=[] if unique else [rows[0]],mutants_in_matrix=15,probe_kills=[p for p in ['PRepository','PUpstream','PSelect'] if row in matrix[p]],subsumer_seconds=None if unique else seconds[rows[0]],vacuous=row==rows[-1],bounded=False,matrix_rows=rows,evidence=f'GOWORK=off ADAMIC_CSS_FIXTURES=/tmp/u017-prettier ADAMIC_MUTANT={mid} timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > {mid}.log 2>&1; '+output)
 if row==rows[0]:obj['members']=['scanner','scanner/profile-runtime','typeaware','markdowninline','yaml','cssstrings','cssnumbers','graphql/printer','markdownblocks/default','markdownblocks/census','selector'];obj['vacuous_subcases']=[]
 out.append(obj)
(e/'rows.json').write_text(json.dumps(out,indent=2));(e/'matrix.json').write_text(json.dumps(matrix,indent=2))
for filename in ['u017-audit.py','u017-rerun.py','u017-witness.py','u017-report.py']:shutil.copy('/tmp/'+filename,e/filename)
report='''Unit u017, origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0.
Five top-level rows, all enabled; clean baseline passed.
Fifteen fixed-menu mutants: twelve killed, three survivors.
Four sacred rows; Prettier row subsumed and vacuous.
Evidence: test-audit/internal-corpusfiles, review/test-audit/internal-corpusfiles/.

Code under test and oracle, declared before mutations
Repository, Upstream, checked, selectFiles, git, names, matches. Their Go implementation selects corpus files from Git and validates pins, roots, cleanliness and physical membership. No Adamic lowering, native product or stage1 port is involved. Git is run to establish repository state; expected selection, ordering and error substrings are self-written. No external-authority classification is claimed.

Scope
All five names are in the starting list.log. No Your rows section was supplied; the entire package was used. No moved/vanished row can be identified without an earlier list. ConvertedPackageContracts has eleven input subcases sharing one checker, recorded as one top-level family row. No top-level helper or harness witness is present. Planted corpus faults are inputs to the collector under test, not mutants of an independent agreement oracle. The collector is treated as this unit's production code rather than as another suite's shard/coverage construction.

'''+json.dumps(out,indent=2)+'\n\nMutant table, every location against starting origin/main\n'
for m in meta:report+=f"{m['id']} internal/corpusfiles/files.go:{m['line']}: {m['change']}; {m['old']!r} -> {m['new']!r}; failed {matrix[m['id']]}\n"
report+='''
Survivors
M01 unguarded helper parsing behavior: witness-clean.log names=["tracked.ts"], witness-M01.log names=["tracked.ts" ""]. Existing selection ignores that empty entry on the tested patterns; no end-to-end corpus-selection difference was demonstrated.
M12 equivalent candidate: selected[absolute] true becomes false, but only map keys are ranged. No differing observable output demonstrated.
M13 unguarded counts: witness-clean.log files=1 counts=[1], witness-M13.log files=1 counts=[2]. Full-package logs double the logged counts without failing.
Witness command: GOWORK=off go test -v -count=1 ./internal/corpusfiles/ -run '^TestU017Observation$'. New temporary observation harness is preserved in survivor-witness.go.txt and was removed after running; existing tests were untouched.

Probes
PRepository returns nil at Repository entry, PUpstream returns nil at Upstream entry, PSelect returns nil files, empty HEAD, zero per-root counts and nil error at selectFiles entry. Returning zero counts sized to roots keeps wrapper logging valid. Own entries: Converted and Deterministic call Repository and selectFiles (Converted also Upstream); Missing and Sparse call selectFiles; Prettier calls Upstream. The Prettier probe passes. No other own-entry probe passes its row. No positive converted subcase completed successfully under its probe. Probes never contribute kills or uniqueness.

Ambiguities and costs
The brief refers to Your rows, but contains no list. Used whole package.
The brief says whole families count as one row; this family is already subtests inside a single top-level Test, so its original top-level name is retained.
The setup-check boundary is ambiguous for a package whose reusable product is itself a test-input collector. These rows directly test collector behavior; normal mutant verdicts are used, with the interpretation disclosed above.
Warm Go works, but go.work references absent cohere/TypeScript/tsc. Initial go test -list failed before any test. GOWORK=off resolves this standard-library-only unit without provisioning unrelated submodules.
/usr/bin/time is absent; used Bash time and Python monotonic timing.
The first scratch switch generator embedded a return as an expression and overlapped match replacements. Its compilation failed; no results from that attempt count. All standalone mutants had passed vet. The corrected switch and valid matrix logs are saved.
Subsystem docs are large; initial combined reads were truncated by output limits. Unit tests and collector implementation were subsequently read whole.
No oracle is weakened because no row is an independent agreement witness. No existing test, Git executable or upstream source was mutated.

Timing and validation
Warm toolchain check: 0.12 seconds, no setup.sh; Go 1.27.1, nproc=5.
Node npm ci: installer reports 709ms, three packages installed. No package test loads node_modules from another directory.
Prettier sparse provisioning completed between tool calls, approximately 19 seconds; no exact monotonic setup timer was captured.
Clean baseline Bash wall 0.448 seconds. Binary package time is recorded in baseline.log. No skips, panic, timeout or bounded narrowing.
Standalone mutants each passed GOWORK=off go vet ./internal/corpusfiles/, logs per id. Corrected switched binary compiled once; Go test runs reuse its build artifacts. No native rebuild needed. Individual timings and wall durations are in runs.json.
Subsumption rests on five caught mutants out of fifteen planted, not a deletion recommendation. Repo-wide uniqueness, other packages, malformed pins/root patterns and exhaustive glob semantics were not covered. Central replay can apply each M*.diff against the starting origin/main. Production source restored; final-baseline.log records the final green run.
'''
run=json.loads((e/'runs.json').read_text());report+=f"Measured commands total {sum(x['wall'] for x in run):.3f}s; timing commands {sum(x['wall'] for x in run if x['log'].startswith('timing')):.3f}s; standalone vet {sum(x['wall'] for x in run if '-vet' in x['log']):.3f}s; corrected switch build {sum(x['wall'] for x in run if x['log']=='switch-build.log'):.3f}s; matrix/probes {sum(x['wall'] for x in run if x['log'] in [m+'.log' for m in matrix]):.3f}s.\n"
(e/'REPORT.md').write_text(report)
print(json.dumps(out,indent=2));print(report[-900:])
