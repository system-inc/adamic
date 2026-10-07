Built: three more Nexus correctness rules in native Adamic .a files, bringing wave 24 to six ports.
Commits: original ports 406e7546; next claim 1d3d2b3f; implementation is the commit containing this report.
Checks: 57 controls / 65 findings and both frozen corpora match Go byte for byte; native sanitizer, checker, bridge and refusal gates pass.
Mutants: exit, timer, blocking and three raw-question mutations exit 0 and differ only in diagnostic bytes; released queries refuse, and bridge ownership/refusal mutants pass.
Uncovered: full repository gate, full upstream fixture matrix, production CLI .a lint and shared emitted-JavaScript rule/profile validation.

## Claim and selection

The original three ports were finished, tested and pushed as `406e7546846aeb58fb918ad4668f09f131a78e1e` before this batch.
Fetched all origin branches: 325 refs. Main was `ef3d907ecdc4c771b016f7d9c52372def057a340`; the bridge branch was `5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`.
Compared VOLUME_REPORT's linked checker-only counts, sorted by descending compiler-plus-repository volume and lexical ties, against the existing 26 ports and 33 distinct Markdown claim blobs across all refs.
The first eligible entries were full-ranking positions 122, 123 and 124, each with zero recorded findings:

1. `nexus/correctness-no-process-exit-after-output`
2. `nexus/correctness-no-uncleared-race-timeout`
3. `nexus/correctness-require-blocking-standard-streams`

Updated `claims/wave-24.md`, committed `1d3d2b3f8d374e90080a714c4cf11fc6ffca62bd`, and pushed before writing implementation code. No further rules were claimed.
The selection snapshot, exclusions and claim locations are in [evidence/selection.json](evidence/selection.json).

## Implementation and ownership

Each rule has its own .a file. The wave-owned suite constructs parser/scanner/bindings before handing them into rule constructors, avoiding the documented stage-0 nested-constructor gap.
New helpers and validation live in this directory. No shared registration generator, existing shared test harness, protected compiler file, existing rule port or submodule was changed.
The only existing file changed in this batch is wave 24's own `bridge/tsgo/checker/wave_24_questions.go`, adding one map entry per question/alias and extending its mode guard. The C archive registration from the original batch remains unchanged.

`no_uncleared_race_timeout.a` checks library Promise/race declarations, ambient timer declarations, executor-local calls, const promise bindings and whether the timer handle is ever read. Symbol identities distinguish shadowed and shorthand bindings. Every report is decided natively.

`no_process_exit_after_output.a` identifies platform writes and exits from declaration ancestry, follows one resolved callee where the production rule does, and walks native syntax paths carrying the minimal set of try chains for prior writes. Catch entry removes writes made inside its try. Nested functions have separate roots; unrecoverable exits in catch bindings or unrecorded syntax decline their root.

`require_blocking_standard_streams.a` computes entry status and the transitive blocking closure from raw import edges. It follows local called/handed-on functions, ordered blocking calls and awaits, and reports one entry finding with the first qualifying exit and exact count/reason text. It retains the production source-text `exit` prefilter.

`path_graph.a` builds syntax-only paths for conditionals, short circuits, loops, switches, try/catch/finally, returns, throws and jumps. Rule predicates, state transitions and reports remain outside the builder in Adamic. For-header separators are scanned outside parsed expression spans, so comments and nested expressions do not supply false separators.
The comparator exposed that production Go emits no finding for a write in the finally crossed by a break before a later exit. That omission is deliberately retained; the rule oracle is unmodified Go cohere, not a proposed semantic repair.

## Raw bridge questions

Matching Go and .a files implement:

| Question | Raw fields after the version/question header |
| --- | --- |
| symbol-provenance, symbol-provenance-alias | Symbol ID and flags; declaration count; per declaration: source path, declaration/library/external flags; ancestry with kind, name, UTF-8 span, NodeFlags and global-augmentation mark |
| resolved-call-origin | Return type flags; declaration present; declaration path/kind/span, external-module mark, function flags and body present |
| program-module-edges | Source-file count; each path, declaration-file mark, computed-load mark, resolved runtime import edges |

Symbol alias and shorthand resolution use the checker; opaque IDs belong to the live program. Module resolution uses the compiler's existing public APIs. Go returns no lint classification, reachability judgment, message, fix or suggestion.
Native decoders use the existing framed protocol, including strict integer/boolean/schema/trailing-data guards. The direct checker test pins Never, Ambient, Const, Function and async/generator flags, verifies positive declaration/signature/import facts, and rejects mismatched questions.

The independent oracle in `testdata/oracle.go` constructs its own compiler program and AST walk, imports no bridge code and calls the three unchanged production Nexus rules. All three rules have no production fixes or suggestions; the full protocol still compares their zero counts and every other diagnostic field, including byte spans, IDs, text, duplicates and file headers.

## Commands and observed output

Toolchain setup succeeded again in the resumed environment:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (19s)
setup: done in 19s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0; `nproc` prints 5, CPU quota is four cores.
Cohere remains `715ba94f3608a6500086b1076ce5cb7e51b836db`; its typescript-go remains `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
Compiler source is TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.
The manifests are exactly the frozen validation-coverage roots, with source hashes recorded here. Declaration roots are retained. No new port files are added to those populations.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave-24-next/validate.py \
  /workspace/wave-24-next-final --stage0 /workspace/wave-24-final-complete/adamic \
  --compiler /workspace/wave-24-corpus > /tmp/wave-24-next-final.log 2>&1

go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave-24-next-checker-final.log 2>&1
go test ./bridge/tsgo/... -count=1 -v > /tmp/wave-24-next-bridge-final.log 2>&1
go test ./stage1/cohere/typeaware \
  -run 'TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags' \
  -count=1 -v > /tmp/wave-24-next-refusals-final.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave-24-next-vet-final.log 2>&1
```

