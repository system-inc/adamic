The three rule predicates and diagnostic text reproduce cohere's pinned
`internal/lint/rules/nexus/correctness_no_process_exit_after_output.go`,
`correctness_no_uncleared_race_timeout.go`, and
`correctness_require_blocking_standard_streams.go` at
`715ba94f3608a6500086b1076ce5cb7e51b836db`.

The raw syntax-flow service copies the generic control_flow_graph implementation
at that same pin (cfg.go, expressions.go, statements.go, patterns.go, helpers.go
and roots.go). Its rslint MIT provenance and Bytedance/typescript-eslint copyright
attribution are retained in wave_27_syntax_flow.go. Only local identifiers are
prefixed and raw graph framing is added; no lint predicate is copied into Go.
The existing type-aware directory license applies. The independent oracle calls
production cohere rules unchanged and imports no bridge implementation.
