Built: 20 original callable member pairs, representing 100 ranked candidate reads, with positive and wrong-arity fixtures.
Commits: base 926a1d39; this branch carries fixture/test commits only.
Commands: original-source verifier PASS; TestCheckedViewCallableShareC and Mutants PASS; share counts updater PASS. Required TestCountsAreRecorded updater FAIL on inherited fixtures (see validation below).
Mutants: each rank removes its callable read certificate in lowered IR; native and JavaScript execute exit 0 instead of pinned exit 70, so all 20 mutants are caught.
Uncovered: remaining share ranks, delegated Set/optional-host/binding families, and preparation boundaries are not certified.

Candidate reads are ledger weights, not measured production execution. The fixture preserves the original member signature and read path. Adjacent interface carriers and implementation bodies are deliberately reduced as in the lane harness; original aliases and numeric kind discriminants remain intact.

The independent upstream checker/source pin is 050880ce59e30b356b686bd3144efe24f875ebc8. Run prepare.cjs then verify.cjs with that checkout root. No cohere implementation code is copied.

Validation: `go test ./internal/oracle -run '^TestCheckedViewCallableShareC(Mutants)?$' -count=1 -v` passed in 20.067s, log /tmp/lane5-c-batch1.log. Native release, ASan/UBSan, successful-fixture LSan/count balance, JavaScript backend, and Node oracle ran. Both negative backend diagnostics are pinned byte-for-byte. `go test ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts` failed after 41.403s on inherited fixtures, log /tmp/lane5-c-required-counts.log; it did not rewrite counts. The scoped share updater appends only this share's measured rows.

Setup: Go 0.076s; Node 0.074s; clang 0.520s; markdownlint 0.981s; submodules 165.671s; go build 353.266s; warm 353.389s; done 353.416s. nproc=5, cpu.max=400000 100000. GOPROXY=https://proxy.golang.org|direct; environment /workspace/adamic-tools/env.sh.

Conservative assumption: no certification is claimed for intrinsic receivers, overloads, namespace callables, generic/rest signatures, or binding reads merely from a reduced replacement. They need their original contexts.
