fixtures: 18 of 40 native runs agree with source Node; 515 counts survey sites, not executable fixtures.
sites: 98 of 98 receiver audits unchanged; individual original-site compiler proofs unmeasured.
mutants: 18 real IR, 20 input, 80 header controls and 4 receiver/ledger mutants caught; 3 synthetic evidence controls rejected.
branch: codex/step10-on-writes-next; report commit is the branch head.
compiler head: 781766edccfd21ed4f2e9f9106dd488d28751e96.

# Checked writes on the rebuilt compiler

Task: #cf15j5c

The two requested commits were cherry-picked as 1d63e09d (dc8d6447) and fa9ec7e4 (3af7a3f6). No compiler code changed. Their original commit messages are preserved; this report commit carries the task trailer.

## Coverage and observations

The supplied corpus has 40 executable boundary fixtures, not 515. It selects 20 source families covering 381 of 515 unresolved census sites. There are no fixtures for the other 134 sites, and a family driver is not an individual-site execution. All 40 source Node records pass exactly. Native and emitted JavaScript compile for 36 inputs and agree with one another on every stdout, stderr and exit code. All 18 fitting inputs that compile agree with source Node; all 18 misfitting inputs stop before the write with exit 70. Those 18 source-Node disagreements are intentional checked-write behavior, recorded individually below. Four inputs do not compile. There are no observed silent miscompiles.

The strict runtime contract passes 31/40 and fails nine inputs. Five failures concern required diagnostic content, not missing runtime checks. Two are unshift refusals and two are NodeArray representation NotYet stops. Expectations were not weakened.

| Negative fixture | First differing stdout line (source Node / native and emitted JS) |
|---|---|
| 01_located-diagnostic_out.a | line 2: `stored undefined` / `<EOF>` |
| 02_shared-empty_out.a | line 2: `length 1` / `<EOF>` |
| 03_flow-node_out.a | line 2: `stored BindingElement` / `<EOF>` |
| 04_resolved-members_out.a | line 2: `stored undefined` / `<EOF>` |
| 05_generic-range_out.a | line 2: `stored 1` / `<EOF>` |
| 06_builder-tracker_out.a | line 2: `stored fallback` / `<EOF>` |
| 08_generated-identifier_out.a | line 2: `stored undefined` / `<EOF>` |
| 09_identifier-flags_out.a | line 2: `stored 1` / `<EOF>` |
| 10_expression-range_out.a | line 2: `stored 1` / `<EOF>` |
| 11_node-flags_out.a | line 2: `stored 0` / `<EOF>` |
| 12_captured-this_out.a | line 2: `stored undefined` / `<EOF>` |
| 13_declaration-array_out.a | line 2: `stored Expression` / `<EOF>` |
| 15_detached-diagnostic_out.a | line 2: `stored main.a` / `<EOF>` |
| 16_flow-assignment-union_out.a | line 2: `stored BindingElement` / `<EOF>` |
| 17_generated-node-union_out.a | line 2: `stored undefined` / `<EOF>` |
| 18_synthetic-super_out.a | line 2: `stored undefined` / `<EOF>` |
| 19_mutable-generated_out.a | line 2: `stored undefined` / `<EOF>` |
| 20_diagnostic-union_out.a | line 2: `stored undefined` / `<EOF>` |

Every negative above also differs in exit (source Node 0, native/JS 70) and stderr (source Node empty, native/JS the panic in observations.json).

