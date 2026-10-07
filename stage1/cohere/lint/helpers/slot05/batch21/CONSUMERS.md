# Consumers

Each of joinRanges, caseExtras and buildCaseTables removes one prerequisite from each of:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Twelve prerequisite occurrences across four rules; zero final blocker removed. Frozen readiness depends on the common AST adapter assumption and does not declare these rules ported. Cumulative slot 05: 59 helpers, 331 prerequisite occurrences across 70 consumers and 50 helper-ready rules under those assumptions.
