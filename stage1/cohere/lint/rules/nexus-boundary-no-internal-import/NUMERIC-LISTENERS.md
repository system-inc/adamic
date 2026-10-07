Built: numeric SyntaxKind listener declarations for all 17 rules owned by wave1-12, one listener.a per owned directory.
Commits: previous rule tip c665dda5a0e511497befb4d791f5109c6194973d; helper tip b32ed264df65f493965eecb7d4d7038b73543738 is unchanged; both contain current main e8ba3d5d81de4d3773c723914fccd4c76248b965.
Checks: 17 declarations / 741 observation bytes match actual Go, source Node, emitted JavaScript and sanitized native, PASS 18.209s; existing independent rule packages pass; vet exits 0; default harness and registry fail.
Mutants: each rule's first listener kind changed to Unknown (0), all 17 compiling mutants caught only by actual-Go comparison on every backend; nine existing semantic decision/path mutants are rechecked.
Uncovered: live numeric dispatch and supplied-node migration require shared parser/context/driver changes; default integration, eight partial ports, the exact counter and the full repository gate remain blocked or uncovered. No new claims.

The public handoff is listener.a exporting syntaxKinds: readonly number[]. These are pinned typescript-go parser numeric values, observed from cohere 715ba94f3608a6500086b1076ce5cb7e51b836db rather than stock TypeScript enum guesses or a second string-name mapping. Sorting makes declaration order deterministic; it does not prescribe callback execution order.

The Go oracle parses a file in /repo/libraries/nexus/pages/index.tsx containing a next/script import, then calls the real rules' Run methods with applicable options and reads their actual listener map keys. The project boundary receives its required libraryDirectory. Conflicting classes are special: no Program means the normal rule declines, so an oracle-only Go overlay exports the union of actual ListenerKinds() and actual declineListeners() keys, including the conditional SourceFile error listener. No cohere source is edited. This proves listener domains, not whole Tailwind behavior or a live dispatch profile.

The .a comparison entry imports every declaration unchanged. The owned Go test builds actual Go using overlays, compiles the same .a entry to sanitized native and emitted JavaScript, and runs it on source Node. Each mutant changes one numeric declaration to Unknown, which compiles and exits successfully on all three backends; only output comparison kills it. These metadata mutants supplement the earlier semantic rule mutants, not replace findings/fix comparisons.

| Owned directory | Numeric SyntaxKind listeners |
| --- | --- |
| base-correctness-require-orm-column-declare | 173 |
| nexus-consistency-no-boolean-outcome | 265,266 |
| next-google-font-preconnect | 286,287 |
| next-inline-script-id | 285,286 |
| next-next-script-for-ga | 286,287 |
| next-no-before-interactive-script-outside-document | 286,287 |
| next-no-css-tags | 286,287 |
| next-no-document-import-in-page | 273 |
| typescript-no-non-null-asserted-optional-chain | 236 |
| typescript-no-this-alias | 227,261 |
| typescript-no-unnecessary-type-constraint | 169 |
| better-tailwindcss-enforce-shorthand-classes | 214,261,292 |
| better-tailwindcss-no-concatenated-classes | 214,261,292 |
| better-tailwindcss-no-conflicting-classes | 214,261,292,307 |
| nexus-boundary-no-internal-import | 214,273 |
| nexus-boundary-no-nexus-outside-import | 214,273 |
| nexus-boundary-no-project-import | 214,273 |

The speed rule cannot yet be completed against the current shared API: stage1/typescript/parser/nodes.ts declares ParseNode.kind: string, RuleContext.node(index) returns that node, and registry/registry.go generates visit(index, parent) calls after switching on context.node(index).kind as a string. It never hands rules the already-read node. Converting that shared parser representation, context helpers and driver is outside the owned rule directories. Existing index-based visitor implementations therefore remain an explicit migration blocker, not represented as speed-compliant. New numeric declarations are available now as requested and are ignored by today's driver. No new driver, string-to-number lookup, local AST copy or duplicated numeric parser was introduced.

Other landing failures reproduce unchanged: profile_test.go:32 cannot range over portFiles because registration makes it a function; the registry rejects better-tailwindcss-enforce-shorthand-classes as missing a rule descriptor, with eight partial JSX/Tailwind directories listed in LANDING.md. Both affected commands exit 1. All three independent rule packages pass in 5.263s, 12.566s and 24.656s; registry fails in 0.116s. Its descriptor-rejection controls are masked by the missing descriptor and are not credited as mutant successes. Full numeric API integration and final findings/fix parity for partial ports are not claimed.

Commands, with /workspace/adamic-tools/env.sh sourced:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import -run '^TestNumericListenerDeclarations$' -count=1 -v -timeout=15m > /tmp/wave12-kind-listeners.log 2>&1
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m > /tmp/wave12-kind-default.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/rules/next-google-font-preconnect ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import ./stage1/cohere/lint/registry -run '^(TestExtractedDecisions|TestTailwindDecisions|TestSharedNexusParserGaps|TestResolvedImportPaths|TestDeterministicRegeneration|TestDescriptorRejections|TestDuplicateOracleAdapter)$' -count=1 -v -timeout=15m > /tmp/wave12-kind-regressions.log 2>&1
go vet ./stage1/cohere/lint/rules/next-google-font-preconnect ./stage1/cohere/lint/rules/better-tailwindcss-enforce-shorthand-classes ./stage1/cohere/lint/rules/nexus-boundary-no-internal-import > /tmp/wave12-kind-vet.log 2>&1
```

Setup succeeds in 33 seconds: Go, clang, Node and submodules ready at 0s; cache warm 33s; done 33s; nproc 5, four-core quota, 17.6 GB. Test output is preserved in numeric-listeners*.log. New Adamic files all use .a. No shared files, new rule claims or helper claims were changed. The helper branch has no new changes and retains its exact current-main oracle evidence from REFRESH.md. No complete gate, new TypeScript/stage1 corpus replay or fresh throughput sample was run. Existing rate observations remain historical, and no speed improvement is inferred from declarations ignored by the driver.
