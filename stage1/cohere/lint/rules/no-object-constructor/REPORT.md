# no-object-constructor

Upstream: `cohere/internal/lint/rules/core/no_object_constructor.go`. The descriptor declares no program reads, exactly as upstream. Constructor resolution uses the node-local node-symbol-details question; the syntax and preceding-semicolon analysis lives in this rule alone. Messages and ordered suggestion edits are copied verbatim.

The facts-only harness initially refused upstream JSX captured under a `.ts` filename (`cohere/internal/lint/rules/core/no_object_constructor_test.go:55`, filename at line 11). Area's captured-recovery metadata and parser recovery were merged without shared-file edits. The restored port is tested on all 59 captured upstream cases and a valid JSX witness, without narrowing its upstream prefix.

The earlier failure and smallest input are retained in the sibling `react-jsx-no-undef/validation/parked-object-constructor` directory. Final results are recorded separately; that old note is historical evidence, not a silent skip.
