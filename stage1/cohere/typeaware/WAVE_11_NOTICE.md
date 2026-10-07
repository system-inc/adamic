# Wave 11 provenance

The three Adamic rule decisions and messages reproduce these pinned cohere files
at `715ba94f3608a6500086b1076ce5cb7e51b836db`:

- `internal/lint/rules/typescript/prefer_regexp_exec.go`
- `internal/lint/rules/typescript/no_for_in_array.go`
- `internal/lint/rules/nexus/correctness_no_identical_branches.go`

The native sources use the MIT license already supplied in this directory.
The independent oracle invokes all three original production rules unchanged.

`bridge/tsgo/regular_expression_syntax/regular_expression_syntax.go` concatenates
cohere's `internal/lint/ecmascript/regexp/{regexp,rewrite,escape,class,casefold,canonicalize}.go`
into a separate raw syntax substrate. Algorithms are unchanged. Package imports
are consolidated and comments replace em dashes with commas. It uses the pinned
`github.com/dlclark/regexp2/v2` dependency, version `v2.5.2`, under MIT.
This question returns only whether a supplied regex pattern compiles. It never
returns a lint verdict, message, range, fix or suggestion. Adamic decides which
pattern to ask about, whether to report, and exactly what edit to propose.
