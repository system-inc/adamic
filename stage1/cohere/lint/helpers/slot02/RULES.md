# Slot 02 rule dependency handoff

The frozen 198-rule ledger had 46 helper-ready rules at this branch base. These three helpers remove 47 listed consumer dependencies across 47 distinct rules. Only `complexity` and `grouped-accessor-pairs` lose their final listed helper blocker from this slot alone, bringing that conditional readiness count to 48. No rule is claimed implemented. Other workers' completed dependencies are not assumed here.

## jsx_attribute_name.a

21 consumers. Final listed blocker removed: none.

- `@next/next/google-font-display`
- `@next/next/google-font-preconnect`
- `@next/next/inline-script-id`
- `@next/next/next-script-for-ga`
- `@next/next/no-before-interactive-script-outside-document`
- `@next/next/no-css-tags`
- `@next/next/no-html-link-for-pages`
- `@next/next/no-page-custom-font`
- `@next/next/no-styled-jsx-in-document`
- `@next/next/no-sync-scripts`
- `@next/next/no-unwanted-polyfillio`
- `react/forbid-dom-props`
- `react/jsx-key`
- `react/jsx-no-duplicate-props`
- `react/jsx-no-script-url`
- `react/jsx-no-target-blank`
- `react/no-children-prop`
- `react/no-danger`
- `react/no-string-refs`
- `react/no-unknown-property`
- `react/void-dom-elements-no-children`

## property_name.a

14 consumers. Final listed blocker removed: complexity, grouped-accessor-pairs.

- `@typescript-eslint/no-dupe-class-members`
- `complexity`
- `consistent-return`
- `grouped-accessor-pairs`
- `nexus/localization-no-untranslated-value`
- `no-dupe-class-members`
- `no-empty-function`
- `no-prototype-builtins`
- `object-shorthand`
- `react/jsx-props-no-spread-multi`
- `react/no-render-return-value`
- `structure/network-no-invalidate-cache-in-on-success`
- `structure/network-require-hook-variables-type`
- `yoda`

## tailwind_reader_for.a

12 consumers. Final listed blocker removed: none.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-concatenated-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

