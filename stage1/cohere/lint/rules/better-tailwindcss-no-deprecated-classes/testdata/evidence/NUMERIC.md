Built: numeric SyntaxKind declarations for all seven owned rules, in each numeric_listener.a.
Commits: parent d3c88b90 remains based on origin/main e8ba3d5d; declarations and this evidence follow on codex/lint-wave1-01.
Checks: TestNumericListeners PASS 2.253s, 329 Go-identical bytes/backend; registry PASS 0.038s.
Mutants: extra KindUnknown subscription per rule compiles/runs, then comparison alone catches every mutation on all three backends.
Not covered: numeric dispatch integration, handed-node listener bodies, speed improvement, full rule certification or a new corpus/full repository gate.

Declarations export readonly numeric syntaxKinds. Values come from the pinned parser's actual Go shim constants, not stock TypeScript constants or a locally enumerated name list. The oracle independently reads each existing rule.json name list and converts names using Go ast.Kind constants. The numeric metadata is inert in today's shared driver, as authorized in the speed instruction. Existing kind-name descriptors stay intact for current registration. Each declaration lives only in its owning rule directory; no shared generator, context, parser or compiler file changes.

Numeric listeners:

- assertions: AsExpression 235, TypeAssertionExpression 217.
- physical direction: StringLiteral 10, NoSubstitutionTemplateLiteral 14, TemplateExpression 229.
- require description: SourceFile 307.
- Google font display: JsxOpeningElement 287, JsxSelfClosingElement 286.
- deprecated classes: JsxAttribute 292, CallExpression 214, VariableDeclaration 261.
- duplicate classes: JsxAttribute 292, CallExpression 214, VariableDeclaration 261.
- unknown classes: SourceFile 307, JsxAttribute 292, CallExpression 214, VariableDeclaration 261. Its SourceFile entry preserves the explicit missing-program refusal.

Observed API blocker: stage1/typescript/parser/nodes.ts ParseNode.kind is a string; RuleContext.node returns that string-kind node; the current shared RuleSet calls visit(index). It supplies neither a numeric-kind node nor a fetched node argument. Therefore the existing rule bodies still read kind strings and refetch nodes, and the new speed rule is only partially implemented. Converting these bodies without an authoritative numeric node field and handed-node callback contract would require shared parser/context/driver changes prohibited by this unit's ownership instructions. No string-to-number hot-loop shim or guessed secondary numeric enum is added. Once the shared driver lands, its callback contract must be used by each owned listener.

Initial development comparison failed (retained numeric-initial-failure.log): Python enumeration omitted a token enum row and generated values one below actual Go after the token range. Regenerating from testdata/numeric_kinds.go corrected this. This unplanned failure is not counted as the deliberate mutant. Explicit mutants insert numeric KindUnknown 0 into one rule declaration at a time; seven scratch builds and all 21 backend observations execute successfully and differ from the independent Go oracle. This proves metadata checking, not driver dispatch semantics.

Commands (output redirected to retained logs):

    bash cloud/setup.sh
    source /workspace/adamic-tools/env.sh
    go run stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes/testdata/numeric_kinds.go
    go test ./stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes -run '^TestNumericListeners$' -count=1 -timeout 10m -v
    go test ./stage1/cohere/lint/registry -count=1 -timeout 10m -v

Setup timings: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, cache warm and total 28s, nproc=5. Only added declarations/test/generation/evidence changed; prior runtime rule sources and last same-base oracle results remain unchanged. Landing still blocked by shared allowLoop routing, multiple-edit representation, live Tailwind program resolution and configured regex provider gaps in LANDING.md. No helper claimed or built. Current main ancestry verified; only the own branch is pushed.
