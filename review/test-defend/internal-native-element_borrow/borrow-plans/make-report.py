import pathlib,json,re,subprocess,shlex
p=pathlib.Path('review/test-defend/internal-native-element_borrow/borrow-plans'); r=json.loads((p/'results.json').read_text()); plan=json.loads((p/'plan.json').read_text()); matrix=json.loads((p/'matrix-rows.json').read_text()); coverage=json.loads((p/'exclusive-coverage.json').read_text()); scope=json.loads((p/'scope.json').read_text()); base=(p/'base.txt').read_text().strip()
rows=[];passes={}
for aimed in plan:
 id=aimed['id']; test=aimed['test']; result=next(x for x in r if x['label']==id); failed=sorted({n for n in result['failed'] if '/' not in n}); passed=sorted({n for n in result['passed'] if '/' not in n});assert failed==[test] and len(passed)==29 and not result['skipped'] and not result['timeout'];passes[id]={'failed':failed,'passed':passed,'passed_subcases':[n for n in result['passed'] if '/' in n]}
 events=[json.loads(l) for l in (p/(id+'.log')).read_text().splitlines() if l.startswith('{')]; failure=next(x['Output'].strip() for x in events if x.get('Test')==test and 'element_borrow_test.go:' in x.get('Output',''))
 oracle={'TestNbodyIndexedElementsBorrow':'Self: exactly five planned indexed declarations, emitted NULL owner and absence of release of each borrowed local.','TestCallTargetsElementBorrowPlan':'Self: harmless bounded closure borrows; unknown callbacks, writing closures and virtual writer alternatives do not.','TestDevirtualizeBorrowDocClaim':'Self: both harmless virtual dispatch and bounded closure permit element borrowing. Internal documentation is context, not an outside authority.'}[test]
 subsumer='TestCallTargetsElementBorrowPlan' if test=='TestDevirtualizeBorrowDocClaim' else 'TestDevirtualizeBorrowDocClaim'
 command='ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+result['cache']+' '+shlex.join(result['command'])
 rows.append({'test':test,'package':'internal/native','prior_verdict':'subsumed','subsumed_by':[subsumer],'defense':'defended','unique_mutant':id+' '+aimed['file_line'],'attempts':[{'mutant':id,'file_line':aimed['file_line'],'change':aimed['change'],'rows_failed':failed}],'evidence':command+'; '+failure,'oracle':oracle,'oracle_kind':'self','bounded':True,'matrix_rows':matrix['rows'],'rows_passed':passed,'passed_rows_file':'passed-rows.json'})
 assert set(scope['added'])<=set(passed)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(p/'passed-rows.json').write_text(json.dumps(passes,indent=2)+'\n')
def witness(kind,id):
 a=(p/('clean-'+kind+'-execute.log')).read_text();b=(p/(id+'-'+kind+'-execute.log')).read_text();oldout=a.split('adamic: counts:')[0];newout=b.split('adamic: counts:')[0];assert oldout==newout
 def counts(s):return {k:int(v) for k,v in re.findall(r'(allocations|frees|retains|releases|peak|regions) (\d+)',s)}
 return {'stdout':oldout,'clean_counts':counts(a),'mutant_counts':counts(b),'delta':{k:counts(b)[k]-counts(a)[k] for k in counts(a)}}
