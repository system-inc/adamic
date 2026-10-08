Built the v9z58pb class/interface literal ruling with structural storage, checked downcasts and actual receiver dispatch.
Code pushed: Adamic `490aadcd`; cohere `3c9f205b`; non-null merge `e9f58564`; current-main merge `bc1c91a0`.
Checks: four Node oracle fixtures, native release, ASan/UBSan and leak checks pass; focused package and non-null integration checks pass; counts refreshed.
Mutants: all eight fail for the intended reason; static value dispatch produces `A literal woof` instead of `A literal literal`.
Uncovered: the original 46 field sites remain in the other worker's `property` function; the one base site does not reproduce. Generic structural-to-class downcasts remain refused.

## Ruled behavior

The requested source is `internal/oracle/testdata/class_interface_literal.a`. Source Node, generated JavaScript and sanitized/release native execution agree byte for byte:

```
A literal literal
A a woof
```

A fresh object satisfying the interface member of `C | I` is an ordinary `ir.ObjectLiteral`, with no class identity. Required interface properties are proved individually so TypeScript's excess-property checks do not misclassify the fresh literal as a class. Existing nominal refusals still apply to plain class targets, unrelated classes, all-class unions, nested class containers and invalid overrides.

Property checks of fixed public member names query own fields, class methods and inherited accessors without invoking a getter. A shared method retains structural dispatch. Reads whose contextual target actually demands a nongeneric class, and explicit casts to that class, use the existing checked class-cast IR and panic on an absent tag. Inferred variable aliases and forwarding results retain their union origin for method dispatch. The additional presence fixture tests those cases, effect-only calls and inherited getters. Two checked fixtures accept a real `A`, print `woof`, then stop with exit 70 on the literal.

Conservative scope: only unions consisting of declared classes and named interfaces are admitted by these helpers. `in` requires a fixed public string-literal member name. An erased class tag cannot prove generic arguments hidden through an interface, so those downcasts remain refused. No original design refusal besides the ruled literal case was relaxed.

The pinned cohere checkout predates the named case “a literal a class and an interface both take”. It was added to clean acceptance coverage; every existing nominal-class refusal case is unchanged. No cohere implementation was copied. Its test-only commit `3c9f205bfa6101c4b06707c36d5375f2372e26c1` was pushed to `codex/notyet-class-construction` and the parent submodule pointer updated.

## Ownership and runtime review

The production hooks are in `classViewRefusal`, the shared `expression` entry, `refuse`, `expressionStatement` and `castProof`. The `490aadcd` commit message mistakenly calls the last function `proveCast`; the actual edited function is `castProof`.

Worker-owned `callOrMethod`, `enumNeverValue`, `callClosure` and `object.go::property` remain untouched. The new `internal/lower/class_interface.go` helper isolates this rule. Non-merge history on all fetched `origin/codex/notyet-*` branches showed the statements file's only post-base edit was `returnStatement`, so the effect-only call hook uses `expressionStatement`.

Automatic review initially rejected an erased-method helper edit. It was left untouched, and review approved the separate expression path after removing worker-owned function edits. A second rejection treated the entire statements file as owned; exact history proved only `returnStatement` was changed, and review approved the `expressionStatement` hook on retry. No approval remains outstanding for this delivery.

Runtime-owner review: `internal/native/runtime/class_interface_property.c` is a new, separate helper only. It searches existing fields, method maps and class accessor metadata; no existing runtime C file changed. The IR adds a flag to the existing two-operand presence node, so flow walkers retain the same operands. JavaScript queries its accessor metadata because generated class accessors are not ordinary prototype properties.

## Commands and observed outputs

All test stdout/stderr went directly to log files. No whole-package test sweep or full gate was run.

| Command | Observed result |
|---|---|
| `go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/class_interface_' -count=1 -v -timeout 10m` | 4 fixtures pass after both merges, 2.649s |
| `go test ./internal/lower -run '^(TestClassInterfaceRuling\|TestInheritance\|TestUncheckableCastsStayRefused\|TestCheckedCastProofAndElision\|TestNonNullAssertion\|TestAdamicNullishAssertionsAreRefused)' -count=1 -timeout 10m` | pass, 2.398s |
| `go test ./internal/ir -run '^Test(CallTargetsIncludeEveryDescendant\|ClosureTargetsBoundOnlyProvenValues\|CallTargetReaders)$' -count=1 -timeout 10m` | pass, 39.377s |
| `go test ./internal/native -run '^Test(PassThroughsAreNotConsumers\|RuntimeFieldLayoutsAreIncluded\|UniformFieldsMatchNode)$' -count=1 -timeout 10m` | pass, 1.364s |
| `go test ./internal/javascript -run '^TestClassInterface' -count=1` | compiles; package has no tests; backend held to Node by oracle |
| In cohere: `go test ./internal/lint/rules/adamic -run '^TestNominalClass' -count=1 -timeout 10m` | pass, 0.086s |
| `go test ./internal/oracle -run '^Test(CheckedNonNullTypeScript\|CheckedNonNullAdamicRefusal\|ImpossibleNonNullFixturesAreRefused\|PossibleNonNullAdamicAssertionsAreRefused)$' -count=1 -timeout 10m` | pass, 1.730s |
| `go test ./cmd/adamic -run '^Test(NonNullExplainChecks\|ExplainChecksDriver)$' -count=1 -timeout 10m` | pass, 3.865s |
| `go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 20m -args -update-counts` | see `ruling-evidence/counts.log`; refresh after both merges |
| `python cloud/notyet-class-interface-mutants.py` | all 8 killed; source restored by `finally` |