| Smallest supplied failing fixture | Observation and classification |
|---|---|
| 02_shared-empty_out.a | Compiler diagnostic gap: adamic: panic: write failed: array[] expects never, got 1; required ['[]', 'never', 'number'] |
| 03_flow-node_out.a | Compiler diagnostic gap: adamic: panic: write failed: flow.node expects BinaryExpression, got object; required ['node', 'BinaryExpression', 'BindingElement'] |
| 07_diagnostic-array_in.a | Compiler gap: /tmp/adamic-gate/checked-writes-a_lvyzgv/07_diagnostic-array_in.ts:9:5: Adamic 0.1 refuses inherited library member unshift read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| 07_diagnostic-array_out.a | Compiler gap: /tmp/adamic-gate/checked-writes-a_lvyzgv/07_diagnostic-array_out.ts:9:5: Adamic 0.1 refuses inherited library member unshift read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method) |
| 13_declaration-array_out.a | Compiler diagnostic gap: adamic: panic: write failed: array[] expects Declaration, got object; required ['[]', 'Declaration', 'Node'] |
| 14_statement-array-range_in.a | Compiler gap: /tmp/adamic-gate/checked-writes-a_lvyzgv/14_statement-array-range_in.ts:15:7: stage 0 can't lower a value of type NodeArray<Statement> & { readonly pos: 0; readonly end: 0; } yet |
| 14_statement-array-range_out.a | Compiler gap: /tmp/adamic-gate/checked-writes-a_lvyzgv/14_statement-array-range_out.ts:15:7: stage 0 can't lower a value of type NodeArray<Statement> & { readonly pos: 0; readonly end: 0; } yet |
| 15_detached-diagnostic_out.a | Compiler diagnostic gap: adamic: panic: write failed: diagnostic.file expects undefined, got object; required ['file', 'undefined', 'SourceFile'] |
| 16_flow-assignment-union_out.a | Compiler diagnostic gap: adamic: panic: write failed: flow.node expects BinaryExpression, got object; required ['node', 'BinaryExpression', 'BindingElement'] |

The diagnostic-array driver calls unshift on its receiver directly. Node accepts it, so the inherited-library-member refusal looks like a compiler gap. The NodeArray intersection still lacks a representation. No changed upstream fact was observed: the source pin, hashes, full audit classifications and Node outputs remain unchanged. Historical refusal headers and tool assumptions were stale fixture infrastructure.

## Proven sites

The whole adapted source was reconstructed using stage3/apply.sh from 12fef296, and verify.cjs confirmed all source hashes and retained spans. recheck.cjs recomputes all 98 original receiver graphs with the pinned TypeScript 6.0.3 audit: 98 unchanged, zero classification changes. This is external audit evidence, not Adamic emission evidence.

Two authored witnesses compile with zero writable-view checks according to --explain-checks, and agree with source Node in ASan/UBSan/LeakSanitizer native and emitted JavaScript. They represent 86 ledger rows. These reductions do not establish separate compile/check observations for each original source site. The compiler observation producer for all 98 original sites is still absent; no PROVEN_COMPILER_RESULTS values were fabricated.

| Witness | Represented IDs | Current compiler result |
|---|---|
| 01_truthiness.a | 30, 31, 43, 50, 51, 52, 54, 57, 61, 62, 63, 64, 65, 83, 90, 102, 103, 105, 117, 132, 134, 136, 141, 143, 145, 146, 160, 171, 191, 193, 202, 203, 204, 205, 206, 215, 216, 217, 220, 221, 225, 227, 228, 234, 269, 271, 272, 275, 288, 289, 290, 291, 301, 302, 303, 304, 305, 314, 318, 320, 325, 326, 327, 329, 344, 346, 347, 348, 351, 359, 383, 384, 385, 391, 393, 395, 396, 398, 399, 400, 426, 607 | Compiles |
| 02_control.a | 55, 200, 323, 338 | Compiles |
| 03_kind_reader.a | 420, 421, 424 | NotYet: /workspace/adamic/stage3/fixtures/checked-writes/proven/03_kind_reader.a:14:45: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet |
| 04_range_reader.a | 415, 418 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/04_range_reader.a:36:33: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 05_diagnostic_related.a | 262, 267 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/05_diagnostic_related.a:25:9: Adamic 0.1 refuses a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| 06_modifiers_initialize.a | 352 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/06_modifiers_initialize.a:12:32: Adamic 0.1 refuses a value of type IndexSignatureDeclaration seen as Mutable<IndexSignatureDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 07_assert_clause.a | 446 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/07_assert_clause.a:12:14: Adamic 0.1 refuses a value of type ImportTypeAssertionContainer seen as Mutable<ImportTypeAssertionContainer>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 08_property_initializer.a | 514 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/08_property_initializer.a:13:54: Adamic 0.1 refuses a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 09_static_modifiers.a | 516 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/09_static_modifiers.a:20:10: Adamic 0.1 refuses a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 10_module_parent.a | 536 | Refused: /workspace/adamic/stage3/fixtures/checked-writes/proven/10_module_parent.a:10:22: Adamic 0.1 refuses a value of type ModuleName seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable); use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |

