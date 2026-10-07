# Slot 01 shared helpers

These three `.a` files port one Go helper each, against cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. They are standalone helper APIs; existing linter entry points and the frozen readiness ledger are unchanged.

| File | API | Contract |
|---|---|---|
| `structure_file_context.a` | `structureFileContext(fileName): FileContext` | All nine Go filename flags. Backslashes become slashes; suffix and substring tests remain case-sensitive. No path cleaning or filesystem lookup. |
| `react_es6_component_class.a` | `reactEs6ComponentClass(nodes, index, isComponentBase): boolean` | Class declarations and expressions, checked heritage/type kinds, then the caller's separately owned base-expression predicate. |
| `tailwind_default_class_literal_settings.a` | `tailwindDefaultClassLiteralSettings(): ClassLiteralSettings` | Exact ordered attribute names, callee names and regex strings, each in an independently mutable fresh array. |

The class adapter is an immutable flat arena of `ReactClassNode`: `kind`, `expression`, `heritage` and `types`. The relevant kind strings are `ClassDeclaration`, `ClassExpression`, `HeritageClause` and `ExpressionWithTypeArguments`; other Go kinds may use their own names or `Other`. Every non-null node keeps its identity as an index. `-1` is a null node or expression, and nil lists are empty arrays. Out-of-range node references panic, rather than silently saying a class is absent. The arena must come from named parser fields, not guesses about child ordering. The test exporter checks that each referenced named field was actually traversed.

`isComponentBase` takes an expression index and must implement Go's `react.isComponentBase` over the same arena. That helper and its leaf predicates belong to other slots; the class file neither copies nor claims them. Go's class predicate considers all heritage clauses without filtering their extends/implements token. The test passes actual Go base-expression answers through this dependency seam, then compares the class result with the actual Go class function. It does not claim end-to-end integration with another worker's arena representation.

`ClassLiteralSettings` has readonly fields containing mutable arrays. The arrays are fresh on every call, including their first elements: custom configuration changes must not alter defaults later returned to another consumer. This file returns regex source strings; compiling those patterns belongs to the separately owned reader.

Run from the repository root after sourcing `/workspace/adamic-tools/env.sh`:

```sh
go test ./stage1/cohere/lint/helpers -run '^TestSlot01' -count=1 -v -timeout=20m > /tmp/lint-helpers-01.log 2>&1
```

`slot01_test.go` reads every consuming rule in the frozen ledger and refuses missing consumer fixture files or count drift. The oracle-only Go overlay calls real private helpers without editing cohere. It extracts statically evaluable string expressions from the consumer Go test files, retaining full constant concatenations, plus explicit path/class controls. Strings include fixture code, filenames, labels and messages; this is a bounded helper corpus, not execution of whole rule fixtures. Dynamically constructed inputs are not evaluated. Each string is a filename input for file context, a TSX parser input for class detection, or a call/freshness repetition for defaults. Every parsed node and an explicit null node are queried for class detection. All nine flags and all list bytes are compared. The same `.a` source runs on Node and sanitized native.

Every delivered helper has a compiling semantic mutant: omit backslash normalization, omit class expressions, or omit the `class` attribute default. Three more mutants share one defaults array across calls and must fail the freshness observations. A sanitizer finding, stderr, panic or compilation failure is not credited as a semantic mismatch. See `SLOT01_REPORT.md` and the retained `evidence/slot01/` logs.

`slot01_readiness.json` lists each consumer and its residual blockers. File context removes the last listed helper for three rules. The other helpers remove dependencies but do not make any rule completely helper-ready alone. No rule is marked implemented.
