Audited a-check headers for the eep presence change; no headers needed.
Source e40bace8a8dae3a11fe5a537b717cf01cc22b36f; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309.
All seven .a files compile to NotYet stops; unchanged pinned Gate.aCheck accepts all seven.
Removed-header and wrong-reason mutants both caught on an independent temporary refusal control.
No full gate or Loom submission; no native rerun because no program or compiler bytes changed.

Branch compiler/loom-r-slice-7-ac starts at the explicitly requested dbdf977f5b base. Fetched origin, then cherry-pick --no-commit of e40bace8. No conflicts. The explicit pinned base takes precedence over the generic latest-main instruction. The resulting one commit contains exactly the prior change plus this audit evidence. No headers were added: README explicitly requires no header on NotYet stops. No source body changed and no diagnostic line moved. No test/report position expectation needed updating.

Read stage3/a-check-headers/README.md whole, its documented pinned Gate.aCheck and scan harness, CLAUDE.md and docs/front-door.md. No cohere code copied.

Inventory from git diff --name-only dbdf977f5b e40bace8 -- '*.a' covers additions/modifications anywhere in the tree, review included. Exactly seven lower fixtures, zero review .a files; source-carried review program snapshots end .a.txt. The new temporary control is retained only as .a.txt, so it adds no scan target. All seven files were read whole, compiled individually with the branch's own ./cmd/adamic compiler using adamic c, with a 60 s process limit and stdout/stderr logs. Full diagnostics and hashes: statuses.json. Compiler built with timeout 120 go build; house -trimpath -buildvcs=false flags preserved.

| File | Actual status | Seconds |
|---|---|---:|
| internal/lower/testdata/eep_presence/eep1m5z-93122fa_t_t5.a | NotYet | 0.033 |
| internal/lower/testdata/eep_presence/eep1m5z-93122fa_t_v1.a | NotYet | 0.029 |
| internal/lower/testdata/eep_presence/eep1m5z-93122fa_t_v2.a | NotYet | 0.05 |
| internal/lower/testdata/eep_presence/eep1m5z-93122fa_t_v5.a | NotYet | 0.028 |
| internal/lower/testdata/eep_presence/eep1m5z-classfeat_method_view.a | NotYet | 0.047 |
| internal/lower/testdata/eep_presence/eep1m5z-classfeat_static_virtual.a | NotYet | 0.031 |
| internal/lower/testdata/eep_presence/eep1m5z.a | NotYet | 0.029 |

Status counts: checked/clean 0, NotYet 7, refused 0, type error 0. Gate.aCheck deliberately calls all seven NotYet observations 'checked', so its result is checked=7, findings=0. NotYet reasons: five union/optional tuple lengths, one method sharing its name with a getter, one static field read with an earlier initializer calling this. statuses.json preserves each full reason. The earlier generic REFUSED label in the source-carried fixture runner includes CLI NotYet stops; it is not the compiler's Refused diagnostic class.

Headers added: none. Grepped eep_presence/eep1m5z references in *_test.go and testdata: the reader is internal/lower/eep_presence_test.go. It checks message text and the presence of main.a:, without a fixed line number. All source bytes still match e40bace8, so no position expectation changes are warranted. Runtime C and counts.md remain byte-identical; no counts regeneration required for this evidence-only pass. The prior one-time counts regeneration remains recorded in review/compiler/loom-r-slice-7/.

Predicate: fetched the README's exact fbac28c62493f27a788edc02a18bc8edb68de5da object. predicate.py loads its unchanged cloud/fast-gate/run.py and invokes Gate.aCheck, not a reimplementation. Its step runs the branch's real go build; spawn runs the real compiler for each path. timeout 300 python3 predicate.py; build steps bounded at 120 s, each compiler process at 60 s. predicate-scan.json: exit 0, all seven accepted.

No changed file qualifies for a refusal header, so removing a real changed-file header would be impossible. To prove the required predicate behavior without mislabeling NotYet, the script materializes a temporary .a index-signature control outside the repository (body retained as refused-control.a.txt). Real compiler diagnostic: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added. Matching header '// a-check: refused an index signature; use a Map, which keeps keys in the order they were added' passes. Remove that header: predicate exit 1, 'refused, expected checked'. Replace the reason with 'deliberately wrong reason': predicate exit 1, wrong reason rejected. All compile processes use the actual branch compiler, no synthetic diagnostics or compile-error kills. The temporary directory is removed automatically. mutants.json and predicate.log record both catches and the control pass. This is a control proof, not a mutant of one of the seven NotYet programs.

Exploratory control construction is retained candidly: the first draft had an unrelated TS2345 console number argument; after String conversion its numeric condition was supported; main's historical any witness gave NotYet. Those trials stopped at assertions before testing mutants and are recorded in predicate-first/second/third-attempt.log. The final index-signature control uses an explicit Refused entry in main's refusal table and completes both mutant catches.

Verification, all output directly to logs:
- timeout 110 go test ./internal/lower -run '^TestEEPPresence' -count=1 -json -timeout 90s: all eight pass, package 0.592 s. Individual seconds in test-seconds.json, all under 60 s on the four-CPU quota.
- timeout 110 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -json -timeout 90s: PASS, package 1.744 s.
- timeout 110 go vet ./internal/lower: PASS, no diagnostics. No new Go test files edited beyond the unchanged prior commit's tests.
- Exact byte comparisons confirm every fixture, all internal/ files, and counts.md unchanged from the stated prior/base commits. No new headers means no additional package reader or runtime behavior can be affected; prior focused Node/native/sanitizer/JS and downstream evidence is preserved unchanged in review/compiler/loom-r-slice-7/.

Setup: export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. PASS. nproc=5, cgroup cpu.max=400000 100000; GOMAXPROCS=4 for build/tests. Node 24.19.0, Go 1.27.1, clang 20.1.8. Timing lines (cumulative seconds): Node 0.022, Go 0.034, submodules 0.052, markdown ready 0.079 (step 0.011), clang 0.181, shared cache 0.495, go build 14.565, test binaries deferred 14.617, build cache warm 14.618, done 14.636. Complete lines are in setup.log.

No PR opened, no ahra CLI run, no Loom submission. Push only compiler/loom-r-slice-7-ac; final user report records the resulting one-commit SHA.
