Built: two continuation ports; the original wave-15 three remain complete.
Commits: original completion 2467cd3a; continuation claim 3dd0684b; implementation 360353d04503648ef44811876eaa6720f822ce0b.
Checks: two-rule suite PASS 50.591s, bridge PASS 51.300s, raw facts PASS, vet and diff check empty.
Mutants: listener narrowing inversion differs at byte 2954; mock owner inversion differs at byte 48; both exit 0 with empty stderr.
Not covered: leaked-number-render needs JSX parser support; stopped following Ahra's correction.

The early continuation request arrived after all original work was pushed.
`git push origin HEAD:refs/heads/codex/typeaware-wave-15` returned Everything up-to-date.
Fetching all heads with `git fetch --recurse-submodules=no origin '+refs/heads/*:refs/remotes/origin/*'`
updated 320 origin refs. The combined 197 checker-rule ranking, excluding the
existing 26 ports, and 32 Markdown claim files mentioning 93 ranked rules selected
global-listener-target-assertion, leaked-number-render, and mock-on-module-namespace.
Their claim was committed and pushed before implementation. Inventory/count
records were not treated as implementations.

`no_global_listener_target_assertion.a` and `no_mock_on_module_namespace.a`
make the production rule decisions in Adamic. The independent Go oracle invokes
the unchanged pinned cohere production rules and imports no bridge code.
Each rule has its own file. All new Adamic sources use .a. Raw bridge questions
`call-declaration-chain`, `namespace-binding`, and `nonnullable-shape` each have
separate Go and Adamic files. The shared facts dispatcher has only their three
registration cases; no shared registration generator or existing test harness
was changed. A new independent wave test uses existing harness helpers.
The source-file and type declaration metadata decoder is local to this wave:
importing Caller pulled in an existing Unused constructor-proof refusal, so the
new port decodes the existing raw question without that dependency.

Complete canonical findings include spans, rule names, message IDs and texts,
fixes and suggestions, preserving duplicates. These two production rules have
no fixes or suggestions. Under default strict settings:

| Population | Findings | Identical bytes |
| --- | ---: | ---: |
| 16 controls | 11 | 6987 |
| 77 compiler roots | 0 | 5318 |
| Frozen 287 repository roots | 0 | 18485 |

All three populations also pass ASan, UBSan and LeakSanitizer comparisons.
The corpus manifests are the same frozen populations already documented in
WAVE_15_REPORT.md. Zero corpus findings alone are not positive rule evidence.
Controls include all four mock methods, wrappers, type-only namespaces, copies,
shadowed bindings, another package's MockTracker, inline and named handlers,
const versus let handlers, event identity in nested callbacks, DOM declarations,
general versus specific element types, unions with null, custom elements,
instanceof narrowing, Unicode and CRLF. CommonJS and NodeNext controls match Go
at 6 findings/4320 bytes; Preserve and ES2020 match at 11/6987 bytes.

Each rule's mutant compiled, ran normally and was caught solely by independent
Go output comparison. The native suite checks that a query after release exits
70 with invalid or released checker handle. Direct tests compare nonnullable
identities to the checker, declaration ancestry and namespace flags, and reject
wrong node kinds, question suffixes and invalid/noncanonical identities.
The complete bridge regression validates 100 ABI buffers surviving release,
stale/zero handles, 162 positions/3261 bytes under sanitizers, and catches
input/output length mutants with ASan, missing frees/heap region allocation
with LeakSanitizer, stale-handle retention by assertion, wrong source position
at oracle byte 6, and missing link opt-in by refusal.

Single final whole-process observations, with identical timed finding streams:

| Population | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.628084s | 0.274906s | 5.92 |
| Repository | 0.238270s | 0.112096s | 2.13 |

These are single observations, not isolated medians or a speed-parity claim.
Corpus rule trigger shapes do not occur and bridge query counts are zero;
these timings chiefly measure program loading and native parsing/traversal.

Commands, all test output redirected to files:

```sh
bash cloud/setup.sh > /workspace/wave15-cont-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE15_CONT_ARTIFACTS=/workspace/wave15-cont-final ADAMIC_WAVE15_COMPILER_MANIFEST=/workspace/wave-15-compiler.manifest ADAMIC_WAVE15_REPOSITORY_MANIFEST=/workspace/wave-15-repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./stage1/cohere/typeaware -run '^TestWave15ContinuationAgreementAndMutants$' > /workspace/wave15-cont-final.log 2>&1
go test -v -count=1 -timeout=15m ./bridge/tsgo/... > /workspace/wave15-cont-bridge.log 2>&1
go test -v -count=1 ./bridge/tsgo/checker -run '^TestWave15ContinuationRawFacts$' > /workspace/wave15-cont-rawfacts.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave15-cont-vet.log 2>&1
git diff --check > /workspace/wave15-cont-diffcheck.log 2>&1
```

Setup: go ready 0s, clang ready 0s, node ready 0s, submodules ready 0s,
build cache warm 30s, total 30s. `nproc`: 5, CPU quota 4. Go 1.27.1,
clang 20.1.8, Node 24.19.0. Production cohere stdin formatting is stable
for all seven new .a files, using virtual .ts path aliases without writing .ts
files; this is formatter verification, not an entire cohere lint gate.

The remaining rule is **nexus/correctness-no-leaked-number-render**. The parser
on this branch has no JSX descent. A native probe constructed Parser with
`control.tsx` and a valid JSX number-render expression. It compiled successfully
and exited 70:

```text
adamic: panic: parser slice expected CloseBraceToken, got AmpersandAmpersandToken at 49 in control.tsx
```

JSX support exists on the separate origin/codex/stage1-jsx-lint branch, but
importing it changes shared parser/scanner files. Ahra's correction says keep
changes inside the rule directories and stop on blockers. No shared parser,
scanner, registration generator or existing harness was edited. The leaked-number
rule has no implementation, byte agreement or rule mutant claimed here. Its
reservation remains recorded as blocked so Ahra can resolve the parser dependency.
No further rules were claimed. Full repository gates, the entire upstream fixture
matrix and .mts/.cts extension cases were not run.

Evidence logs, compressed canonical streams and matching SHA-256 hashes are in
validation-wave-15-continuation. Headers retain original absolute paths, so
hashes describe this run. Submodule pins and the frozen corpora are unchanged.
