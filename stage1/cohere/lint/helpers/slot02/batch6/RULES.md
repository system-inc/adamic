# Consumers and residual blockers

Each helper removes one dependency from the same six rules: eighteen entries across six distinct rules. Zero final blockers are removed by this batch. This is a helper handoff, not a claim that these rules are fully ported.

| Rule | loadFile | resolveImport | ingestCustomVariant |
|---|---|---|---|
| better-tailwindcss/enforce-canonical-classes | yes | yes | yes |
| better-tailwindcss/enforce-consistent-class-order | yes | yes | yes |
| better-tailwindcss/enforce-consistent-variant-order | yes | yes | yes |
| better-tailwindcss/enforce-shorthand-classes | yes | yes | yes |
| better-tailwindcss/no-conflicting-classes | yes | yes | yes |
| better-tailwindcss/no-unknown-classes | yes | yes | yes |

readiness.json removes this slot's eighteen delivered helpers only. Other workers' implementations and integration are not assumed.
