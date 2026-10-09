Unit u022: all nine internal/ir rows at 7b18d0576930caca4e22ce2eef92fcf563af52d0.
Clean baseline: PASS, 24.634 seconds, no skips, nproc 5.
Verdicts: six sacred, two subsumed, one setup-check.
Matrix: 17 eligible mutants, one supplemental, nine probes; two eligible survivors.
Evidence: test-audit/internal-ir, review/test-audit/internal-ir/.

```json
[
  {
    "test": "TestArgumentLayouts",
    "package": "internal/ir",
    "file": "internal/ir/argument_slots_test.go",
    "seconds": 0.029,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M08",
      "M11",
      "M12"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M12: argument_slots_test.go:24: mixed fixed/rest layout: {Fixed:[1] Rest:[{Start:0 Element:1}] Count:false}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P03",
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; argument_slots_test.go:24: mixed fixed/rest layout: {Fixed:[1] Rest:[{Start:0 Element:1}] Count:false}",
    "vacuous_subcases": []
  },
  {
    "test": "TestClosureArgumentsCountTargets",
    "package": "internal/ir",
    "file": "internal/ir/arguments_length_test.go",
    "seconds": 0.062,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M11",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12: arguments_length_test.go:13: type 11: false, want true",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestArgumentLayouts"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P03",
      "P05",
      "P09"
    ],
    "subsumer_seconds": 0.029,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; arguments_length_test.go:13: type 11: false, want true",
    "vacuous_subcases": []
  },
  {
    "test": "TestPrimitiveViewMembers",
    "package": "internal/ir",
    "file": "internal/ir/view_primitives_test.go",
    "seconds": 0.011,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M13",
      "M14",
      "M15"
    ],
    "unique_kills": [
      "M13",
      "M14",
      "M15"
    ],
    "last_proven_fail": "M15: view_primitives_test.go:27: lost packed optional contract",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; view_primitives_test.go:27: lost packed optional contract (full output in M15.log)",
    "vacuous_subcases": []
  },
  {
    "test": "TestViewUnionDiscriminantOverlaps",
    "package": "internal/ir",
    "file": "internal/ir/view_unions_untagged_test.go",
    "seconds": 0.008,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M16",
      "M17",
      "M18"
    ],
    "unique_kills": [
      "M16",
      "M17",
      "M18"
    ],
    "last_proven_fail": "M18: view_unions_untagged_test.go:15: overlapping tags selected a member",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P07"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; view_unions_untagged_test.go:15: overlapping tags selected a member",
    "vacuous_subcases": []
  },
  {
    "test": "TestCallMayThrowUsesReachableTargets",
    "package": "internal/ir",
    "file": "internal/ir/call_targets_effects_test.go",
    "seconds": 0.009,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M03: call_targets_effects_test.go:34: static throw outside target set: CallMayThrow = true, want false",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P01",
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; call_targets_effects_test.go:34: static throw outside target set: CallMayThrow = true, want false",
    "vacuous_subcases": []
  },
  {
    "test": "TestDirectClosureTargetsUseEncodedIndex",
    "package": "internal/ir",
    "file": "internal/ir/call_targets_effects_test.go",
    "seconds": 0.01,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M04"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04: call_targets_effects_test.go:62: later function: ClosureMayThrow = true, want target 3's effect",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P03",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; call_targets_effects_test.go:62: later function: ClosureMayThrow = true, want target 3's effect",
    "vacuous_subcases": [
      "P04: first function and later function retain passing nonthrowing assertions"
    ]
  },
  {
    "test": "TestCallTargetReaders",
    "package": "internal/ir",
    "file": "internal/ir/call_targets_guard_test.go",
    "seconds": 0.894,
    "oracle": "Go-resolved source field reads compared with handwritten targetReaders allowlist.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S01: call_targets_guard_test.go:169: unapproved call-target read internal/native/arguments_length.go:spreadArguments:Call.Function at /workspace/adamic/internal/native/arguments_length.go:136:34; use CallTargets or ClosureTargets",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "S01: timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run '^TestCallTargetReaders$'; call_targets_guard_test.go:169: unapproved call-target read internal/native/arguments_length.go:spreadArguments:Call.Function at /workspace/adamic/internal/native/arguments_length.go:136:34; use CallTargets or ClosureTargets",
    "vacuous_subcases": []
  },
  {
    "test": "TestCallTargetsIncludeEveryDescendant",
    "package": "internal/ir",
    "file": "internal/ir/call_targets_test.go",
    "seconds": 0.045,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: call_targets_test.go:40: want four implementations, got [3]",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCallMayThrowUsesReachableTargets"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P01",
      "P08"
    ],
    "subsumer_seconds": 0.009,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; call_targets_test.go:40: want four implementations, got [3]",
    "vacuous_subcases": []
  },
  {
    "test": "TestClosureTargetsBoundOnlyProvenValues",
    "package": "internal/ir",
    "file": "internal/ir/call_targets_test.go",
    "seconds": 0.012,
    "oracle": "Handwritten IR expectations; no external value checked.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M06"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M06: call_targets_test.go:78: ir.Conditional: Unknown failed to include throwing function",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "P03",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestArgumentLayouts",
      "TestClosureArgumentsCountTargets",
      "TestPrimitiveViewMembers",
      "TestViewUnionDiscriminantOverlaps",
      "TestCallMayThrowUsesReachableTargets",
      "TestDirectClosureTargetsUseEncodedIndex",
      "TestCallTargetReaders",
      "TestCallTargetsIncludeEveryDescendant",
      "TestClosureTargetsBoundOnlyProvenValues"
    ],
    "evidence": "ADAMIC_MUTANT=M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; call_targets_test.go:78: ir.Conditional: Unknown failed to include throwing function",
    "vacuous_subcases": [
      "P04: all nine bounded callback/comparator/constant cases pass; unknown throwing cases fail"
    ]
  }
]
```

