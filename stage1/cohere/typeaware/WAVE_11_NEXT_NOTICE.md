# Wave 11 continuation provenance

The native decisions and messages reproduce these unchanged production cohere
rules at `715ba94f3608a6500086b1076ce5cb7e51b836db`:

- `internal/lint/rules/nexus/correctness_no_collection_misuse.go`
- `internal/lint/rules/nexus/correctness_no_discarded_outcome.go`
- `internal/lint/rules/nexus/correctness_no_discarded_pure_result.go`

The MIT license already supplied in this directory applies to the ports.
`testdata/oracle_wave_11_next.go` invokes the original registry rules unchanged,
with its own program loader, AST traversal and complete diagnostic serialization.
It imports no bridge implementation.

`declaration-lineage` returns raw symbol declarations, every parent node's
identity/kind/name/file, source-library provenance, child identities and the
alias's declared-type link. It does not identify known outcomes or pure methods.
`awaited-type-shape` returns the checker's awaited type graph. The Adamic rules
select the production declaration names and paths, group union arms, classify
collection operations and callable arguments, and decide finding messages/ranges.
Neither question returns findings, fixes or suggestions.
