Declared numeric SyntaxKind listener exports for all twelve owned .a rule ports, using the pinned external Go parser values.
Commits: previous rule landing cec59f82; helper branch remains b28c0599; this commit contains the listener declarations and fresh evidence on main e8ba3d5d.
Checks: 472 supported fixtures and 4,176 corpus pairs match Go on source Node, emitted JavaScript and sanitized native; all twelve declarations separately match Go.
Mutants: twelve rule mutants, the message interpolation mutant and a wrong numeric listener value all compile, finish normally and fail only byte comparisons on all three backends.
Not covered: the required supplied-node numeric handlers, blocked by today's shared parser/driver API; no performance improvement or new claim is asserted.

Each rule exports syntaxKinds as a readonly numeric array in its own rule.a. The legacy rule.json remains unchanged: its strict decoder only accepts string kinds and rejects unknown fields. The new numeric exports are available to the incoming shared driver even though today's generated dispatcher ignores them. Values come from the pinned Go parser used as this stage1 parser's oracle, not stock TypeScript's differently numbered enum.

| Rule | Go parser kind names | Numeric listener kinds |
| --- | --- | --- |
| @next/next/no-assign-module-variable | VariableStatement | 244 |
| @typescript-eslint/default-param-last | FunctionDeclaration, FunctionExpression, ArrowFunction, MethodDeclaration, Constructor, GetAccessor, SetAccessor | 263, 219, 220, 175, 177, 178, 179 |
| structure/tailwind-no-physical-direction | StringLiteral, NoSubstitutionTemplateLiteral, TemplateExpression | 10, 14, 229 |
| @typescript-eslint/no-unnecessary-type-constraint | TypeParameter | 169 |
| @typescript-eslint/prefer-as-const | VariableDeclaration, PropertyDeclaration, AsExpression | 261, 173, 235 |
| @typescript-eslint/prefer-enum-initializers | EnumDeclaration | 267 |
| nexus/import-require-node-namespace | ImportDeclaration | 273 |
| structure/network-no-invalidate-cache-literal-key | CallExpression | 214 |
| structure/network-no-string-literal-query | CallExpression | 214 |
| no-multi-str | StringLiteral | 10 |
| no-nonoctal-decimal-escape | StringLiteral | 10 |
| no-octal | NumericLiteral | 8 |

The speed requirement is only partially implemented. Shared ParseNode currently declares readonly kind: string and has no numeric kind field. The shared generated dispatcher uses switch(context.node(index).kind) and calls rule.visit(index), passing no node. Therefore the existing handlers still inspect symbolic kinds and fetch nodes through their context. They cannot consume the requested supplied numeric node against this API without altering shared files or making an unproven cast. No such cast or shared change was made. The incoming parser/driver must expose a numeric node and pass it into the rule callback before the handler conversion can be completed. No speedup is claimed; throughput was not rerun for this declaration-only preparation.

Commands after source /workspace/adamic-tools/env.sh, all comparison commands exit zero:

    python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/verify-syntax-kinds.py --scratch /tmp/wave15-syntax-kind-proof
    python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py --scratch /tmp/wave15-speed-six --typescript /tmp/lint-wave1-15-typescript --mutants
    python3 stage1/cohere/lint/rules/no-multi-str/validate.py --scratch /tmp/wave15-speed-literals --typescript /tmp/lint-wave1-15-typescript --mutants
    python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate-standalone.py --scratch /tmp/wave15-speed-three --typescript /tmp/lint-wave1-15-typescript --mutants
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v

The declaration check runs Go's ast.Kind values, constructs the expected listener rows from the existing descriptors, then executes the actual exported .a arrays on source Node, emitted JavaScript and sanitized native. NumericLiteral 8 -> 9 in no-octal is a deliberate declaration mutant: normal rule comparisons would miss it because the existing dispatcher ignores the export. The new byte comparison catches it on all three backends, each compiled and exited successfully with empty stderr.

The three full owned rule comparisons cover 348 files (77 compiler files and 271 stage1 .ts/.a files), 4,176 pairs in total. TypeScript v6.0.3 is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. Fixture counts and identical bytes: six-rule group 206 / 83,165; literal group 74 / 26,425; final three 192 / 51,192. Corpus identical bytes: six 79,645,656; literals 39,663,897; final three 39,671,239. Complete finding/message/range and fix/suggestion proposal data are compared as recorded in the existing standalone reports. Every original selected Go test family passes. hashes.json records all four sides and compressed Go observations preserve the compared corpus and fixture output.

All ordinary rule mutants were rerun independently. Each compiled, exited normally with empty stderr and differed only in complete output bytes on all three backends:

- next-no-assign-module-variable: "this.context.node(name).text === 'module'" -> "this.context.node(name).text === 'moduleNever'".
- typescript-default-param-last: "this.defaulted(parameter) || this.flag(parameter, 'QuestionToken')" -> "this.defaulted(parameter) || this.flag(parameter, 'NeverQuestion')".
- structure-tailwind-no-physical-direction: "if(directionAware) {" -> "if(directionAware && false) {".
- typescript-no-unnecessary-type-constraint: "kind !== 'AnyKeyword' && kind !== 'UnknownKeyword'" -> "kind !== 'AnyKeyword' && kind !== 'NeverKeyword'".
- typescript-prefer-as-const: "node.text === this.context.node(initializer).text" -> "node.text !== this.context.node(initializer).text".
- typescript-prefer-enum-initializers: "`${position + 1}`" -> "`${position + 2}`".
- nexus-import-require-node-namespace: "'NodeFileSystem'" -> "'NodeWrongFileSystem'".
- structure-network-no-invalidate-cache-literal-key: ".text !== 'invalidate'" -> ".text !== 'invalidateNever'".
- structure-network-no-string-literal-query: "this.binding(unwrapped, node.text)" -> "-1".
- no-multi-str: "raw.includes('\\u2029')" -> "raw.includes('NEVER')".
- no-nonoctal-decimal-escape: "previousNull = digit === '0'" -> "previousNull = true".
- no-octal: "digit <= '9'" -> "digit <= '7'".
- Literal message interpolation: split/join -> replaceAll replacement-string interpolation; findings counts remain unchanged, but message bytes differ.
- Numeric listener declaration: NumericLiteral 8 -> 9, caught only by the new declaration comparison.

The independent one-byte sentinel passes in 0.264s with one native and one Node miss, zero hits. Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, then failed test warm-up at shared profile_test.go:32 because it ranges over portFiles instead of calling it. No successful warm/done timing was printed. Installed Go 1.27.1, clang 20.1.8 and Node 24.19.0 ran the successful comparisons; nproc=5. All test output went to logs.

The existing .a registry discovery, profile compilation, repair serialization and explicit parser refusals remain documented limitations. Nine literal exclusions and the physical-direction JSX probe still explicitly refuse on every backend while Go accepts them. The full gate and default formatter/order/converging fixer are not certified. The shared numeric handler API is an additional blocker under the new speed rule, so no new helper or rule claim was made. Both existing branches still contain the fetched current main e8ba3d5d; the unchanged helper branch has its passing current-main evidence at b28c0599.
