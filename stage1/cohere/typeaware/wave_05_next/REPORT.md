Built: no-uncleared-race-timeout and raw symbol-ancestry facts; two new claims remain blocked and unported.
Commits: prior three complete at c5403d46; continuation claim d8b312d9; implementation d4320843.
Commands and outputs: local Python verifier PASS; bridge checker go test PASS 0.103s; touched bridge vet clean.
Mutants: timer verdict exits 0 and fails Go byte comparison; alias and suffix mutants fail direct checker assertions.
Not covered: process-exit and blocking-stream ports, full repository gate, emitted JavaScript, and the complete upstream matrix.

# Wave 05 continuation 1

The first three rules were already completed and pushed. After fetching all origin heads, selection used the combined 197-rule checker-dependent volume table, descending totals with lexical ties, excluding ports on main and the bridge branch and reservations in all origin claim files. Inventory and skip-types references were not counted as ports. The first three available candidates were:

- `nexus/correctness-no-process-exit-after-output`
- `nexus/correctness-no-uncleared-race-timeout`
- `nexus/correctness-require-blocking-standard-streams`

All have zero compiler and repository findings in the recorded volume table. Claim-only commit `d8b312d9` was pushed before writing implementation. No further rules were taken.

## Completed timer port

`no_uncleared_race_timeout.a` owns all rule judgments. The new Go `symbol_ancestry.go` supplies raw declaration ancestry, source-module/library flags, module keywords and alias information. Its decoder is `symbol_ancestry.a`; the wire format is documented beside it. The only shared implementation edit is a single physical switch-registration line in the bridge. The shared registration generator and test harness are untouched. The wave-owned driver and verifier are in this directory.

The default production Go rule is invoked unchanged through a separate compiler host and AST walk in `testdata/oracle.go`. It imports no bridge implementation. Comparison includes complete finding ranges, message IDs/text, fixes and suggestions, retaining duplicate diagnostics. This rule has no fixes or suggestions, and those zero fields are held by comparison.

| Population | Findings | Identical bytes | Checks |
| --- | ---: | ---: | --- |
| 22 DOM controls | 12 | 7,800 | Normal, ASan, UBSan, LSan |
| 22 Node controls | 12 | 7,800 | Normal, ASan, UBSan, LSan |
| Frozen 287-file repository | 0 | 18,485 | Normal, ASan, UBSan, LSan |
| TypeScript compiler, 77 files | 0 | 5,318 | Normal, ASan, UBSan, LSan |

Controls cover inline and const-bound timeout promises, lost statements and expression bodies, unread initializers and assignments, retained/cleared handles, shorthand reads, shadowing, nested functions/classes, globalThis, void, multiple timers, helper declines, nonlibrary receivers, return/assignment-value uses, and Unicode/CRLF. Both DOM globals and Node's `declare global` function/namespace merge are checked. Sanitizers are clean. Querying the new fact after releasing the program exits 70 with exactly `adamic: panic: invalid or released checker handle`.

Reproduction, after sourcing `/workspace/adamic-tools/env.sh`:

```bash
WAVE05_NEXT_SKIP_BENCH=1 \
ADAMIC_WAVE05_REPOSITORY_MANIFEST=/workspace/wave-05-repository.manifest \
ADAMIC_WAVE05_COMPILER_MANIFEST=/workspace/wave-05-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-05-typescript \
python3 stage1/cohere/typeaware/wave_05_next/validate.py \
  > /workspace/wave-05-next-verified.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-05-next-checker.log 2>&1
go vet ./bridge/tsgo/... > /workspace/wave-05-next-vet.log 2>&1
```

The manifests are the same frozen inputs as the original wave. Cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, and TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8` are unchanged. The earlier successful cloud setup and its 82s timing are retained in the original wave evidence; `nproc` is 5, cgroup CPU capacity is four cores.

## Mutants

- Reversing the native `lost` verdict builds, exits 0 with empty stderr, and differs from independent production Go findings. It is caught by bytes only.
- Disabling the new suffix guard causes `TestSymbolAncestryAndAliasModes` to fail with `accepted suffix: <nil>`.
- Disabling alias resolution causes that test to fail because the resolved declaration remains `ImportEqualsDeclaration` rather than `FunctionDeclaration`.

The local verifier checks released-handle refusal and preserves its direct logs. The original wave's bridge lifecycle and sanitizer mutants were already validated; they were not rerun wholesale in this continuation.

## Exact blocker and stop

The other two claimed rules require forward traversal of a control-flow graph with expression/statement hooks. Reusing this branch's native CFG builder through a derived class failed before native generation. The minimal [cfg_probe.a](gaps/cfg_probe.a), containing no new rule judgments, is checked by the local verifier and produces:

```text
adamic: /workspace/adamic/stage1/cohere/typeaware/bindings.ts:49:33: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this
```

```bash
/workspace/wave-05-next-validation/adamic build \
  stage1/cohere/typeaware/wave_05_next/gaps/cfg_probe.a \
  -o /workspace/wave-05-cfg-probe \
  --tsgo /workspace/wave-05-next-validation/checker.a
# exit 1; refusal above, no native executable
```

This is the observed blocker for the attempted CFG integration, not proof that every alternative implementation is impossible. The process-exit and blocking-stream rules are **unported**. Their positive controls, corpus equivalence, rule mutants and timings are therefore **not claimed**. Under Ahra's instruction to stop on other blockers instead of editing shared files, no shared bindings, compiler, generator or test harness was changed to repair it. The reservations are marked blocked in the claim file.

## Timing and limits

Three alternating Go/native count-only rounds were rerun separately after all build/test jobs finished. Medians measure complete process loading, parsing, rule execution and teardown:

| Population | Go | Native | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.123211s | 0.233578s | 1.90x |
| Compiler | 0.288553s | 1.591788s | 5.52x |

These are zero-finding fast-path measurements. Native issued zero checker queries on these corpora, so they do not measure active fact-query performance. The timer port is slower than Go in this measurement. Raw rounds are in [timings.log](evidence/timings.log).

The complete upstream fixture matrix and emitted-JavaScript comparison were not run. The shared `.a`/JavaScript harness work was left to its owner. Pinned cohere's CLI still cannot lint `.a`; the existing scratch formatting overlay formats them without editing the submodule. Production source format checks are clean; the deliberately refused fixture remains in `gaps`. Source hashes, compressed oracle/mutant streams, checker test logs, refusal logs and sanitizer comparisons are retained in [evidence](evidence).

Automatic review rejected one earlier driver-generation command after interpreting a read of the completed driver as its replacement. Read-only inspection proved that driver unchanged; a separate command created the new driver explicitly. This was resolved without modifying the completed driver, and nothing remains blocked by automatic approval review.
