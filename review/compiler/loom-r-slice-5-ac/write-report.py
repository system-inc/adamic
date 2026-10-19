from pathlib import Path
import json,hashlib,re,subprocess
out=Path(__file__).resolve().parent;root=out.parents[2];final=json.loads((out/'final.json').read_text());before=json.loads((out/'before.json').read_text());edits=json.loads((out/'headers.json').read_text());changed={r['path']:r for r in edits};mutants=json.loads((out/'header-mutants.json').read_text());tests=[];failures=[]
for name in ['header-readers.jsonl','header-tests-final.jsonl','fresh-path-readers.jsonl']:
 for line in (out/'logs'/name).read_text().splitlines():
  try:r=json.loads(line)
  except ValueError:continue
  if r.get('Test') and r.get('Action') in ['pass','fail','skip']:tests.append(dict(log=name,**r))
  if r.get('OutputType')=='error':failures.append(dict(log=name,test=r.get('Test'),error=r['Output'].strip()))
(out/'test-seconds.json').write_text(json.dumps(tests,indent=2)+'\n');(out/'reader-failures.json').write_text(json.dumps(failures,indent=2)+'\n')
shifts=[]
for path in changed:
 a=before['paths'][path]['stderr'];b=final['paths'][path]['stderr'];pattern=re.escape(path)+r':(\d+):(\d+):';old=re.search(pattern,a);new=re.search(pattern,b);assert old and new and int(new[1])==int(old[1])+1 and old[2]==new[2]
 shifts.append(dict(path=path,before_line=int(old[1]),after_line=int(new[1]),column=int(old[2]),reason=final['paths'][path]['reason']))
(out/'diagnostic-shifts.json').write_text(json.dumps(shifts,indent=2)+'\n')
setup=(out/'setup.log').read_text();timings=[s for s in setup.splitlines() if s.startswith('setup:')];(out/'setup-observations.json').write_text(json.dumps(dict(timing_lines=timings,nproc=5,cpu_max='400000 100000',toolchain={'go':'1.27.1','node':'24.19.0','clang':'20.1.8'},recovery='Initial source application stopped with No space left on device before setup could start. Removed only disposable /home/agent/.cache/adamic-build/local (about 11 GB), then restored the exact 505835b7 tree into index/worktree. restoration.txt is empty, proving equality before headers. Repaired setup completed successfully.'),indent=2)+'\n')
lines=[
'Added 13 missing first-line a-check refusal headers; every program body is byte-identical to 505835b7.',
'Delivery is one commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309 containing 505835b74876b61595864dcc5ddac6d40ba3767a plus these headers and review evidence; its SHA is in the handoff.',
'Actual branch compiler and unchanged pinned Gate.aCheck accept 58/58: 42 checked, 0 NotYet, 16 refused, 0 type error; oracle header tests, reader guard and vet pass.',
'Removed-header and wrong-reason mutants both exit 1 through the actual a-check predicate; the source file is restored and the final scan passes.',
'Not full-branch green: 13 inherited stale freshness readers and their coverage guard fail. No full gate or prior unrelated consumer/budget rerun is claimed.',
'',
'The explicit pinned base overrides the current-main boilerplate. origin/main advanced to 99c369e0 during fetch; this branch remains based exactly on dbdf977f5b. The conservative scope is the requested source change plus headers: no compiler/runtime/Go-test edits, no fixture moves, no unrelated test exclusions, and no counts changes relative to 505835b7. No cohere code was copied. No ahra command, PR or Loom submission was made.',
'',
'Inventory and real predicate',
'',
'paths.txt is every added, modified, copied or renamed .a in the change, derived from git diff --name-only --diff-filter=ACMR dbdf977f5b 505835b7 -- "*.a". It includes all 13 relocated refusal controls. There are no added or modified review/ .a files. Deleted old paths are not programs in the resulting tree. The final diff adds no other .a files.',
'',
'check-headers.py invokes unchanged Gate.aCheck from README-pinned commit fbac28c62493f27a788edc02a18bc8edb68de5da. gate-source.py is the exact fetched file, SHA-256 '+hashlib.sha256((out/'gate-source.py').read_bytes()).hexdigest()+'. Its own build step runs go build -o <tree>-binaries/adamic-acheck ./cmd/adamic on this working tree, then adamic-acheck c <path> for every program. The harness supplies logging and hard time limits without changing the predicate. This measures the checker/lowering/C-emission front end, not clang execution. before.json and final.json retain full real diagnostics, status, exit, duration and emitted-C size/hash.',
'',
'The pre-header predicate rejected exactly the 13 missing headers. After insertion it exits 0 for all 58. Three existing refusal headers were already accepted and remain unchanged. All 42 checked programs have no expected-error header; there are no NotYet or type-error cases.',
'',
'| File | Status | Refusal reason / checker code | Header added |',
'|---|---|---|---|']
for path,r in final['paths'].items():
 detail=r.get('reason') or r.get('code') or ''
 lines.append('| `'+path+'` | '+r['status']+' | '+detail.replace('|','\\|')+' | '+('yes' if path in changed else 'no')+' |')
