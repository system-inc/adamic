# Consumers

`regexp.rewrite` removes a prerequisite for:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

`regexsyntax.IsHexDigit` removes a prerequisite for:

- no-control-regex
- no-regex-spaces
- no-useless-escape

`regexsyntax.AllHexDigits` removes a prerequisite for:

- no-control-regex
- no-regex-spaces
- no-useless-escape

Ten prerequisite occurrences across seven rules; zero final blocker removed by this batch. Frozen readiness assumes the common AST adapter and does not declare any rule ported. Cumulative slot 05: 68 helpers, 365 prerequisite occurrences across 73 consumers and 50 helper-ready rules under those assumptions.