The stage0 path is the previously built compiler; no stage0 code changed. It can be rebuilt with `go build -o /tmp/adamic ./cmd/adamic`, then supplied to `--stage0`.
The validator builds normal/sanitized archives and native suites, generates .a controls plus ambient declaration fixtures, builds the Go oracle with an overlay, compares diagnostics and compiles/runs every mutant. Subprocess stdout/stderr go directly to numbered files, never through a pipe.

Final results:

| Population | Findings | Identical bytes, normal and sanitized |
| --- | ---: | ---: |
| 57 controls | 65: 33 per-exit, 7 timers, 25 blocking entries | 39,673 |
| Repository, 287 roots | 0 | 18,485 |
| Compiler, 77 roots | 0 | 5,010 |

The controls exercise ordering, branches, negation/short circuits, loop backedges and comments, labels/switches, catch reset and inherited writes, abrupt finally paths, dead code, local/nested/async/generator/never callees, platform streams and look-alikes, lost/reachable timer handles, shadowing, shorthand reads, const versus let timeout promises, computed module loads, Nexus declaration imports, shebangs, Unicode/CRLF and escaped identifiers.
Native normal and ASan/UBSan/LeakSanitizer runs match complete Go bytes with empty native stderr. Zero-finding corpora alone are not taken as proof; every rule has positive controls.
Checker PASS 0.133s, also PASS 0.157s inside the bridge gate. Bridge PASS 70.341s; frame/refusal/flags PASS 38.691s. Vet and gofmt logs are empty.
The final provenance query after release exits 70 with exact `adamic: panic: invalid or released checker handle` stderr. The bridge regression also proves that ownership and lifecycle guards can fail.

## Mutants and catchers

All six new mutants compile, exit 0 and have empty stderr. Only the independent production Go diagnostic bytes catch them. Offsets are zero based in the final controls output:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| exit | Ignore the one-level writing callee and consider only direct writes | 19,374 |
| timer | Reverse the lost-handle predicate | 31,073 |
| blocking | Reverse the entry/import exemption | 672 |
| provenance-question | Mark declaration files as non-declaration files | 147 |
| callee-question | Add Never to all resolved return flags | 19,374 |
| modules-question | Drop the computed-load seed | 4,849 |

The bridge gate reran the input/output length +1 mutants (ASan), stale registry retention (stale-handle assertion), source-file position substitution (byte oracle), link opt-in removal (refusal), missing C output free (LSan), and region result allocated on the heap (LSan).
Refusal tests reran wrong-kind and unknown-question guard removals, both caught by the missing required panic 70. The 19 wire mutants reran and each panicked 70: empty, length-negative, length-short, trailing, version, question, not-integer, rounded-integer, noncanonical-integer, negative-natural, bad-boolean, zero-id, negative-tuple, missing-root, missing-constraint, missing-element, duplicate-id, missing-link-17 and missing-link-18.
Logs name each catcher. The earlier filtered native/Node/JS oracle and one-byte mutant evidence remain in the [original wave report](../WAVE_24_REPORT.md); that gate was not rerun for this batch because no compiler/runtime code changed.

Compilation refusals during development were resolved in the wave-owned files: explicit undefined checks, constructor dependency injection, explicit sort comparator, removing optional calls, typed empty-array fallbacks, template interpolation, array splice and replacing a recursive initializer closure with a method. They are not counted as mutant catches.
The Go byte comparator also caught the wrong Never flag and the break/finally behavior noted above; direct flag tests now pin the actual checker values.

## Native time against Go

Three alternating complete count runs after builds and gates stopped:

```sh
python3 bridge/tsgo/profile/volume_bench.py \
  /workspace/wave-24-next-final/native /workspace/wave-24-next-final/oracle \
  /workspace/wave-24-next-final-timings \
  --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-24-next-final/repository.manifest \
  --corpus compiler /workspace/wave-24-corpus/src/compiler/tsconfig.json /workspace/wave-24-next-final/compiler.manifest \
  --rounds 3 > /tmp/wave-24-next-final-timings.log 2>&1
```

| Corpus | Native median process | Go median process | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.667169s | 0.085857s | 7.77x |
| Compiler | 4.333930s | 0.329260s | 13.16x |

Observation: native is slower on both zero-finding populations. These include load, native parse/binding indexing/walks, bridge work and process overhead; they are not isolated per-rule or per-C-call costs. No speedup is claimed. Raw samples and phase stderr are saved in evidence.

## Scope and limits

No shared-harness gap blocks the native-versus-Go result: this directory's validator loads the .a suite directly through native Adamic and supplies .a roots to the independent Go oracle.
The shared .a module/profile/emitted-JavaScript work named by Ahra was not available as `origin/codex/lint-harness-dot-a` in the fetched refs. No shared harness or generator was changed, and no emitted-JavaScript rule/profile comparison or full production CLI .a lint pass is claimed.
The controls cover a resolved Nexus declaration import, but do not establish every imported .a implementation-body/module-resolution case.
No full `go test ./...` gate or complete upstream fixture matrix ran. Arbitrary TypeScript/JSX inputs, decorators, every class initializer/control-flow shape, CLI display, suppressions and edit application are not claimed covered. These rules propose no edits.
The evidence establishes the frozen corpora, explicit controls, direct facts, six byte-only mutants and the ownership/refusal gates; it is not a proof for all source programs.

[evidence](evidence) contains canonical compressed diagnostics, controls, manifests, source hashes, selection records and logs. Binaries, C archives, scratch mutations and oracle overlays are not committed. No PR was opened.