| ID | origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/ir/call_targets.go:10 | `if call.Virtual != 0 {` to `if call.Virtual == 0 {` | TestCallMayThrowUsesReachableTargets, TestCallTargetsIncludeEveryDescendant |
| M02 | internal/ir/call_targets.go:15 | `return targets` to `return targets[:1]` | TestCallMayThrowUsesReachableTargets, TestCallTargetsIncludeEveryDescendant |
| M03 | internal/ir/call_targets.go:22 | `if p.Functions[target].MayThrow {` to `if !p.Functions[target].MayThrow {` | TestCallMayThrowUsesReachableTargets |
| M04 | internal/ir/call_targets.go:46 | `[]int{call.Direct - 1}` to `[]int{call.Direct}` | TestDirectClosureTargetsUseEncodedIndex |
| M05 | internal/ir/call_targets.go:72 | `[]int{target - 1}` to `[]int{target}` | TestClosureTargetsBoundOnlyProvenValues |
| M06 | internal/ir/call_targets.go:75 | `return FunctionTargets{Unknown: true}` to `return FunctionTargets{Unknown: false}` | TestArgumentLayouts, TestClosureArgumentsCountTargets, TestClosureTargetsBoundOnlyProvenValues |
| M07 | internal/ir/call_targets.go:84 | `if function.MayThrow {` to `if !function.MayThrow {` | survivor |
| M08 | internal/ir/argument_slots.go:11 | `start := len(function.Parameters) - 1` to `start := len(function.Parameters)` | TestArgumentLayouts |
| M09 | internal/ir/argument_slots.go:30 | `fixed = max(fixed, n)` to `fixed = min(fixed, n)` | survivor |
| M10 | internal/ir/argument_slots.go:100 | `} else if of.IsMaybe() {` to `} else if !of.IsMaybe() {` | survivor |
| M11 | internal/ir/argument_slots.go:69 | `layout.Count = layout.Count || p.PackedCountNeeded(target)` to `layout.Count = layout.Count && p.PackedCountNeeded(target)` | TestArgumentLayouts, TestClosureArgumentsCountTargets |
| M12 | internal/ir/argument_slots.go:150 | `if f.ReadsArguments || f.RestElement != 0 && (f.Closure || f.Receiver) {` to `if f.ReadsArguments && f.RestElement != 0 && (f.Closure || f.Receiver) {` | TestArgumentLayouts, TestClosureArgumentsCountTargets |
| M13 | internal/ir/view_primitives.go:15 | `if contract.Unsupported != "" {` to `if contract.Unsupported == "" {` | TestPrimitiveViewMembers |
| M14 | internal/ir/view_primitives.go:50 | `if contract.Undefined {` to `if !contract.Undefined {` | TestPrimitiveViewMembers |
| M15 | internal/ir/view_primitives.go:36 | `contract.Of = Number` to `contract.Of = Boolean` | TestPrimitiveViewMembers |
| M16 | internal/ir/view_unions_untagged.go:33 | `|| seen[literal]` to `|| !seen[literal]` | TestViewUnionDiscriminantOverlaps |
| M17 | internal/ir/view_unions_untagged.go:24 | `own.Name != field.Name || own.Optional ||` to `own.Name != field.Name || !own.Optional ||` | TestViewUnionDiscriminantOverlaps |
| M18 | internal/ir/view_unions_untagged.go:48 | `if valid {` to `if !valid {` | TestViewUnionDiscriminantOverlaps |
| P01 | internal/ir/call_targets.go:9 | `func (p *Program) CallTargets(call Call) []int {` to `return nil` | TestCallMayThrowUsesReachableTargets, TestCallTargetsIncludeEveryDescendant |
| P02 | internal/ir/call_targets.go:20 | `func (p *Program) CallMayThrow(call Call) bool {` to `return false` | TestCallMayThrowUsesReachableTargets |
| P03 | internal/ir/call_targets.go:41 | `func (p *Program) ClosureTargets(call Expression) FunctionTargets {` to `return FunctionTargets{}` | TestArgumentLayouts, TestClosureArgumentsCountTargets, TestDirectClosureTargetsUseEncodedIndex, TestClosureTargetsBoundOnlyProvenValues |
| P04 | internal/ir/call_targets.go:80 | `func (p *Program) ClosureMayThrow(call Expression) bool {` to `return false` | TestDirectClosureTargetsUseEncodedIndex, TestClosureTargetsBoundOnlyProvenValues |
| P05 | internal/ir/argument_slots.go:52 | `func (p *Program) ClosureArgumentLayout(call CallClosure) ArgumentLayout {` to `return ArgumentLayout{}` | TestArgumentLayouts, TestClosureArgumentsCountTargets |
| P06 | internal/ir/view_primitives.go:6 | `func PrimitiveViewMembers(program *Program, id ViewContractID) ([]ViewContract, bool) {` to `return nil, false` | TestPrimitiveViewMembers |
| P07 | internal/ir/view_unions_untagged.go:6 | `func ViewUnionHasDiscriminant(contracts []ViewContract, root ViewContract) bool {` to `return false` | TestViewUnionDiscriminantOverlaps |
| P08 | internal/lower/lower.go:20 | `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {` to `return &ir.Program{}, nil` | TestCallTargetsIncludeEveryDescendant |
| P09 | internal/ir/call_targets.go:101 | `func (p *Program) ClosureReadsArgumentsCount(call CallClosure) bool {` to `return false` | TestClosureArgumentsCountTargets |

