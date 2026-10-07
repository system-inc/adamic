# CFG evaluation and throwable forks

Three helpers, one production .a file each. updateExpression calls patternReads and then patternWrites with the identical operand, including nil. parameter obtains a parameter projection from its parser adapter, calls expr on its type, then bindWithDefault on the original name and initializer. An absent projection is a no-op. The Go parser cast is an assertion: callers must supply actual parameter nodes, not wrong-kind AST nodes. firstThrowableFork checks current reachability, resolves the enclosing handler, skips absent or already-forked handlers, sets both flags, links the handler, allocates and links a continuation, and enters it. Repeated calls reuse the original fork. Caller adapters supply real graph selection and transitions.

Node projections, pattern processing, default binding, handler selection/target, block creation and graph transitions remain explicit external dependencies. This unit ports their composition, callback order, argument identity and owned state changes, not those separately owned implementations. Generic node/block values preserve identity. Invalid frame indices refuse rather than silently succeeding.

The private Go overlay renames only dependency calls inside the actual three Go methods; it never replaces their bodies. Wrappers execute real Go dependencies and record the helper's direct callback boundary. Nested calls inside an external dependency still execute but their trace belongs to that dependency and is excluded. Production cohere and the shared harness remain unchanged. The Go pin is 715ba94f3608a6500086b1076ce5cb7e51b836db. [NOTICE.md](NOTICE.md) retains the vendored MIT attribution.

All four consumer test files' string literals are inspected, including configuration and expected text. Added controls exercise identifier/property/element updates, default/destructured/rest/typed parameters and Unicode names. The corpus has 801 distinct strings: 13 update operands including nil and 71 parser parameters. Fork tests use 2,560 initial states at depths zero through four, both reachability values, try/catch positions, finally availability, existing flags and absent handler targets. Each fork state is invoked twice: 2,644 case rows, 5,204 total calls. Source Node, emitted JavaScript and sanitized native output must agree byte-for-byte with Go.

Supported states have distinct frame records, an existing current block and immutable parser projections. Invalid/wrong-kind ASTs, forged typed-nil parser data, arbitrary dependency implementations, callbacks that mutate frames during selection, arbitrary graph states and full rule diagnostics are outside this comparison. The absent parameter projection and invalid frame-index refusal are defensive adapter paths, not claims of upstream parser behavior. Fork state controls are synthetic, not extracted full graph executions from consumer rule tests. Every consumer file is inspected, but this is not a replay of its complete findings.

```
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch16 -count=1 -v -timeout=20m > /tmp/lint05-batch16-verified.log 2>&1
```

[REPORT.md](REPORT.md) records the successful run, all fifteen compiling semantic mutants and superseded failures. [CONSUMERS.md](CONSUMERS.md) lists each helper's four rules; readiness counts prerequisite removals, not implemented rules.
