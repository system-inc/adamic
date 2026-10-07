Built: numeric listenerKinds exports in all nine owned rule modules, ready for shared driver registration.
Commits: based on pushed landing commit 242946b4 and unchanged current main e8ba3d5d; this report accompanies the listener commit.
Checks: nine listener declarations match Go numeric AST kinds under sanitizers; all six production ports and supported React subsets re-green.
Mutants: wrong CallExpression subscription compiles and is caught only by numeric-byte comparison; existing rule and kernel mutants caught again.
Uncovered: shared parser still exposes string kinds and the driver still invokes whole-file run methods; full node-callback migration and React parity remain blocked.

Each module exports `listenerKinds: readonly number[]`. The numbers are pinned
to the TypeScript SyntaxKind enum held by production Go cohere and verified
independently through `ast.Kind` from its shim. No shared parser, harness,
registration, generator or dispatcher source was edited. New native test source
is .a.

| Rule | Numeric subscriptions |
| --- | --- |
| no-loop-func | FunctionDeclaration 263, FunctionExpression 219, ArrowFunction 220, MethodDeclaration 175, GetAccessor 178, SetAccessor 179, Constructor 177 |
| no-require-imports | CallExpression 214, ExternalModuleReference 284 |
| no-import-cycle-load-time-read | SourceFile 307 |
| no-process-exit-after-output | SourceFile 307 |
| no-uncleared-race-timeout | CallExpression 214 |
| require-blocking-standard-streams | SourceFile 307 |
| react-hooks/globals | Identifier 79 |
| react-hooks/immutability | SourceFile 307 |
| react-hooks/no-deriving-state-in-effects | SourceFile 307 |

`validate_listeners.py` builds the owned native probe under ASan/UBSan using the
sanitized checker archive, then compares its complete output against a Go
constant oracle. It mutates the require rule's call subscription from 214 to 215.
The mutant compiles, exits zero with empty stderr, and the independent Go bytes
catch it. An initial test-formatting error printed Go Kind names through its
Stringer; the oracle now explicitly converts each kind to int. The portable
validator and all final passing streams are committed.

The original rule agreement gate passed in 232.886 s. Normal and sanitized
controls have 31 findings; compiler 77 has 28 / 20,884 bytes; repository 287 has
two / 19,558 bytes, all identical to Go. Compiling loop/require/cycle mutants
are caught at bytes 56/7433/12229. Released-handle refusal remains exit 70, and
the retained-registry mutant exits zero and fails that expectation.

Process rules match Go on all 100 upstream programs and both corpora normally
and under sanitizers. The callee mutant is caught in six programs at byte 164,
blocking-order mutant in nineteen at byte 1690. Each new question rejects a
released handle before output. Race controls 22 have 13 findings, matching Go
normally and under sanitizers, as do both corpora. Its lost-timeout mutant is
caught by production byte comparison. Globals controls 35 have 34 findings;
both corpora match normally and under sanitizers, and its compiling ancestry
mutant is caught at byte 9207. Both React kernels match Go on 52 sanitized
results; join and dependency-count mutants compile and are caught by kernel
bytes. Those are kernel proofs, not complete production rule mutants.

The existing shared `ParseNode` in main and the harness branch has only
`readonly kind: string` and no numeric kind property. The current shared driver
invokes whole-file `run` methods rather than supplying a node and its numeric
kind. The legacy code therefore still reads string kinds and refetches nodes.
Removing those operations requires the prohibited shared parser/driver API
change. A per-rule string-to-number shim would retain the string read and lookup
cost and would not satisfy the speed rule. This change prepares declarations
only; it does not claim callback migration or a speed improvement. Once that
shared contract exists, the owned callbacks can be migrated to the handed node.

JSX parsing and source-to-React-HIR/SSA remain the separate blockers for full
React parity, detailed in REPORT.md. No new rules were claimed. Main was checked
again before finishing and remained e8ba3d5d; the assigned branch stayed based on
it. Only the assigned branch is pushed. No compiler or bridge production code
changed in this update, so the full ownership/raw-fact gate from LANDING_REPORT.md
was not repeated; rule released-handle and sanitizer checks were repeated.

Reproduce `validate_listeners.py` with --artifacts, --stage0, --archive and
--asan-archive. Other validators retain the invocation forms in the earlier
reports. All test output went to files. listener-validation/ holds logs, result
metadata and compressed base64 raw streams. Concurrent process timing observed
compiler native 3.969416 s versus Go 0.531185 s; repository 0.545114 s versus
0.217052 s. These are concurrent observations; the quiet landing measurements
remain in LANDING_REPORT.md. No full repository Go gate was run.