S01 is a setup construction probe: call_targets_guard_test.go:27 renames the allowlisted spreadArguments reader to missingReader. TestCallTargetReaders fails at line 169 on the now-unapproved actual reader, and at line 185 on the stale entry. It is excluded from production kills.

Survivors:
- M07, unguarded: witness-clean.log prints unknown all-throwing effect true; witness-M07.log prints false.
- M10, unguarded: witness-clean.log prints mixed optional fixed layout [7] (MaybeNumber); witness-M10.log prints [1] (Number).
- M09, supplemental survivor: witness-clean.log prints two-parameter fixed slots 2; witness-M09.log prints 0. Excluded from all verdicts because max-to-min is outside the menu.
Witness command: go build -o /tmp/u022-witness review/test-audit/internal-ir/survivor-witness.go; ADAMIC_MUTANT=<id> /tmp/u022-witness. The binary was built while the scratch switch was installed. It has no expected-answer assertion and is not a test mutant.

Limits and brief issues:
- No Your rows section was supplied. The scope therefore uses every Test from go test -list. No claimed moved or vanished names can be checked without an assigned list.
- Warm env.sh worked but cohere and its TypeScript submodule were absent. SSH failed because port 22 was unavailable. CLAUDE.md's HTTPS URL fixed initialization. npm ci succeeded in 0.725 seconds.
- The first discovery compilation hit the 120-second backstop with an empty log. The next clean run finished compilation and passed. This was compilation overhead, not a test timeout. No mutant test binary exceeded 90 seconds.
- The initial function inventory omitted NumberConstant.Type, Concat.Type, MakeClosure.Type, CallClosure.Type, Read.Type and Call.Type. reached-functions.txt records the later complete observed IR coverage inventory. This is a procedure limitation, not concealed pre-mutation coverage evidence. Lowering's whole implementation was not enumerated or mutated. Its empty-entry probe P08 was checked.
- M09 was recognized as outside the fixed menu and marked supplemental. An initial switch-generation scope error was repaired before any matrix results; driver-execution.log and switch-vet.log retain that history.
- The initial warm-up command attempted discovery from stage3/api by mistake. It failed before running tests. Discovery and all baseline/matrix commands subsequently ran from repository root.
- Go JSON can split long output lines into chunks. rows.json quotes a failing prefix for the primitive-view assertion; the full output is in M15.log.
- M01 and M05 panic. Every row was rerun alone for each, using the same 90-second budget; only these isolated observations establish their matrix cells. No unknown cells remain.
- Subsumption is a hint based on three caught mutants for ClosureArgumentsCountTargets, and two for CallTargetsIncludeEveryDescendant. The latter also checks lowering facts which these IR mutants do not distinguish. Neither result justifies deletion.
- All expected values are self oracles. The descendant row checks target membership and a count, not executed dispatch behavior. No outside authority was checked. Repo-wide uniqueness remains for central replay.
- Empty-answer probes do not establish any sacred or subsumed verdict. The nine probes all ran against the complete package. No production row is vacuous under its own entry probe. The setup-check has vacuous null. Some nonthrowing positive subcases survive P04, listed in rows.json.
- No other package tests or repository-wide gate ran. TestCallTargetReaders itself invokes go list on native/lower/fresh/flow as its normal oracle construction; this is package listing, not running other package tests.
- Production source was restored exactly. git diff --check passes. The restored package passed again with coverage, 0.676 seconds and 49.2 percent statement coverage. All 18 standalone Go mutant diffs passed go vet ./internal/ir/ against the starting source. Probe diffs and the combined switch are separate evidence.

Timing:
- Toolchain setup was skipped because env.sh worked; nproc 5. Environment startup and HTTPS submodule initialization are separate overhead. Exact setup subprocess durations were not captured.
- The failed first cold discovery used the full 120-second compile backstop. The initial baseline binary took 24.634 seconds, including its guard's cold go-list export work.
- The 27 isolated timing commands used 67.333 wall seconds including Go command overhead. Each row's reported median uses its binary ok line only.
- Mutation/probe commands, including isolated panic recovery and the extra entry probes, used 147.504 wall seconds including compilation. The successful switch gofmt/vet build used 6.334 seconds. Per-command durations are in run-meta.json. Separate standalone-vet durations were not captured; their passing logs are retained. No native product was built.
