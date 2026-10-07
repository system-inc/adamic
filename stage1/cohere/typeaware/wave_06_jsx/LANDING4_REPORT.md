Built: rebased all wave 06 work onto main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. No new rules or shared implementation edits.
Commit: rebased implementation parent d05d4ccafff621790b0e68f22a9cbcea7bb2bb40; this evidence commit is pushed only to codex/typeaware-wave-06.
Checks: 165 post-rebase executions pass; fresh JSX Go/native/ASan/Node/JS comparison matches 61 controls, 47 findings and 22276 bytes; bridge, uncached Node and vet pass.
Mutants: all retained rule/fact/core/report/refusal mutants remain caught, including the JSX regex U+0085 mutant; fresh bridge and one-byte Node mutations also pass their failure checks.
Not covered: numeric JSX parser/driver/raw-fact adapter and native React source analysis still block source parity. Full root gate not run; no new claims.

Main advanced from f8013f0b through the Stage 3 landing. Rebase was clean. The only changed file under the compiler/oracle implementation is internal/oracle/stage3_hook_test.go; compiler, native runtime, lowering, bridge, shared parser, rules and config inputs are unchanged. No developer-tool leak helper change appears in this main snapshot. Nothing is reverted.

The earlier twelve complete rule suites and React prepared-HIR validators are re-green through a replay of their retained built artifacts. This is intentional build reuse, not a fresh rebuild claim. The unchanged compiled input paths were checked with git diff before reuse. Every replay reruns actual processes and compares stdout and exit against retained canonical results; timing-only stderr is excluded from byte equality. The Go production oracles are rerun independently. Prepared graph generation/build commands are not rerun. Normal controls, both frozen corpora, ASan binaries, Go/type-fact mutants and stale-handle guards execute again. All 165 executions pass. Original rule mutants are additionally checked against the freshly rerun Go control bytes. Existing complete-input records retain findings, fixes and suggestions.

The JSX suite is rebuilt fresh with its current RegExp literal and run against the unchanged Go registry rules, native normally and under ASan/UBSan, source Node and emitted JavaScript. Its 61 source controls produce 47 findings and 22276 equal bytes, including path headers. The changed byte total comes from the scratch-path length. All three rule mutants and the missing-U+0085 regex mutant build and exit zero, with only Go byte comparison catching them. Three source-refusal probes still exit 70. This remains prepared-node parity, not a native JSX source frontend.

The bridge is freshly rebuilt/tested, exercising C ABI ownership and released/zero handles, type bytes, sanitizer/length/link/leak/region mutants. The filtered Node oracle runs uncached and passes functions, generic functions, closures, method closures, regexp cycle closures, maps/text, sorting, string indexing and lone surrogates, plus the one-byte mutant. go vet ./... passes with empty output. All test output goes to files.

The new origin/codex/lint-regex table and README were inspected. The pinned table records 107 regexp compile sites, including 82 fixed sites, rather than the planned 92 count. It has no row for the inline-adjacency Unicode whitespace helper, which uses Go unicode.IsSpace rather than regexp. The existing single RegExp literal preserves that exact class, including U+0085 and excluding BOM. No regex options pattern is involved. No replacement or unrelated shared table edit is made.

Source JSX/numeric parser and raw declaration wiring remain absent; area/stage1-lint and lint-harness-dot-a still expose string ParseNode kinds. The three JSX source analyses explicitly refuse and remain reserved partial ports. Earlier React HIR claims stay parked on native source lowering, SSA/captures/memoization and compilation-unit selection, per #dnv6f2c. Their portable validators and post-dominance already exist. No further claims are taken while the current syntax integration remains blocked.

Reproduction sources /workspace/adamic-tools/env.sh and sets TMPDIR=/workspace:

```sh
python3 /workspace/wave-06-landing4-replay.py > /tmp/wave-06-landing4-replay.log 2>&1
python3 stage1/cohere/typeaware/wave_06_jsx/validate.py --scratch /workspace/wave-06-landing4-jsx > /tmp/wave-06-landing4-jsx.log 2>&1
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /tmp/wave-06-landing4-bridge.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /tmp/wave-06-landing4-node.log 2>&1
go vet ./... > /tmp/wave-06-landing4-vet.log 2>&1
```

The replay script, every command/exit and observation, fresh fixtures and outputs are retained in landing4_evidence with uncompressed SHA-256 hashes. Earlier source and mutant implementations remain in existing evidence directories. Toolchain configuration is unchanged; prior setup took 105s with nproc 5. No source throughput speedup or full root-gate pass is claimed. The worker branch alone is replaced using an exact lease on prior pushed 0687e7cccc89fe7f2f3dec451b6308c8dab6892c.
