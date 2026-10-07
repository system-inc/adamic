Built: numeric listenerKinds declarations in all 12 owned .a rule modules; no new claims.
Commits: starts from pushed 6dba35783a403307767b2e1d64df0ab60c4da63c, based on main e8ba3d5d81de4d3773c723914fccd4c76248b965; final commit reported after verification.
Checks: independent pinned-Go SyntaxKind validation passed for all 12 modules; all eight owned native suites passed across the initial run and retry; results below.
Mutants: wrong listener ID 218 -> 213 caught by independent enum check; existing native rule mutants rerun below.
Not covered: active kind-indexed dispatch, elimination of string-kind reads and node refetches, and three React ports remain blocked by shared dependencies.

Each module exports listenerKinds: readonly number[] for the forthcoming driver. Values come from the pinned cohere/TypeScript/tsc/internal/ast/kind_generated.go enum rather than stock npm TypeScript, whose enum may differ. The declaration contains numeric values; names appear only in comments. No shared parser, registration generator, test harness or protected compiler file changed.

Declarations:

- no-useless-default-assignment: Parameter 169, BindingElement 208.
- prefer-find: ElementAccessExpression 212, CallExpression 213.
- require-array-sort-compare, no-global-listener-target-assertion, no-mock-on-module-namespace: CallExpression 213.
- no-leaked-number-render: JsxExpression 294.
- no-label-var: LabeledStatement 256.
- no-invalid-regexp: CallExpression 213, NewExpression 214.
- no-throw-literal: ThrowStatement 257.
- prefer-arrow-callback: FunctionExpression 218.
- no-misleading-character-class and no-useless-backreference: RegularExpressionLiteral 13 for their existing partial literal adapters. Missing constructor/reference-tracker support is not advertised as implemented.

The shared native parser currently declares readonly kind: string in stage1/typescript/parser/nodes.ts, with no numeric SyntaxKind field. Parser.make accepts a string; Rules does not supply a shared numeric dispatch API. Consequently the current run methods still perform their original traversal, string-kind reads and node lookups. These declarations prepare the requested dispatch contract but do not implement the full speed rule or claim a speedup. Replacing the string field locally or building a per-rule name-to-number mapping would either violate the shared-file boundary or retain the conversion overhead the user is removing. Further runtime conversion waits for the shared numeric parser/driver. The arrow callback judgment additionally needs shared lexical-frame information currently built by its walk.

The independent contract check enumerated the pinned Go enum and compared each declared numeric array with its commented syntax-kind names. An in-memory copy changed arrow FunctionExpression 218 to CallExpression 213; the independent enum check rejected it. No production mutant was left behind. Today's driver ignores the declaration, so output comparison alone cannot catch this metadata mutant; no claim of an active dispatch oracle is made. Existing finding/fix/suggestion mutants remain independently checked against Go.

Fresh fetch: 479 origin refs. Main remains e8ba3d5d; the branch was already rebased and green against that main before these declaration edits. All-origin authored .a/.ts search for ForFunctionWithoutManualMemoization, ControlDominators, UnconditionalBlocks and AsCompilationUnit returns exit 1 with no matches. Existing React claims set-state-in-effect, set-state-in-render and static-components remain blocked on native HIR/captures, memoization, dominance and JSX/taint analyses as detailed in WAVE_15_FIFTH_REPORT.md. No further rules are claimed. This stops at Ahra's non-harness shared-dependency boundary.

Toolchain environment was reused from the immediately preceding successful setup (83s, nproc 5, cgroup quota four cores); source /workspace/adamic-tools/env.sh in this run. The previous landing report retains the full setup output. Tests write to log files, never through a pipe.

Oracle command: go test -v -count=1 -timeout=20m ./stage1/cohere/typeaware -run '^TestWave15'. Suite artifact roots are /workspace/wave15-speed/{original,continuation,leaked,core,regex,throw,arrow,backreference}. Compiler and repository manifests are the same pinned 77-file TypeScript and 287-file repository populations used for landing; source roots and artifact variables follow WAVE_15_LANDING_REPORT.md.

Results: the initial eight-suite command exited 1 after 323.105s because the disk filled. Six suites passed: arrow 62.39s, backreference 36.24s, continuation 60.57s, label 45.24s, throw 44.06s and leaked-render 44.74s. Regex mutants could not create their source directories; the original suite could not create its artifact directory. These failures were not counted as passing checks. Cleanup removed 149 disposable ELF/archive files belonging to this unit's completed landing runs, freeing 6094762975 bytes while preserving logs and canonical output streams. Exact paths are in cleanup.log.

The affected suites were rerun with fresh artifact roots /workspace/wave15-speed-retry and -run '^TestWave15RegexPartialControlsAndMutants$|^TestWave15AgreementAndMutants$'. Retry PASS 111.828s: regex 40.41s, original three rules 71.41s. All eight suites therefore passed on the same source tree and current-main base. All twelve existing mutants (arrow, backreference, listener, mock, label, throw, leaked, flags, suggestion, default, find, sort) exited zero after compilation and were caught by independent Go finding bytes. Ordinary/ASan controls and each suite's declared corpus scope match Go; released-checker-handle tests panic 70. The actual metadata contract assertion was also invoked on the 218-to-213 listener mutant and raised 'listener mismatch: expected [218], got [213]'.

Numeric metadata is ignored by today's driver; it does not change serialized output. Every timed native/Go run and differing mutant byte is retained in mutants-and-timings.txt; canonical streams and SHA256 values are retained here too. Examples: arrow repository native 286.311448ms versus Go 129.904696ms, compiler 2.246063107s versus Go 365.745291ms; label compiler 1.50214046s versus Go 280.984814ms; throw compiler 1.341854046s versus Go 284.83595ms. These runs remain slower than Go; no speed improvement is claimed.

No bridge code or shared harness changed, so bridge and filtered Node checks from the landing report were not repeated. The inherited ten-rule coverage suite and full repository gate were not repeated for these owned metadata-only changes. Partial regex, JSX and label scopes remain explicitly limited by the earlier reports. Before publishing, remote main was checked to remain e8ba3d5d and only codex/typeaware-wave-15 is pushed.
