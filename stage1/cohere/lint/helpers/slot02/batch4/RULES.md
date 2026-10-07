# Consumers and remaining dependencies

Each helper removes one dependency from each of the same six rules (18 dependency edges). These rules still have other blockers in the frozen ledger; this is not a claim that the rules are fully ported.

| Rule | isValidThemePrefix | namespaceForVariantRoot | convertUnderscoresToWhitespace |
|---|---|---|---|
| better-tailwindcss/enforce-canonical-classes | yes | yes | yes |
| better-tailwindcss/enforce-consistent-class-order | yes | yes | yes |
| better-tailwindcss/enforce-consistent-variant-order | yes | yes | yes |
| better-tailwindcss/enforce-shorthand-classes | yes | yes | yes |
| better-tailwindcss/no-conflicting-classes | yes | yes | yes |
| better-tailwindcss/no-unknown-classes | yes | yes | yes |

See readiness.json for residual dependencies after this slot’s twelve delivered helpers only. Other workers’ implementations and integration status are not inferred.
