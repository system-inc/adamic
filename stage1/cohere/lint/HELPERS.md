# Shared lint helper claim

Owner: codex/lint-helpers. Base: origin/main, merged with codex/lint-inventory.
Inventory denominator: 198 unported syntax rules waiting on helpers, frozen at the inventory's branch observations. Counts below remove helper blockers, not claim rule implementations or fixture parity.

Order and territory:

1. `helpers/option_schema.ts`: strict upstream option schema validation, with `helpers/options_json.ts` as its strict JSON value reader. 103 consumers; 32 have options as their only recorded blocker. Start with these 32 complete schemas. Unknown schema features must refuse explicitly. Per-rule defaults and conversion to a rule's own settings remain the rule worker's responsibility.
2. `helpers/policy_message.ts`: policy message rendering, generated TypeScript catalog in `helpers/testdata`, including object phrase selections and strict placeholder values. 54 consumers; after options, up to 16 additional rules lose their last recorded helper blocker. Catalog loading validation is performed by Go cohere during generation, not reimplemented as a runtime filesystem loader.

One public helper per source file. Tests and generation live under `helpers/`; existing lint entry points are not changed. Compare each helper against pinned Go cohere on its consumer rules, run the same Adamic source on Node and sanitized native, and require a compiling semantic mutant for each source helper. Publish exact supported rule lists and conservative residual blockers after validation. Do not mark any rule ported merely because a helper is available.

This claim precedes implementation. Measured evidence follows; the projection above is retained as the initial claim.


## Measured handoff

The published claim was commit `acb5e0f`, pushed before implementation. Four separate files deliver the two highest-yield helper families: the JSON prerequisite, upstream schema validation, strict target validation, and policy rendering. The initial 48-rule projection was reduced to **46**, because `boundaries/dependencies` and `no-extra-boolean-cast` still need custom option work. `helpers/readiness.json` records the 198-rule ledger; the original inventory is unchanged.

| Helper | Supported consumers | Additional fully helper-ready rules |
|---|---:|---:|
| OptionsJson | prerequisite for 103 | 0 |
| OptionSchema | 94 upstream schemas | 0 alone |
| StrictOptions | 97 Go target types | 30 with JSON/schema prerequisites |
| PolicyMessage | 54 | 16 after options |
| Total | overlapping consumers, not a sum | 46; 152 still have listed blockers |

“Helper-ready” means no remaining helper in the frozen inventory list, assuming its common AST adapter. No rule is marked ported. Rule-local defaults, custom decoding, fixes, full fixture parity and integration remain the rule worker's work. Validation and limits are detailed in `helpers/README.md` and `helpers/REPORT.md`.

### Options complete the last recorded shared helper for these 30 rules

- `@typescript-eslint/consistent-type-assertions`
- `@typescript-eslint/consistent-type-definitions`
- `@typescript-eslint/init-declarations`
- `@typescript-eslint/no-explicit-any`
- `@typescript-eslint/no-inferrable-types`
- `@typescript-eslint/no-restricted-types`
- `@typescript-eslint/no-unused-expressions`
- `@typescript-eslint/unified-signatures`
- `class-methods-use-this`
- `consistent-this`
- `func-name-matching`
- `max-classes-per-file`
- `max-depth`
- `max-nested-callbacks`
- `no-cond-assign`
- `no-inner-declarations`
- `no-restricted-properties`
- `no-self-assign`
- `no-underscore-dangle`
- `no-unsafe-negation`
- `no-unsafe-optional-chaining`
- `no-useless-computed-key`
- `react-hooks/gating`
- `react/forbid-foreign-prop-types`
- `react/jsx-no-useless-fragment`
- `react/no-invalid-html-attribute`
- `react/no-unescaped-entities`
- `react/no-unsafe`
- `react/self-closing-comp`
- `sort-vars`

### Policy rendering completes 16 more after options

- `adamic/no-definite-assignment`
- `base/boundary-no-global-container`
- `base/consistency-no-hand-built-declared-error`
- `base/consistency-require-pagination-argument-name`
- `base/correctness-require-orm-column-declare`
- `nexus/consistency-no-boolean-outcome`
- `nexus/consistency-no-for-in`
- `nexus/consistency-no-hand-rolled-delay`
- `nexus/consistency-no-return-void`
- `nexus/consistency-no-stuttering-name`
- `nexus/consistency-no-utils-folder`
- `nexus/import-require-module-alias`
- `nexus/import-require-node-namespace`
- `structure/network-no-invalidate-cache-literal-key`
- `structure/network-no-string-literal-query`
- `structure/tailwind-no-physical-direction`

