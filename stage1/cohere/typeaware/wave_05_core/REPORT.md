Built: require-await, symbol-description and valid-typeof as numeric handed-node .a callbacks; React claims parked.
Commits: claim fe82d9d4c pushed before code, rebased claim a39b04bec; implementation 3ffd452610d9db97e64550abc20b12abe35edd6b.
Commands and outputs: validate.py PASS 106 controls, 76 findings, complete suggestions, both frozen corpora, sanitizers and released handles; landing revalidation recorded below.
Mutants: all three rule verdict mutants caught only by production Go bytes; two bridge guard mutants and three numeric metadata mutants caught by their assertions.
Not covered: parked React analysis, require-atomic-updates, full option matrix, full repository gate and shared emitted-JavaScript comparison.

## Implementation and selection

The three reservations were published before code. [selection.json](selection.json) records the 529 origin refs, 31 unique claim blobs and combined by-volume ranking. Remaining React rules need JSX or native analysis; require-atomic-updates uses control-flow and captured-binding analysis. These three selected rules need neither high-level IR nor SSA/capture modules. The previous six implemented rules are retained, making nine completed wave-05 ports. The React claims remain explicitly parked, not ported: globals awaits JSX/root detection through area/stage1-lint; immutability and no-deriving-state-in-effects await #dnv6f2c HIR/SSA/capture analysis plus JSX.

Each rule owns rule.a and rule.json. Numeric listeners are require-await [263,219,220,175,178,179,177], symbol-description [214], and valid-typeof [227], from the pinned parser enum. The private driver indexes callbacks by kind and passes its cached Node. Parser kind strings are converted once in the adapter; rule files never compare kind strings or refetch parser nodes. No shared registration generator or harness was changed.

require-await implements function boundaries, async generators and empty bodies, await/for-await/await-using, contextual thenable exemptions (including generic calls and heritage), exact function heads and async-removal suggestions including ASI repair. symbol-description follows production Go's first declaration-file test for Symbol. valid-typeof checks both operands, cooked strings/templates and non-string literals, and preserves the undefined-to-string suggestion. Default rule options are used.

Two isolated raw checker questions expose declared generic call signatures and type projections, with native decoding and rule judgments in .a. Their contracts are [declared_call_signature.md](declared_call_signature.md) and [type_projection.md](type_projection.md). Each adds one physical line to facts.go; no Go lint verdict is exported. No protected compiler file changed.

## Validation

The independent oracle runs the unmodified production Go rules with a separate loader and AST walk, without the bridge. It compares canonical ranges, IDs, complete messages, fixes, suggestions and edits. All 106 generated controls parse successfully in Go. Findings are 63 require-await, three symbol-description and ten valid-typeof; 63 async-removal suggestions and two typeof suggestions match completely. All automatic-fix fields are empty in production Go and native. Controls include production require-await fixtures and added generic, heritage, ASI, Unicode, shadowing and literal cases. The verifier generates .a controls outside the repository.

Both frozen corpora match Go normally and under ASan, UBSan and LSan: repository 287 files (212 .a, 75 .ts), 18,485 identical bytes; TypeScript compiler 77 files, 5,318 identical bytes. These three rules have zero findings on both corpora, so positive controls provide verdict coverage. Complete canonical outputs and source hashes are retained in evidence.

The three native verdict mutants independently invert the Symbol declaration-file predicate, valid-typeof string membership, and require-await await detection. Each compiles, exits zero with empty stderr, and is caught only by differing Go bytes. The raw-question mutants remove declared-signature suffix validation or projection canonical-ID validation and are caught by invalid-question tests. Released-handle probes for both new questions exit 70 with exactly `adamic: panic: invalid or released checker handle`. Numeric metadata mutants are caught by production-enum/listener comparison; these are metadata checks, not verdict mutants.

Commands (all test output redirected directly to logs):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_05_core/validate.py > /workspace/wave-05-core-final.log 2>&1
python3 stage1/cohere/typeaware/wave_05_core/check_metadata.py > /workspace/wave-05-core-metadata.log 2>&1
go test ./bridge/tsgo/... -count=1 -timeout=15m > /workspace/wave-05-core-bridge.log 2>&1
go vet ./bridge/tsgo/... > /workspace/wave-05-core-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /workspace/wave-05-core-node-fresh.log 2>&1
python3 stage1/cohere/typeaware/wave_05_core/bench.py > /workspace/wave-05-core-timings.log 2>&1
```

Bridge tests PASS (66.392s bridge, 0.122s checker), including the new raw-fact test. Fresh filtered Node oracle and one-byte oracle mutant PASS (1.393s; native 0 hits/28 misses, Node 0 hits/19 misses). Vet and diff checks pass. Full repository gate was not run. Formatting used a wave-owned scratch formatter overlay for .a modules; no shared formatter or harness was edited.

## Timing and provenance

Three alternating isolated fresh-process rounds measured the three-rule aggregate. Final-base compiler native median 2.014132281s versus Go 0.322512154s (6.245x slower); repository native 0.289423044s versus Go 0.140517868s (2.060x slower). The earlier pre-rebase measurements are retained separately in timings.json; landing4-timings.json records the final-base rounds. Both sides report zero findings, native zero checker queries, on these corpora. These do not measure positive-query throughput or prove the cause of the remaining native overhead. Rebase revalidation overlaps other tests and is not benchmark evidence.

Pins: cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Setup previously passed in 82 seconds, nproc 5 (four cgroup CPU cores). Frozen manifests and hashes are retained.

## Landing

Main advanced to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 before publication. The requested rebase applied without conflicts. All nine completed rules are revalidated on that base, with both corpora, sanitizers, verdict mutants and released handles. The filtered uncached Node oracle and all bridge packages are rerun. First-wave TestWave05AgreementAndMutants PASS 139.908s; output and timer validators PASS; core validator PASS. Bridge packages PASS (93.163s bridge, 0.197s checker); uncached filtered Node oracle PASS 1.885s, native 0 hits/28 misses and Node 0 hits/19 misses. Numeric metadata, prior six listeners and vet also PASS. Exact landing outputs are in evidence/landing4-*.log. This worker publishes only codex/typeaware-wave-05, using an exact lease on its previous owned tip after the requested rebase. No main or area branch is pushed.

The parked React probe is historical evidence, not React rule parity. Complete upstream option matrices, automatic application of suggested edits, and shared emitted-JavaScript comparison are outside these checks. No shared finding-model migration SHA has been supplied.
