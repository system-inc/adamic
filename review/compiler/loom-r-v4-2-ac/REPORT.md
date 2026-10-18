Built: a-check audit of every V4-added or modified .a file, including review evidence, on compiler/loom-r-v4-2-ac.
Commit: one commit from exact dbdf977f5b8b0979d414ebe6f20e7aadd312e309 carrying e5a4b1369fc76d0a42bb573362d99360381f83a4; pushed SHA is in the final response.
Commands and outputs: branch compiler scanned all 47 programs; unchanged Gate.aCheck accepts all 47 with exit zero; focused readers, reader guard and vet results follow.
Mutants: removing one refusal header and independently replacing its reason both fail the unchanged predicate; original bytes restored.
Not covered: full tree/full gate, ahra loom check, unrelated historical a-check findings. No compiler, runtime, program body, test expectation or counts edits.

Read stage3/a-check-headers/README.md whole and used its exact pinned predicate, Gate.aCheck at fbac28c62493f27a788edc02a18bc8edb68de5da. That historical object was absent locally; bounded git fetch origin <pin> succeeded (gate-fetch.log). No gate source was changed. Its complete file SHA256 is c81d5f29ae6a06c5671fd8caf961828bc040c1f565df0c0f2cb660ce28a29317. check-headers.py loads that file, substitutes the already-built branch compiler for the executable path, and leaves classification, first-line parsing, reason/code matching and refusal behavior unchanged. It uses actual in-place compiler diagnostics for the baseline and captured actual diagnostics for the two header-predicate mutants, exactly as the README describes. It does not invoke the rest of the historical gate.

Fetched origin, created compiler/loom-r-v4-2-ac at the exact requested dbdf977f5b base, and applied e5a4b136 with cherry-pick --no-commit. The explicit exact-base request overrides the generic current-main instruction. There were no conflicts. The conservative choice is to retain an existing header when the unchanged predicate accepts it, rather than replace an accepted diagnostic substring with a longer spelling. In particular, the factory refusal header's text is an exact substring of the actual reason and the predicate accepts it.

There are 47 added or modified .a paths: 41 added, six modified. The inventory uses git diff --name-only --diff-filter=AM dbdf977f5b -- '*.a'; git diff --name-status confirms there are no omitted renamed .a paths. It includes all nine review .a witnesses. Compiled each path with this branch's own compiler, /tmp/adamic-v4-ac c <path>, with a 60 s process bound and four workers. The final predicate independently runs the real compiler again, sequentially, on all 47 paths. statuses.json records full diagnostics, exit codes, first lines, compile seconds, source SHA256 and body SHA256. status-table.csv and status-table.md give every path and observed status.

| Status | Files |
|---|---:|
| checked, C emitted | 40 |
| NotYet | 1 |
| refused | 5 |
| checker error | 1 |
| total | 47 |

Headers added: zero. The six negative files already have correct first-line headers in e5a4b136. All 47 whole files, including every program body and existing header, remain byte-for-byte identical to that commit. Each statuses.json row verifies this. The review construct-misfit.a stops with “stage 0 can't lower new an Identifier yet”; it correctly has no header. The other eight review programs emit C. No NotYet was silently counted as a refusal.

Existing negative headers retained:

| File under internal/oracle/testdata/views_v4_refused/ | First line |
|---|---|
| construct-no-signature.a | // a-check: type error TS7009 |
| factory-165-wrong-arity.a | // a-check: refused readonly field onEmitNode becomes writable |
| view-parameter-wrong-members.a | // a-check: refused an unproven escaping callable read of viewed.run |
| view-parameter-wrong.a | // a-check: refused an unproven escaping callable read of viewed.run |
| view-result-wrong-members.a | // a-check: refused an unproven escaping callable read of viewed.run |
| view-result-wrong.a | // a-check: refused an unproven escaping callable read of viewed.run |

Position-sensitive readers were searched by full paths and negative basenames in *_test.go and testdata. path-readers.log and negative-readers.log retain findings. No position expectation changed: no line was inserted. checked_views_v4_construct_refusals_test.go already strips the existing TS7009 header before asserting no-construct.ts:5:37; the NotYet construct witness stays at construct.ts:7. v4Lane5Source already strips an a-check first line before boxed.a:5 and wider-write.ts:6:48 assertions. Eight exact reader tests rerun these paths and both source-profile/native boundary semantics.

Predicate proof: gate-baseline.json records exit zero for all 47. header-mutants.json records both failures on internal/oracle/testdata/views_v4_refused/view-result-wrong.a. Removed-header gives “refused, expected checked”; wrong-reason gives “refused, expected refused deliberately unmatched V4 reason”. The compiler diagnostics are captured from the unchanged actual source, while the unchanged predicate rereads each mutant first line. A finally block restores the original file; statuses.json confirms all source bytes against e5a4b136 afterwards. No mutant program is left in the tree.

Commands, with output directed to the evidence logs:

    source /workspace/adamic-tools/env.sh
    timeout 120 go build -o /tmp/adamic-v4-ac ./cmd/adamic
    timeout 180 python3 review/compiler/loom-r-v4-2-ac/check-headers.py
    timeout 180 go test -c -o /tmp/v4-ac-oracle.test ./internal/oracle
    timeout 95 /tmp/v4-ac-oracle.test -test.run '^<reader>$' -test.count=1 -test.timeout=90s -test.parallel=1 -test.v
    timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -json
    timeout 180 go vet ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/load ./internal/lower ./internal/native ./internal/native/runtimeproducts ./internal/oracle ./stage1/cohere/json

The test binary runs from internal/oracle; each focused leaf has a separate log and a 95 s outer bound. At most two focused leaves run concurrently on the four-CPU quota. No new or edited _test.go file exists beyond the carried V4 change. No counts regeneration is needed: no fixture bytes or compiler/runtime lines changed, and internal/oracle/counts.md is identical to e5a4b136. The previous V4 report and validation remain carried, with this focused header audit as additional evidence.

Setup passed: GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh, then source the printed env.sh. setup.log retains all timing lines: Node 0.019 s; Go 0.020 s; markdown ready 0.059 s; submodules 0.066 s; clang 0.234 s; shared cache 0.615 s; Go build skipped by validated warm stamp 3.839 s; tests deferred 3.841 s; cache warm 3.842 s; done 3.869 s. nproc=5, cpu.max=400000 100000, house clang 20.1.8, Go 1.27.1, Node 24.19.0. No cohere code was copied.

Final focused results: eight pass, no skips, maximum 5.612 s. TestCallTargetReaders passes in 2.05 s. Vet passed with empty output. No source whitespace issue or new test leaf was introduced.

| Path reader | Command seconds |
|---|---:|
| TestV4EscapeAdapterConstructSignatureRefusal | 3.144 |
| TestV4EscapeAdapterWithoutConstructSignatureRefusal | 0.165 |
| TestV4Lane5BoxedParameter | 5.612 |
| TestV4Lane5BoxedParameterMembers | 5.610 |
| TestV4Lane5BoxedResult | 5.207 |
| TestV4Lane5BoxedResultMembers | 5.251 |
| TestV4Lane5BoxedAdamicRefusal | 3.257 |
| TestV4Lane5FactoryWiderWriteRefusal | 2.979 |

The exact unchanged predicate method is retained as a-check-method.py.txt. Only this evidence directory differs from e5a4b136. Headers added: none; retained and verified: six. Full per-file status table: [status-table.md](status-table.md); complete diagnostics and body hashes: [statuses.json](statuses.json).
