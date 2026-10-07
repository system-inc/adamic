# Slot 05 consumer handoff

These lists come from the complete frozen readiness ledger. A removed dependency is not an implemented rule. The common AST adapter is still assumed. Imports removes the final listed helper blocker for exactly two rules:

- `@next/next/no-location-assign-relative-destination`
- `@typescript-eslint/no-import-type-side-effects`

## react.isIdentifierNamed

25 consumers:

- `react/no-arrow-function-lifecycle`
- `react/no-children-prop`
- `react/no-did-mount-set-state`
- `react/no-did-update-set-state`
- `react/no-direct-mutation-state`
- `react/no-string-refs`
- `react/no-this-in-sfc`
- `react/no-typos`
- `react/no-unused-class-component-methods`
- `react/no-will-update-set-state`
- `react/prefer-es6-class`
- `react/prefer-stateless-function`
- `react/require-optimization`
- `react/require-render-return`
- `react/state-in-constructor`
- `react/void-dom-elements-no-children`
- `structure/consistency-require-matching-file-name`
- `structure/next-no-page-state`
- `structure/react-component-no-destructuring`
- `structure/react-component-no-display-name`
- `structure/react-component-no-forward-ref`
- `structure/react-component-no-separate-named-export`
- `structure/react-component-require-properties-parameter`
- `structure/react-hook-no-properties-in-dependencies`
- `structure/react-hook-require-effect-comment`

## imports.BindingsOf

15 consumers:

- `@next/next/inline-script-id`
- `@next/next/no-location-assign-relative-destination`
- `@next/next/no-unwanted-polyfillio`
- `@typescript-eslint/no-import-type-side-effects`
- `base/security-require-context-access`
- `no-restricted-imports`
- `react-hooks/config`
- `react-hooks/incompatible-library`
- `react/no-deprecated`
- `react/no-typos`
- `structure/consistency-require-organized-imports`
- `structure/import-require-react-namespace`
- `structure/network-no-forbidden-import`
- `structure/next-require-page-default-export`
- `structure/react-component-no-forward-ref`

## react.isComponentBase

14 consumers:

- `react/no-arrow-function-lifecycle`
- `react/no-did-mount-set-state`
- `react/no-did-update-set-state`
- `react/no-direct-mutation-state`
- `react/no-string-refs`
- `react/no-this-in-sfc`
- `react/no-typos`
- `react/no-unused-class-component-methods`
- `react/no-will-update-set-state`
- `react/prefer-stateless-function`
- `react/require-optimization`
- `react/state-in-constructor`
- `structure/consistency-require-matching-file-name`
- `structure/react-component-no-display-name`

The 54 dependency occurrences overlap across 38 distinct rules. The frozen helper-ready count rises from 46 to 48 using only this slot. Remaining dependencies for every affected rule are retained in [readiness.json](readiness.json). No global ledger or rule status is rewritten.
