Built: scanner stop 1 now resolves unconstrained U and U | undefined per instantiation, including an already optional U.
Commits: early passing probe pushed as 51b73d89c62c4d49f69c89bf0f6d49962d75a5aa; final validation is in the commit containing this report.
Commands/results: exact probe and all requested fixtures match source Node, emitted JavaScript, release native and ASan/UBSan native; counts and vet pass.
Mutants: disable the mapper hook restores the exact original diagnostic; missing number becomes present zero is caught only by Node stdout with clean sanitizer execution.
Limits: no native scanner-slice acceptance, no fresh 123-family census and no full repository gate; Map union values and later scanner blockers are outside this change.

The requested priority item comes from origin/codex/stage3-scanner-proof at
3e3791e5, stage3/drivers/scanner/BLOCKERS.md and
probes/generic-optional-callback-result.a. The probe was copied byte for byte,
including its upstream utilities.ts:746-755 attribution. On main 71d7e491 it
failed before C emission at line 2:10 with a function returning U | undefined.
The original source on Node prints x. The passing implementation was pushed
immediately, before the remaining package validation, as requested.

The previous census branch already contained the required resolved-mapper hook
in 9642bb4, but current main did not. This unit brings that implementation onto
a separate branch from current main, without the structural-dispatch changes.
No cohere code was copied. It uses the existing checker mapper bridge. Generic
lowering reads the resolved signature's substitutions before the existing
signature inference. The old inference cannot recover U from an optional union
whose result has collapsed, or distinguish a U that already includes undefined.
Ordinary monomorphization and representations then handle each concrete result.
No constraint was added to U and no source annotation was erased.

The additional .a fixture keeps first<U> unconstrained and exercises number zero,
number present/missing, dynamic empty/present strings, and reference objects with
dynamic strings. optional<T> calls first<T | undefined>, proving nested caller
substitution and the collapsed union for number, string and object instantiations.
Source Node output, exit 0:

```text
0 7 undefined
|wordword|undefined
itemitem missing
0 undefined
uu|undefined
itemitem missing
```

Both accepted fixtures are registered with the ordinary three-way oracle. Counts
were regenerated: original probe allocations/frees 2/2, retains/releases 1/3,
peak 2; expanded probe 33/33, 25/48, peak 6; regions zero for both. No existing
count row changed. The runtime mutant changes all emitted absent maybe-number
literals into present zero. It compiles, exits zero with empty sanitizer stderr,
and differs solely on stdout. The mapper-disabled Go overlay mutant fails the
original fixture with its original lowering diagnostic, not a clang warning.

Toolchain setup from this same-main working session: Go ready 0.030s, Node ready
0.036s, submodules ready 0.082s, clang ready 0.301s, Markdown dependencies ready
1.158s, go build ready 37.562s, test binaries deferred, cache warm 37.680s,
done 37.714s. nproc 5; cpu.max 400000 100000. Toolchain versions Go 1.27.1,
clang 20.1.8 and Node 24.19.0. Source /workspace/adamic-tools/env.sh for every
build/test shell. Setup log /tmp/empty-array-setup.log.

Commands, with test output redirected directly to logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(generic-optional-callback-result|generic_optional_callback_results)' -v -count=1 -timeout 30m > /tmp/scanner-optional-focused.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/scanner-optional-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestGenericOptionalUndefinedMutant -count=1 -v > /tmp/scanner-optional-mutant.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/scanner-optional-mutant-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/generic-optional-callback-result' -count=1 -timeout 30m > /tmp/scanner-optional-mapper-mutant.log 2>&1
go vet ./... > /tmp/scanner-optional-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/scanner-optional-gate.log 2>&1
```

Focused oracle exits 0 (0.539s), counts exits 0 (19.694s), runtime-mutant test
exits 0 (0.815s, meaning the deliberately wrong binary was caught), and vet exits
0 with no diagnostics. Mapper-disabled mutant exits 1, as required.
Original source Node output is /tmp/scanner-optional-node.log. No native scanner
binary was built and no later scanner diagnostic is claimed cleared.
The original census's 123 uninstantiated generic declarations were not rerun;
this probe establishes concrete call behavior, not a new census count.

Final touched-package gate exits 0: internal/lower 21.905s and the complete
uncached internal/oracle suite 124.182s. gofmt reports no files and git diff
--check passes. Current origin/main remains 71d7e491. The full repository test
suite was not run; the complete oracle and touched lowering package were run.
