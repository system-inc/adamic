# CFG Build

build.a ports the public Go control_flow_graph.Build orchestration for valid, non-null IsRoot AST views. BuildOps supplies fresh builder construction, statement/parameter/expression traversal, block identity and graph storage. Source files, static blocks, properties and function roots follow Go callback order. Entry flags, reachable final marking and endReachable are owned here. This is not a complete CFG or native rule implementation.

Run `go test -count=1 -v -timeout=20m ./stage1/cohere/lint/helpers/slot04_wave23` with the cloud environment sourced, redirecting output to a log. The private Go overlay preserves Build and real block allocation/final marking while recording subordinate traversal callbacks. Source Node, emitted JavaScript and sanitized native compare with that oracle.