lines += ['', 'Headers added', '', 'Each of the following 13 files gains exactly `// a-check: refused the non-null assertion !` as its first line. No type-error header was necessary. headers.json records every exact header and original body hash; headers.patch is the zero-context insertion diff.', '']
lines += ['- `'+r['path']+'`' for r in edits]
lines += ['',
'Body bytes, line positions and counts', '',
'body-proof.json and final-integrity.json compare the entire post-header body against git show 505835b7:<path>, not normalized text. All 13 comparisons are byte-for-byte equal. Header insertion moves each refusal location by one line, preserving its column and reason; diagnostic-shifts.json records every old/new location.',
'',
'The required exact-path/basename/family grep over *_test.go and testdata is retained in path-references.txt. Active Oracle readers are TestImpossibleNonNullFixturesAreRefused, TestPossibleNonNullAdamicAssertionsAreRefused and TestCheckedNonNullAdamicRefusal. Their .a cases check refusal kind, reason, fix and suffix without pinning line positions. The :1:13 pins belong to untouched .ts controls; the :47:16 pin belongs to untouched non_null_refused/initialized.a. No active positional expectation needed an edit. Historical source reports and captures remain observations of their original revision; fresh post-header diagnostics are in final.json.',
'',
'internal/oracle/counts.md is byte-identical to 505835b7, SHA-256 1c750ef4bdddda92d0e3314944234fe2fd7c89db486424720222a34b9b9219d8. No counts regeneration was run: only refused, uncounted controls gained comments. TestCountsAreRecordedTable passes.',
'',
'Header mutants', '',
'prove-headers.py edits non_null_refused/non_null_possible_read.a in place, runs the actual compiler and pinned Gate.aCheck on each mutant, and restores the exact original bytes in finally. No intentionally headerless .a mutant is left in review evidence.',
'',
'| Mutant | Actual result | Predicate catcher |','|---|---|---|',
'| Remove first-line header | Compiler still refuses the non-null assertion; a-check exits 1 | refused, expected checked |',
'| Replace reason with deliberately wrong slice5 header reason | Compiler still refuses the non-null assertion; a-check exits 1 | refused, expected refused deliberately wrong slice5 header reason |',
'',
'header-mutants.json contains commands and exact failure strings; removed-header.json and wrong-reason.json contain real diagnostics. The final complete scan runs after both restorations and accepts every file.',
'',
'Tests and commands', '',
'All test output goes directly to logs. GOMAXPROCS=4 and GOFLAGS preserve -buildvcs=false -trimpath and append -p=4; GOCACHEPROG is unset for the local Go cache. Every go test has -count=1 -json -timeout 90s and an outer 150s or 180s limit. A-check compiler invocations have 45s limits; build steps have 120s limits, with an outer 240s script limit. No test leaf added or edited in this pass; observed affected-test and guard times are in test-seconds.json.',
'',
'```sh',
'python3 review/compiler/loom-r-slice-5-ac/check-headers.py before',
'python3 review/compiler/loom-r-slice-5-ac/prove-headers.py',
'python3 review/compiler/loom-r-slice-5-ac/check-headers.py final',
"go test ./internal/oracle ./internal/ir -run '^(TestImpossibleNonNullFixturesAreRefused|TestPossibleNonNullAdamicAssertionsAreRefused|TestCheckedNonNullAdamicRefusal|TestCallTargetReaders)$' -count=1 -json -timeout 90s",
"go test ./internal/oracle -run '^(TestImpossibleNonNullFixturesAreRefused|TestPossibleNonNullAdamicAssertionsAreRefused|TestCheckedNonNullAdamicRefusal|TestCountsAreRecordedTable)$' -count=1 -json -timeout 90s",
'go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle',
'```',
'',
'Both Oracle/reader commands and vet exit 0. The final Oracle command, after mutant restoration, completes in 1.596s. TestCallTargetReaders passes in 2.74s. The initial Oracle .ts controls also run source Node and ASan/UBSan native observations; their largest observed leaf is 8.26s. No output comparison or diagnostic expectation was weakened.',
'',
'Inherited freshness failure found by the path audit', '',
'The prior relocation left 13 internal/fresh generated readers naming ../oracle/testdata/non_null_*.a instead of the files now under non_null_refused/. All 13 selected readers fail loading absent files; TestFreshCorpusUnitsCoverEveryProgram fails at generated index 550 for ../oracle/testdata/non_null_possible_read.a. The focused command exits 1 in 0.007s. These are not header-mutant kills and are not claimed green:',
'',
'```sh',
"go test ./internal/fresh -run '^(TestFreshCorpusUnitsCoverEveryProgram|TestFreshWrites____oracle_testdata_non_null_(possible_(read|typeerror)|refuse_(argument|checked|comparison|conditional|element|expression|logical|null|return|spread|undefined))_a_[0-9a-f]+)$' -count=1 -json -timeout 90s",
'```',
'',
'inherited-freshness.json proves every old path was already absent in 505835b7 and that fresh_test.go, corpus_units_test.go and corpus_coverage_test.go are unchanged from that revision. reader-failures.json and logs/fresh-path-readers.jsonl retain the exact failures. This pass does not modify another package\'s generated tests or suppress their coverage to force green. A freshness reader relocation update remains necessary before claiming the full branch is green.',
'',
'Setup and limits', '',
'The initial source application failed with No space left on device while creating internal/lower/testdata/placeholders; review output could not be created, so setup did not start. Removed only the disposable local product cache /home/agent/.cache/adamic-build/local (about 11 GB), restored the exact requested 505835b7 tree, and verified an empty restoration.txt diff before any header edit. Setup then succeeded. Source env: /workspace/adamic-tools/env.sh; GOPROXY=https://proxy.golang.org|direct; nproc=5; cpu.max=400000 100000, a four-CPU quota. Go 1.27.1, Node 24.19.0 and house clang 20.1.8 on Linux. Timing lines:',
'']
lines += ['- '+s for s in timings]
lines += ['',
'Not covered: whole packages, the full fast gate, ahra/Loom submission, a whole-tree scan beyond the requested changed .a inventory, or a repeat of the original compiler change\'s unrelated consumer suites and recorded semantic mutants. This pass adds no semantic compiler/runtime change; 505835b7\'s existing validation limitations and runtime-clearance report still apply. Its over-budget consumer tests were not rerun. The header phase is green; the inherited freshness failures above prevent full-branch clearance.']
(out/'REPORT.md').write_text('\n'.join(lines)+'\n')
summary=dict(base='dbdf977f5b8b0979d414ebe6f20e7aadd312e309',source='505835b74876b61595864dcc5ddac6d40ba3767a',files=58,counts={'checked':42,'NotYet':0,'refused':16,'type error':0},headers_added=13,predicate_exit=0,header_mutants_caught=2,header_sensitive_tests='pass',reader_guard='pass',vet='pass',counts_unchanged=True,program_bodies_unchanged=True,inherited_freshness_failures=14,full_branch_green=False)
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary))