Relative to the original observations, truthiness and control changed from NotYet to Compiles, and kind_reader changed from Refused to NotYet. The seven remaining witnesses still refuse, now with current diagnostics. Every affected original ID is listed above; the external proof classifications themselves did not change.

## Exact adaptations

- Added 30-second subprocess limits to verify.py and observe.py, and 30/60-second limits to verify.cjs and compiler-contract-controls.cjs. Every whole command also has an outer timeout.
- observe.py now takes --compiler and --logs, uses the built compiler, runs sanitized native and emitted JavaScript, and saves --explain-checks output. This avoids repeated go run cold builds and records actual zero-check witness evidence.
- Added recheck.cjs to rerun every original receiver graph on the pinned source. It labels the result as a TypeScript audit rather than compiler proof.
- Added headers.py and headers.json. All 50 .a files were checked where they sit. Seven proven witnesses needed refusal headers; their program bodies remain unchanged. The 40 checked-write headers already matched; truthiness, control and kind_reader correctly have no header.
- Refreshed observations.json, counts.md, proven/status.json, proven/counts.md and mutant evidence. Added a 50-row status.json in the current fixture-harness format, including repository-relative diagnostics and proven/ paths.
- Extended mutant.go.txt for container-check flags, allocation contracts and the never-array marker, and made nested interface traversal mutable and nil-safe. mutants.py now tests any matching native/JS exit-70 write stop, so a diagnostic-content gap cannot hide a testable runtime check.

## Integration blocker