### Remaining option gaps

- `boundaries/dependencies`: custom/unsupported Go target decoder, regex schema or boundaries extension
- `id-length`: custom/unsupported Go target decoder
- `nexus/import-require-path-alias`: custom/unsupported Go target decoder
- `no-constant-condition`: custom/unsupported Go target decoder
- `no-extra-boolean-cast`: custom/unsupported Go target decoder
- `no-restricted-exports`: regex schema or boundaries extension
- `object-shorthand`: custom/unsupported Go target decoder

## Next claim: parser-anchored comments

Claimed before implementation on the same branch, following the 46-rule handoff. Territory: new `helpers/comments/` files and this claim. Existing rule entry points and first-unit helper files remain owned by their current workers.

Next bounded dependency bundle: `comments.All` and `comments.ForFile`, with `canBeginAt`, `collectListInteriors`, and `sortByPosition` in separate source files. Together these five symbols serve 24 of the 152 remaining rules and remove the last recorded helper blockers from 16. None removes a last blocker individually; the cache completes this bundle. The next package bundles are JSX (8 rules), imports (7), and property names (5). React has 18 package-only candidates but spans 64 symbols; it needs smaller dependency-set ranking rather than treating its whole family as a single helper.

The claim covers comment collection, source-order deduplication, UTF-8 ranges and byte columns, EOF/shebang/trailing comments, parser-owned empty and trailing-comma list interiors, and a cache explicitly scoped to one immutable parsed file. Compare actual Go cohere helpers on consumer fixtures, Node and sanitized native running identical TypeScript, plus a compiling semantic mutant for each helper file. Unsupported parser shapes must refuse explicitly. No rule is claimed implemented. Publish measured readiness only after those comparisons pass.

## Second measured handoff: comments

Claim `29990b4` was pushed before implementation. Five separate helper files implement the parser-anchored comment scan and per-file sharing; `comment.ts` is its result data model. Original Go consumer fixtures, boundary inputs and compiling mutants hold the implementation. See [the API and evidence boundary](helpers/comments/README.md), [the report](helpers/comments/REPORT.md), and [the residual ledger](helpers/comments/readiness.json).

| Helper, in prerequisite order | Consumers of the completed scan bundle | Additional fully helper-ready rules |
|---|---:|---:|
| `canBeginAt` | 24 | 0 alone |
| `collectListInteriors` | 24 | 0 alone |
| `sortByPosition` | 24 | 0 alone |
| `All` | 24 | 0 until `ForFile` is available |
| `ForFile` | 24 | 16 with the scan prerequisites |
| Cumulative first and second units | overlapping consumers | 62 of 198; 136 still helper-blocked |

These remain inventory dependency counts under its common AST adapter assumption. The helpers are compared on all 4,513 captured inputs from 24 consumers, using only Go's actual AST geometry as adapter input and the unmodified Go helpers as the answers. Forty-seven boundary sources also use independently parsed Adamic trees. The owned independent-parser adapter explicitly refuses four valid JSX boundary sources; the underlying parser was observed to silently reinterpret one before this guard. This is still an integration gap for `react/jsx-curly-brace-presence`, even though the comment helper works with its JSX AST. No rule is claimed ported or fully integrated.

The 16 additional candidates:

- `@typescript-eslint/ban-tslint-comment`
- `@typescript-eslint/no-invalid-this`
- `@typescript-eslint/prefer-function-type`
- `arrow-body-style`
- `default-case`
- `max-lines`
- `no-extra-bind`
- `no-extra-label`
- `no-fallthrough`
- `no-inline-comments`
- `no-invalid-this`
- `no-irregular-whitespace`
- `no-unused-labels`
- `no-useless-rename`
- `operator-assignment`
- `react/jsx-curly-brace-presence`

Remaining adjacent work includes JSX (8 package-only candidates), imports (7 package-only candidates; `BindingsOf` alone adds 4 after this comment bundle), and property names (5 package-only candidates). The 136-rule residual ledger should drive the next claim; these projections are not implemented here. `Comment.IsJsDoc`, `ContentLines`, `LeadingRunFor`, directives and rule-local options are outside this claim.
