Built: rule.json name/kinds declarations for all thirteen owned rule modules, validated against production Go listener maps.
Commits: based on current main e8ba3d5d; prior numeric declaration commit 697b26e1 is pushed; this metadata continuation is committed separately.
Checks: TestWave21ListenerDeclarations PASS 28.428s, normal and ASAN; vet PASS. Current main and owned remote landing state were verified.
Mutants: native enum listener 267 changed to SourceFile 307 compiles, exits 0 with empty stderr, and independent Go bytes catch byte 15.
Not covered: native React source lowering, numeric parser/handed-node integration, shared finding-model integration, a new speed measurement or the full gate. No new reservations.

The declarations are under wave21_rules/<owned-module>/rule.json, with canonical rule names and the production Go kind names. This matches the JSON metadata schema used by the shared harness branch; the JSON names are metadata, not string comparisons inside native rule callbacks. The existing native syntaxKinds exports retain the pinned numeric compiler values. The private test parses each unmodified production Go rule.Listeners composite literal, resolves keys using exported Go ast.Kind constants, validates the JSON kinds, then compares a compiled native program importing all thirteen numeric declarations. ASAN repeats the complete declaration comparison. The native wrong-kind mutant remains a byte-only failure. Shared parser, generator, harness and finding-model files were not edited.

These are declaration-only metadata directories. They do not contain registered source handlers or a create/visit wrapper. The three React validator cores still take prepared HirFunction inputs, not native source ParseNodes; treating their run methods as registered source visitors would be incorrect. These manifests therefore are not put into the active shared lint/rules registry. Actual handed-node integration waits for the shared node/driver contract and a native source-to-HIR producer. Existing source ports still use legacy parser kinds and refetches; this change does not claim compliance with the new dispatch performance bar.

Observed current dependencies:

- Main remains e8ba3d5d81de4d3773c723914fccd4c76248b965 and is an ancestor of the owned branch. No rebase was needed this turn. This unit has pushed only codex/typeaware-wave-21; no main or area branch was updated.
- The shared harness branch is f4d98cab50048692781da3599131317dc569d466. Its commits add .a module loading, emitted-JavaScript comparison, shared profiling compilation and suggestion certification. Those former harness limitations have been addressed on that branch; they are not the remaining React HIR blocker.
- The local shared ParseNode still exposes readonly kind: string, with no numeric field. No native React source-to-HIR/SSA producer or graph-transform adapter is present in the owned pipeline. JSX parser refusals and the separately observed Go two-creator phi nondeterminism remain. The validators' prepared-HIR agreement is recorded in WAVE_21_REACT_CORE_REPORT.md and WAVE_21_LANDING_REPORT.md, and is not full native source agreement.
- No shared batch-8 Diagnostic landing SHA was supplied in this message. No feature-branch merge or shared-file modification was attempted. Rebase onto its main landing when it is named, as instructed.

An availability preflight fetched 493 origin refs and deduplicated 33 Markdown claim documents. Their exact contents are archived in validation-wave-21-rule-json/origin-claims.json.xz. No claim was written from this preflight. The ranking contains entries with no claim mention, including React JSX rules; the ranking is not declared exhausted. A final released-reservation-aware claim selection remains deferred while the existing React source ports are partial. The original released promises/spread/lost-update claims have continuation reservations, as the earlier selection reports document. This audit does not treat historical released-handle test wording as release of a rule reservation.

Command:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE21_LISTENER_ARTIFACTS=/workspace/wave21-rule-json go test ./stage1/cohere/typeaware -run '^TestWave21ListenerDeclarations$' -count=1 -timeout=10m -v > /workspace/wave21-rule-json.log 2>&1
go vet ./stage1/cohere/typeaware > /workspace/wave21-rule-json-vet.log 2>&1
```

No rule decision code changed this turn. The six wave-21 rule suites and inherited bridge oracles previously passed on this same unchanged main; their 880.182s and 579.690s runs, rule mutants, both frozen corpora, released handles and sanitizer coverage remain archived in validation-wave-21-landing. Setup on this same base was 119s and nproc 5. Neither those finding suites nor a full repository gate or timing benchmark was repeated for JSON metadata and its contract check. New Adamic code remains .a.
