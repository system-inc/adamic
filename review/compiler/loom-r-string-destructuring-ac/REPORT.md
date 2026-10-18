Verified every .a added by c7302ad4 with this branch's own compiler; both pass the actual pinned a-check predicate without headers.
Source c7302ad4f9b866f085fbe0f5b69a9a9c3fbca79c; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; one commit on compiler/loom-r-string-destructuring-ac.
Commands/results: adamic c for both files, unchanged Gate.aCheck, focused lowering/oracle/flow/counts tests, reader guard and vet all passed; raw logs and commands are retained here.
Removed-header and wrong-reason mutants on an existing refused control each fail Gate.aCheck with exit 1; restored control's pinned-position test passes.
Not covered anew: full gate, all repository .a files, the 52-reader sweep or the four lowering/count mutants already recorded in c7302ad4; no header was needed or added.

The new branch starts at the exact requested dbdf977f5b, despite origin/main having moved to 99c369e0 during fetch. Applied only c7302ad4's own change with git cherry-pick --no-commit, without conflicts. Compiler, tests, fixture bytes, runtime and counts are identical to c7302ad4; this pass adds review evidence only. Conservative interpretation: obey the measured NotYet/refusal distinction, rather than inserting an incorrect expected-refusal header on a NotYet program. This assumption is also in the commit message.

Read CLAUDE.md, docs/front-door.md and the complete stage3/a-check-headers/README.md. Its pinned Gate.aCheck was fetched via devtools/fast-gate at fbac28c62493f27a788edc02a18bc8edb68de5da. gate-reference.py.txt is the exact Git blob, not a replacement predicate. The harness calls its unchanged Gate.aCheck method. The initial local wrapper had an incorrect spawn signature and failed before scanning; a-check-harness-initial-error.log.txt retains that error. Correcting only the wrapper produced the results below.

| Changed program | Real compiler status | Header added | A-check |
|---|---|---|---|
| internal/oracle/testdata/string_destructuring.a | checked, exit 0 | none | pass |
| review/compiler/string-destructuring/unsupported-member.a | NotYet, exit 1 | none | pass |

Totals: checked 1, NotYet 1, refused 0, type error 0. Gate.aCheck calls clean and NotYet outcomes checked, so its own summary is checked 2. The finer status table is in status-table.json. There are exactly two .a paths in the source commit's own diff; no additional .a file is introduced by this pass.

The complete NotYet reason is:

    stage 0 can't lower string member 'charCodeAt' in object destructuring; call methods on the saved string receiver or wrap the call in an arrow yet

It occurs at review/compiler/string-destructuring/unsupported-member.a:1:9. It is not an Adamic 0.1 refusal or a checker error. Adding a refusal header would make a-check fail, because the predicate requires the actual outcome to be refused when that header is present. Both complete files remain byte for byte identical to c7302ad4, not only their bodies. body-preservation.json records source blob IDs and SHA-256 hashes. Headers added: none. Position expectations changed: none. path-readers.log.txt records the repository-wide *_test.go/testdata path search; only the oracle fixture registration matched.

There was no new expected-error header to remove, so the two required predicate mutants used an existing, explicitly refused control: internal/oracle/testdata/non_null_refused/initialized.a. Its existing first line is // a-check: refused the non-null assertion !. The same branch compiler first confirmed the valid header passes. Then the actual file's header was removed and independently replaced with // a-check: refused deliberately-wrong-reason. Both actual compiler invocations still refused the non-null assertion, while unchanged Gate.aCheck rejected the expectations:

- removed header: exit 1, refused, expected checked;
- wrong reason: exit 1, refused, expected refused deliberately-wrong-reason.

This exercises real compiler diagnostics and the actual gate method, rather than a copied predicate or synthetic diagnostic. a-check.py restores the entire control file in finally and asserts byte equality. The control remains outside the delivered diff. TestNonNullAdamicRefusal was rerun after restoration and passes its :3:25 expectation (0.12s). No pinned position moved in any delivered file.

Validation used source /workspace/adamic-tools/env.sh, GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1 and bounded commands. validate.py and validation.json retain exact command arguments and results, with output in individual .log.txt files. All seven validation commands exited 0:

- Guards across internal/ir, internal/lower, cmd/adamic-gate and internal/oracle: sorted IR maps, fixture registration, test parallelism, duplicated product declarations and runtime-set witnesses.
- internal/lower -run '^(TestStringDestructuring.*|Test.*Destructur.*)$' -count=1 -timeout 90s -v -parallel 4.
- internal/oracle source-report selector for the string fixture, -count=1 -timeout 90s -v -parallel 4. The log proves TestNativeAgreesWithNode_097/internal/oracle/testdata/string_destructuring.a actually ran: 5.84s, with three native and two Node cache misses. Source Node, both backends, sanitizers and leak checks passed.
- internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -v.
- internal/flow -run '^(TestFlowCorpusRemainder|TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusSetupIsShared)$' -count=1 -timeout 90s -v. The string fixture was included.
- internal/oracle -run '^TestCountsAreRecorded_001$/internal/oracle/testdata/string_destructuring[.]a$' -count=1 -timeout 90s -v -parallel 4. Existing measured count product and recorded row agree.
- go vet ./internal/lower ./internal/oracle, bounded by the harness's 180s subprocess limit.

The separate restored-control command was timeout 120 go test ./internal/oracle -run '^TestNonNullAdamicRefusal$' -count=1 -timeout 90s -v -parallel 4. It passed. No new or edited Go test is added beyond the unchanged source commit's tests. No test leaf exceeded 60s and no over-budget leaf was rerun. All other packages' reader checks from c7302ad4 remain archived evidence of the byte-identical compiler; this pass does not claim a fresh 52-package sweep.

Counts.md is byte-identical to c7302ad4. No fixture body changed, no new fixture was added, and no count regeneration was needed. The source commit's one added row remains 34 allocations / 34 frees / 18 retains / 46 releases / 6 peak / 0 in regions. Its existing full regeneration evidence is preserved. Runtime changed lines: none. No IR map, runtime feature set or compiler-consuming package was added. No code was copied from cohere.

Setup: exported GOPROXY='https://proxy.golang.org|direct', ran timeout 600 bash cloud/setup.sh and sourced its printed env file. Succeeded with Go 1.27.1, Node 24.19.0 and house clang 20.1.8. nproc=5, cpu.max=400000 100000 (four CPU quota). Timing lines in setup.log.txt: Go 0.114s; submodules 0.123s; Node 0.128s; markdown install step 0.043s, ready 0.297s; clang 0.609s; shared cache 1.323s; Go build 27.002s; test binaries deferred 27.170s; cache warm 27.173s; done 27.222s. No setup workaround was needed.
