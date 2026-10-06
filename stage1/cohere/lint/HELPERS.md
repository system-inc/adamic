# Shared lint helper claim

Owner: codex/lint-helpers. Base: origin/main, merged with codex/lint-inventory.
Inventory denominator: 198 unported syntax rules waiting on helpers, frozen at the inventory's branch observations. Counts below remove helper blockers, not claim rule implementations or fixture parity.

Order and territory:

1. `helpers/option_schema.ts`: strict upstream option schema validation, with `helpers/options_json.ts` as its strict JSON value reader. 103 consumers; 32 have options as their only recorded blocker. Start with these 32 complete schemas. Unknown schema features must refuse explicitly. Per-rule defaults and conversion to a rule's own settings remain the rule worker's responsibility.
2. `helpers/policy_message.ts`: policy message rendering, generated TypeScript catalog in `helpers/testdata`, including object phrase selections and strict placeholder values. 54 consumers; after options, up to 16 additional rules lose their last recorded helper blocker. Catalog loading validation is performed by Go cohere during generation, not reimplemented as a runtime filesystem loader.

One public helper per source file. Tests and generation live under `helpers/`; existing lint entry points are not changed. Compare each helper against pinned Go cohere on its consumer rules, run the same Adamic source on Node and sanitized native, and require a compiling semantic mutant for each source helper. Publish exact supported rule lists and conservative residual blockers after validation. Do not mark any rule ported merely because a helper is available.

This claim precedes implementation. Subsequent evidence and measured counts will replace projected counts here.