Toolchain setup and timing lines, `nproc=5` (CPU quota 4), and the source reconstruction/hash proof remain in [the original report](REPORT.md). No setup rerun was needed.

## Mutants

| Mutant | What catches it |
|---|---|
| Erase inferred alias origin | Presence fixture stdout differs; an inferred `A` dispatches the literal as a class |
| Dispatch value calls by static class | Exact requested fixture prints `A literal woof` natively; JavaScript loses the absent class table |
| Dispatch effect-only calls by static class | `speaker` is printed for the plain structural value instead of `plain` |
| Refuse the structural literal | Requested fixture fails lowering with the old nominal-class refusal |
| Remove contextual checked downcast | Native exit 0 instead of required exit 70 |
| Remove explicit checked downcast | Native exit 0 instead of required exit 70 |
| Query only own fields | Presence fixture stdout differs for class prototype methods and accessors |
| Invoke getter while checking presence | Unexpected `getter` stdout in native execution |

These mutants ran against the completed ruling implementation before the non-null and current-main merges. No ruling helper changed in those merges; the full focused fixtures passed again afterward. Detailed failure logs are in `ruling-evidence/`.

## Census replay after c41c0e06

`c41c0e062e99da37820f822968d4df1b48cdaee7` was fetched and merged as requested. Replay overlays were regenerated from the merged production sources, rather than reusing the old overlay that would retain the former `NonNullExpression` stop. Current `origin/main` `749a69adbfae2a7bf22c1f0436d9ef73345c3070` was merged too.

The worker was built using `make_overlay.py` and `go build -buildvcs=false -overlay=/tmp/class-construction-non-null-trace-overlay.json -o /tmp/class-construction-after-replay ./stage3/census/latent/replay/worker`. Each replay used `LATENT_FULL=1`, the unchanged verified census project `/tmp/class-construction-census-source/src/tsc/tsc.ts`, its exact `-where`, `-kind NotYet` and original `-reason`. There were 21 supplied examples: 20 exact signatures reproduce; the base-construction example remains non-reproducing. These observations are measured on a checker-rejected entry-root program, not evidence that TypeScript's compiler can now compile.

The first string/number example, `checker.ts:47945:13`, now passes its old non-null stop and reproduces the listed field reason. Every field diagnostic's runtime caller is still `internal/lower/object.go::property` at line 484. `origin/codex/notyet-object-property` has active edits to that function (notably `5a870515`); the revised territory rule still forbids editing another worker's function. None of the 46 field roots is claimed as lowered by this unit.

| SHA | Kind | Covered table roots | Status |
|---|---|---:|---|
| `bc1c91a0` | `string \| NodeArray<JSDocComment> \| undefined` | 23 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `"boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>` | 4 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[]` | 4 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `string \| number \| undefined` | 3 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `AnyBuildOrder \| undefined` | 2 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `boolean \| (() => boolean) \| undefined` | 2 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `false \| string[] \| undefined` | 2 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `"boolean" \| "list" \| "number" \| "object" \| "string" \| Map<string, string \| number>` | 1 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `"boolean" \| "number" \| "object" \| "string" \| Map<string, string \| number>` | 1 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `0 \| boolean \| undefined` | 1 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `boolean \| (() => boolean)` | 1 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `string \| false` | 1 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `string \| false \| undefined` | 1 | Skipped: actual caller is the other worker's property function |
| `bc1c91a0` | `a base that is not a declared class` | 1 | Skipped: signature does not reproduce; EmitNode union at 423:34, then reading autoGenerate at 424:18 |

The added ruled example is lowered at `490aadcd` and remains green at `bc1c91a0`. It is additional to the 47 original table roots, so no census root-site reduction is claimed. The combined replay findings and exact caller traces are preserved in `ruling-evidence/after-non-null-replays.json`.