w={'D1':witness('nbody','D1'),'D3':witness('doc','D3')};(p/'cost-witness.json').write_text(json.dumps(w,indent=2)+'\n');assert w['D1']['delta']['retains']==15000031 and w['D3']['delta']['retains']==2
whole=[json.loads(l) for l in (p/'baseline.log').read_text().splitlines() if l.startswith('{')]; assert not any(x.get('Test') and x['Action']=='fail' for x in whole);assert any('panic: test timed out' in x.get('Output','') for x in whole);elapsed=whole[-1]['Elapsed'];skipped=[x['Test'] for x in whole if x['Action']=='skip' and x.get('Test')];(p/'whole-baseline-skips.json').write_text(json.dumps(skipped,indent=2)+'\n')
clean=next(x for x in r if x['label']=='bounded-baseline'); assert clean['exit']==0 and len([n for n in clean['passed'] if '/' not in n])==30
lines=[]
for x in rows:lines.append('| '+x['unique_mutant'].split()[0]+' | '+x['attempts'][0]['file_line']+' | '+x['test']+' | '+x['evidence'].split('; ',1)[1]+' |')
text=f'''All three requested rows are defended in the complete 30-row bounded matrix.
Each has a distinct production mutant; all other 29 matrix rows pass, with no skips.
Production and tests restored; complete-package and repo-wide uniqueness remain unknown.

# Element-borrow defense

Starting origin/main: {base}. Audit base b902a0ccc09e97940571388a1634450da3383559. Read CLAUDE.md, audit REPORT.md and rows-report.json (all fourteen row oracle/verdict records), target test file whole, element_borrow.go whole, relevant borrow.go/IR call-target definitions and neighboring loop, inheritance and chain test bodies. README.md, docs/0.1.md and docs/memory.md are unchanged from the earlier warm session. Fetched audit with the requested full refspec and retained its report, row reports, row list and menu. Fetch encountered HTTP 503 once, then succeeded.

CODE UNDER TEST: native planElementBorrows, borrowable, assignedLocals, changingFunctions, changes, unchanging and borrowElement, plus native C assembly and ownership emission reached by nbody. ORACLE: self-written IR membership and generated-C assertions in these three rows. No Node execution or outside-authority cross-check in the target rows. Neighbor tests may compare to Node. No oracle, test, fixture or harness was changed. The fixed aimed plan is saved in plan.json, written before any mutant outcome.

## Disk and clean baseline

First df /tmp /workspace: /tmp 8.8G total, 258M used, 8.6G free; /workspace 32G total, 17G used, 14G free. Prior-unit /tmp/adamic-gate removal first encountered intentionally unwritable filesystem-test directories. Restored owner permissions only under that named scratch tree, removed it, and recreated TMPDIR mode1777. Second df: /tmp 238M used, 8.6G free; /workspace unchanged. The requested 15G cannot fit on the /tmp mount. No repository or tools removed, no disk-failure baseline.

Warm env.sh worked; setup skipped, nproc=5. npm ci --prefix stage3/api ran before baseline. Native split/job overrides stayed unset. The whole package exceeded its binary budget at {elapsed}s with no earlier observed test failure. It was stopped by Go's timeout and narrowed before any mutation. Exact skips and incomplete whole run are saved, not credited as passing replay evidence.

Current go test -list reports {len(scope['current'])} top-level functions versus audit {len(scope['audit'])}; none vanished. Two added rows, {', '.join(scope['added'])}, were included in every clean and mutant matrix. Three requested bodies are unchanged from the audit.

Source searches for C(), planElementBorrows() or unchanging() in test files selected 28 top-level rows, including every test in each matching file, not just the targets. Added both new rows: complete 30-row matrix in matrix-rows.json. This is a conservative source-call inventory, not complete transitive native reachability. {len(matrix['unknown'])} other top-level tests remain outside this replay and their mutant kills are unknown. The clean bounded baseline passed without skips: {clean['package_elapsed'][0]}s binary, {clean['wall_seconds']:.3f}s command wall.

## Coverage and semantic differences

Each target ran alone with -coverpkg=./internal/native, -coverprofile=<Name.cover> and anchored -run '^Name$'. All passed. Coverage profiles and exclusive block lists are retained. Counts relative to the named subsumers: nbody has {len(coverage[0]['exclusive'])} exclusive positive blocks; call-target has {len(coverage[1]['exclusive'])}; documentation claim has {len(coverage[2]['exclusive'])}.

Nbody reaches borrowElement's fallback owner declaration at starting-main element_borrow.go:221, unlike its planner-only subsumer. Its unique assertion demands owner = NULL and no release of the indexed borrowed local for all five planned declarations. D1 changes the existing emitted declaration's constant/ownership option to retain the existing element. Its format string gains the pointer operand needed to express that ownership change; it does not add another emitted statement. This leaves answers correct but adds ownership work.

Call-target coverage reaches the unknown-closure return at line177, absent from the documentation claim. Its mutable callback can switch to an element-writing implementation; the doc fixture's callbacks are bounded and read-only. D2 changes the unknown target return from false to true, treating an unbounded callback as harmless. It violates the closureRead negative assertion without breaking the doc positives.

Documentation claim has no exclusive blocks. Its shared-line semantic difference is benign multi-target virtual dispatch: Reader and Counter implementations both only read array length. The call-target row's virtual alternatives include a writer, so already must refuse borrowing. D3 returns early for virtual calls before checking their actual targets. It conservatively rejects even the benign alternatives, preserves answers, and loses the throughVirtual borrow. This is the permitted return-early mutation, not an arbitrary injected operation.

## Observed catches

| ID | Origin-main file:line | Only failing matrix row | Failing line |
|---|---|---|---|
'''+ '\n'.join(lines)+f'''

Every Dn.diff is standalone against the starting main, no selector or harness dependency. All three passed go vet ./internal/native/ and git apply --check after restoration. The matrix is timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run <matrix.regex>, with ADAMIC_GATE_UNCACHED=1 and each ID's own ADAMIC_BUILD_CACHE_DIR=/tmp/defend-elements/cache/Dn. Full actual argv, timings, passed subcases and failures are in results.json; all other 29 top-level rows are listed separately per mutant in passed-rows.json. No narrow run timed out or aborted, and none skipped.

Only one aimed mutant was needed for each row; the three-attempt requirement applies to a not-defended result. No such verdict remains. No survivor, empty-answer probe, test edit, oracle edit or weakening was used. These are not executor twins. D1 and D3 additionally exercise answer-preserving cost regressions, rather than relying solely on different answers.

## Compiled cost witnesses

D1's clean and mutant nbody programs were actually compiled and run with --count. Both print -0.169075164 and -0.169086185. Clean retains/releases are 10/14; D1 15000041/15000045, an increase of 15000031 each. Allocations and frees stay 8/8 and peak stays7. This proves changed work with unchanged output, beyond a formatting-only generated-C difference. The changed C was accepted by the real native build tool.

D3's clean and mutant documentation fixture were also compiled and run with --count. Both print virtual1, virtual1, closure2. Retains/releases rise from 6/13 to 8/15; allocation/free/peak/region counts remain unchanged. The conservative fallback changes ownership work, not these answers. This witness compares clean and mutant executions, not a newly claimed Node oracle. Raw build/run logs, counts and exact commands are retained. D2 changes planner membership, already directly observed by its row; its dangerous callback program was not additionally executed.

The clean three-row run passed after the matrix. Sources were restored in finally blocks after both additional cost witnesses; final git diff --exit-code -- internal succeeds. No production source or test change is committed.

## Brief ambiguities, costs and owner findings

* The 15 GB floor cannot be reached on an 8.8 GB /tmp mount; workspace has14G free. Only identified old-unit scratch was removed. Permission recovery was needed because old oracle fixtures deliberately lock directories. No full-disk failure occurred.
* One audit fetch returned HTTP503 and required retry. Its historical base differs from required current origin/main. Two new tests were included; requested bodies did not move or change.
* Whole-package timeout required a bounded replay. All 30 selected rows complete; outside-set uniqueness is unknown. The source inventory includes whole caller files and newly added rows but does not prove exhaustive indirect native reachability. No other packages were tested.
* Go coverage is measured for native Go planner/emitter code, not C runtime execution. Zero exclusive lines for the doc row did not rule out the semantic difference demonstrated by D3.
* D1 changes an existing emitted declaration and its format operand; D3 adds only the guarded early return. They are ownership-option and return-early changes from the permitted menu, with no fixture-specific filename/name/selector condition.
* All target oracles are self. Internal documentation and benchmark comments are not external authority. The counted executions establish relative changed work, not a general performance threshold or a separate Node agreement claim.
* Nbody's count-five assertion does not name-check the exact set of locals. It then verifies emitted ownership for whatever five the plan selected. The observed D1 defense protects ownership work, not the strength of that identity check.
* The documentation row checks borrowing through virtual/closure targets. It does not assert that native dispatch itself was devirtualized. The call-target row asserts membership for bounded readers, unknown/writing callbacks and virtual writers; it does not execute the miscompiled callback. All three remain defended for the assertions they actually make.
* Existing requested defense branch already contains other rows' evidence. Preserve it with a normal merge, storing this run in borrow-plans/. D1/D2/D3 identifiers are local to this subdirectory; root artifacts retain their own identifiers. No force push or history rewrite.
* Setup was skipped. npm install wall time was not separately instrumented. Go compilation, native artifact builds and test execution are combined in command walls where noted, not falsely reported as isolated compiler-only time.

Timing: whole binary {elapsed}s; clean bounded command {clean['wall_seconds']:.3f}s. Mutant command walls: '''+', '.join(f"{x['label']}={x['wall_seconds']:.3f}s (binary {x['package_elapsed'][0]}s)" for x in r if x['label'] in ('D1','D2','D3'))+f'''. Vet total {sum(x['wall_seconds'] for x in r if x['label'].endswith('-vet')):.3f}s. Restored clean command {next(x['wall_seconds'] for x in r if x['label']=='restored'):.3f}s. Additional counted build/run command timings are separately saved in nbody-witness-timings.json and doc-witness-timings.json. No full-package completion, repo-wide uniqueness or exhaustive C coverage is claimed.
'''
assert '\u2014' not in text
(p/'REPORT.md').write_text(text)
(p/'DISK.md').write_text('First df: /tmp 8.8G total, 258M used, 8.6G free; workspace 32G total, 17G used, 14G free. Removed prior-unit /tmp/adamic-gate after restoring owner permissions on deliberately locked fixture directories. Recreated mode1777. Second df: /tmp 238M used, 8.6G free; workspace unchanged. /tmp cannot satisfy15G floor.\n')
(p/'environment.json').write_text(json.dumps({'nproc':int(subprocess.check_output(['nproc'])),'setup':'warm env.sh, skipped','go':subprocess.check_output(['go','version']).decode().strip(),'node':subprocess.check_output(['node','--version']).decode().strip()},indent=2)+'\n')
print('three defended rows; cost deltas',w['D1']['delta'],w['D3']['delta'])
