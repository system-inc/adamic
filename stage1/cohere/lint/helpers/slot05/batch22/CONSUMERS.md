# Consumers

Each of decodePropertyEscape, decodeNumericEscape and readClass removes one prerequisite from each of:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Twelve prerequisite occurrences across four rules; zero final blocker removed. Frozen readiness assumes the common AST adapter and does not declare any rule ported. Cumulative slot 05: 62 helpers, 343 prerequisite occurrences across 70 consumers and 50 helper-ready rules under those assumptions.
