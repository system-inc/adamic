The rule decisions and diagnostic messages reproduce pinned cohere
715ba94f3608a6500086b1076ce5cb7e51b836db:

- internal/lint/rules/core/prefer_rest_params.go
- internal/lint/rules/core/prefer_regex_literals.go
- internal/lint/rules/react/exhaustive_deps.go

The native regex helper reproduces the syntax and character-span behavior of
cohere's ecmascript/regexsyntax and ecmascript/regexpattern shelves. Native
reference tracking follows the ecmascript/reference shelf. The existing
license in the type-aware directory applies. The independent Go oracle invokes
the unchanged production rules and imports no bridge implementation.
