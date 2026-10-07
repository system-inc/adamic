# Wave 07 source provenance

The default judgments and diagnostic text reproduce the production rules in
cohere at `715ba94f3608a6500086b1076ce5cb7e51b836db`:

- `internal/lint/rules/core/no_undef_init.go`, originating in ESLint.
- `internal/lint/rules/typescript/prefer_for_of.go`, originating in
  typescript-eslint.
- `internal/lint/rules/typescript/consistent_indexed_object_style.go`,
  originating in typescript-eslint.

The MIT license in this directory is retained. The independent Go oracle calls
these production implementations unchanged. Generated controls extract their
production Go test source literals. The checker fact question carries only
syntax and direct checker symbol/declaration metadata.
