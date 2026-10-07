# Ninth batch consumers

These are removed helper dependencies, conditional on the common AST adapter and callback dependencies. They do not establish complete rule parity.

## imports.ImportedNameOf

Remaining consumers: 4.

- react-hooks/config
- react-hooks/incompatible-library
- structure/network-no-forbidden-import
- structure/react-component-no-forward-ref

## regexp.isDecimalDigit

Remaining consumers: 4.

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

## regexp.countGroups

Remaining consumers: 4.

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Twelve dependency occurrences across eight rules; no final listed blocker removed by this batch alone. See readiness.json for every residual dependency. CallExpressionSource and HasAttributeNamed are withdrawn to slot 02 and are not delivered or counted.