TestFixtureDirectoriesHaveTopLevelTests fails with `fixture directory checked-writes has no top-level test`. The current harness only scans top-level stage3/fixtures/*_test.go. Adding the required two-line TestFixturesCheckedWrites registration there is outside the explicit territory; a scope question is pending. status.json is ready. The proposed registration is:

```go
func TestFixturesCheckedWrites(t *testing.T) {
    t.Parallel()
    testFixtureDirectory(t, "checked-writes")
}
```

No parent-harness file was edited. This unit cannot be called green or pushed under the instruction to push only after its tests pass until that scope conflict is resolved.

## Toolchain and commands

GOPROXY=https://proxy.golang.org|direct was exported before setup. nproc=5; cpu.max=400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8.

Setup timing lines: Go 0.059s; Node 0.081s; clang 0.478s; markdown install step 0.933s, ready 1.130s; submodules 17.513s; shared cache 23.785s; build cache warm 380.635s; total 380.667s. Setup passed. The first cold compiler build reached its 180-second timeout and produced no binary; its bounded warm-cache retry passed. A mistaken attempt to archive src directly from an Adamic revision failed because the adapted source is reconstructed by apply.sh; the subsequent full reconstruction passed.

All commands sourced /workspace/adamic-tools/env.sh and wrote output to /tmp/cf15j5c-*.log. Commands and observed outputs:

```text
timeout 540 bash cloud/setup.sh: exit 0, timing lines above
timeout 180 go build -o /tmp/cf15j5c-adamic ./cmd/adamic: cold timeout, retry exit 0
timeout 480 bash /tmp/cf15j5c-replay/stage3/apply.sh /tmp/cf15j5c-adapted: exit 0
timeout 240 python3 .../verify.py --compiler /tmp/cf15j5c-adamic --compiler-revision 781766ed... --survey /tmp/cf15j5c-survey.json --observe: exit 0, 40 Node, 31 contracts, 9 pending, 20 input mutants, 80 header mutants
timeout 120 python3 .../verify.py [same compiler/survey] --require-checked-writes: exit 1, the nine fixtures listed above
timeout 90 python3 .../headers.py --compiler /tmp/cf15j5c-adamic: exit 0, all 50 headers
timeout 180 python3 .../proven/observe.py --compiler /tmp/cf15j5c-adamic --logs /tmp/cf15j5c-proven-final: exit 0, 2 Compiles, 1 NotYet, 7 Refused
PROVEN_SOURCE_TREE=/tmp/cf15j5c-adapted timeout 90 node .../proven/verify.cjs: exit 0, 98 sites, 10 witnesses, three mutants
timeout 90 node .../proven/recheck.cjs /tmp/cf15j5c-adapted: exit 0, 98/98 classifications unchanged
timeout 90 node .../proven/real-mutant.cjs /tmp/cf15j5c-adapted --expect-proven: expected exit 1 at original proven-site contract
timeout 90 node .../proven/compiler-contract-controls.cjs: exit 0, stand-in accepted; three bad-evidence controls rejected
timeout 90 go build -o /tmp/cf15j5c-mutant ./stage3/fixtures/checked-writes/mutant-tool: exit 0
timeout 180 python3 .../mutants.py --tool /tmp/cf15j5c-mutant: exit 0, 18 real IR mutants caught, 2 blocked families
timeout 90 go test ./stage3/fixtures -run ^TestFixtureDirectoriesHaveTopLevelTests$ -count=1 -timeout 90s: exit 1, missing registration; test 0.011s
```

NODE_PATH=/tmp/cf15j5c-api/node_modules supplies pinned TypeScript 6.0.3 for all CJS audit commands. No whole-package or full gate run was used as confirmation.

## Mutants

The exact twenty input mutants and all runtime mutants are enumerated in the refreshed JSON evidence. The twenty fitting-to-misfitting input changes are caught by exact fitting-source Node stdout; each independently matches its negative twin on Node. The eighty removed/wrong-header controls are caught by verify.py header_accepts, not claimed as cohere Gate mutants.

Proven witness mutants: --bad-write is caught by the receiver classification contract; --drop-row by ordered identity coverage; --node-output by exact source Node stdout. Real site 420 adds emitNode = undefined inside the actual isPrivateIdentifier receiver: both original union members reject it, TypeScript diagnostics remain unchanged, and --expect-proven fails at the original proven-site contract. Synthetic compiler-result controls reject an inserted check, a refused site and a dropped row; they are explicitly not implementation evidence.

All eighteen negative fixtures in the disagreement table catch their real IR check-removal mutant through the exit-70 expectation in both backends. runtime-mutants.json records each individually; diagnostic-array and NodeArray-range are the two blocked families. Initial expanded container-mutant attempts were masked by the retained never guard or failed in the tool reflection traversal; those are not counted as successful catches. The final tool must store silently with exact source Node output, exit 0, and no sanitizer or leak failure before its exit-70 comparison counts as a catch.

Not covered: the 134 smaller unresolved families/sites, original-site Adamic emission measurements for all 98 proven rows, full TypeScript computation or reachability, arbitrary heap aliases and callback escapes beyond the pinned audit, or a full repository gate.

Required lane checks passed on the committed compiler-based branch: `lane checks 16.3 s: gofmt and tools on 41 Go files, t.Parallel on 5 test packages; a-check 15 .a files; vet skipped, over 10 s`. The script compares with main and therefore also inspects inherited compiler-branch changes. Its own vet timeout is reported as a skip, not a vet pass. This checkout originally tracked only main, so explicit tracking refspecs were needed for devtools/fast-gate and cloud/merge-tree. The final run sourced the setup environment. Full log: logs/lane-checks.log.gz.
