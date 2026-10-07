Built: numeric listener declarations in listener.a for all eleven existing owned rules; no new rule or helper claimed.
Commits: declaration implementation follows this report; helper branch b272eab3 remains rebased and green on unchanged origin/main e8ba3d5d.
Commands and outputs: numeric declaration comparison and compiling mutant PASS 0.570s; owned package vet PASS; setup exits 1 at the known shared profile API mismatch, nproc 5.
Mutant: SourceFile 307 changed to VariableDeclaration 261, compiles and completes cleanly on source Node, emitted JavaScript and ASan/UBSan native; only pinned Go subscription comparison catches it.
Not covered: numeric execution dispatch, per-rule node refetch removal, whole-rule oracle refresh on main, new helpers or a full gate; shared registration rebase remains blocked.

Each owned directory now exports listenerSyntaxKinds: readonly number[] from
listener.a. Values come directly from the pinned cohere TypeScript parser's
ast.Kind through its generated shim, the same source used by registration's
current kind-name validation. No hand-guessed enum or runtime string-to-number
conversion was added. The current generator ignores this companion declaration,
as allowed by the speed-rule handoff instruction; the future kind-indexed driver
can import it by the same name from every rule directory. Existing rule.json is
unchanged because the current descriptor decoder rejects extra fields.

| Owned rule directory | Numeric parser kinds |
|---|---|
| consistent-this | SourceFile 307 |
| func-name-matching | VariableDeclaration 261, BinaryExpression 227, PropertyAssignment 303, PropertyDeclaration 173 |
| eslint-comments-require-description | SourceFile 307 |
| next-google-font-display | JsxOpeningElement 287, JsxSelfClosingElement 286 |
| tailwind-no-physical-direction | StringLiteral 10, NoSubstitutionTemplateLiteral 14, TemplateExpression 229 |
| typescript-no-non-null-asserted-optional-chain | NonNullExpression 236 |
| typescript-no-non-null-assertion | NonNullExpression 236 |
| typescript-no-this-alias | VariableDeclaration 261, BinaryExpression 227 |
| tailwind-important-position | JsxAttribute 292, CallExpression 214, VariableDeclaration 261 |
| tailwind-variable-syntax | JsxAttribute 292, CallExpression 214, VariableDeclaration 261 |
| tailwind-variant-order | SourceFile 307, JsxAttribute 292, CallExpression 214, VariableDeclaration 261 |

Eleven declarations, twenty-five subscriptions. listener_test.go compiles all
owned declarations together, comparing their exact output with ids read from
actual pinned Go ast.Kind. The wrong-kind mutant changes only a scratch copy of
consistent-this/listener.a. It is not credited for a compile, runtime or sanitizer
failure. This proves the declaration contract, not a semantic finding catch or a
kind-indexed dispatch speedup. Earlier rule-semantic mutants remain historical
validation of unchanged rule logic.

Command, with source /workspace/adamic-tools/env.sh:
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/rules/func-name-matching
-run '^TestNumericListenerDeclaration' -count=1 -v -timeout=10m
> /tmp/wave104-numeric-listeners-tests.log 2>&1: PASS 0.570s.
go vet ./stage1/cohere/lint/rules/func-name-matching: exit zero, empty log.
All logs are retained under evidence/numeric-listeners-*.log.

Current shared API blocker: stage1/typescript/parser/nodes.ts declares
ParseNode.kind: string and has no numeric SyntaxKind field. RuleContext and the
generated visit hook take indexes, not an already fetched node. Existing rules
still use that API. Therefore these declarations do not claim execution complies
with the full no-string-kind/no-refetch speed contract. A rule-local cast or a
per-rule string lookup would merely hide or repeat the work the speed rule forbids;
none was added. The parser/driver owner must supply numeric kinds and the handed
node before the owned implementations can be converted without editing shared
files. This note preserves the exact boundary instead of claiming a speedup.

Landing remains blocked: fresh all-origin fetch still shows main e8ba3d5d and
registration 48ecd930. git rebase origin/main again stops at inherited commit
29175443 with conflicts in shared README.md, lint.ts, lint_test.go and
testdata/oracle.go. The rebase was aborted without editing those files. Ahra's
shared-file ownership restriction and landing-first cap remain in force. No new
helper reservation was made. The helper branch remains published as b272eab3,
with its previous 18.910s differential/mutant gate on the unchanged current main.

The required bash cloud/setup.sh reached Go, clang, Node and submodules ready at
0s, then cache warming exited 1 because shared profile_test.go:32 ranges over
portFiles as a slice after registration changed it to a function. There is no
cache-warm or done line. The owned package test bypasses that unrelated shared
package compile failure without patching it. Tools: Go 1.27.1, clang 20.1.8,
Node 24.19.0; nproc 5. No shared generator, harness, parser or compiler source was
edited. The current work stops at the shared landing and numeric-node handoffs.
