# Utility compilation and React compiler options

`compileCandidate` accepts a candidate view, a removal mode and lookup, deep-clone, walk and removal callbacks. The walk receives a fresh evaluation state. Its closure captures the candidate value and modifier payloads and the evaluator theme. The helper enforces Go's four postconditions, then selects identity-based dropped declarations and ratio-splice declarations. Mode 0 applies both removals, mode 1 keeps dropped declarations, and mode 2 keeps non-ratio declarations. `compile` is the public mode-0 wrapper. Rejected results have ok false and an empty projection; callers must treat that flag as a nil Go result, not a successful empty body.

A real caller supplies CSS node objects or arena identities as T. The callbacks must deep-clone definitions, preserve identity within that clone, resolve functions without mutating the definition, and remove by identity. `unionNodeSets` comes from this slot's previously verified wave12. Clone, walk and removal are explicitly separate helper dependencies, not additional implementations claimed here.

`decodeCompilerRuleOptions` returns a fresh empty options object and either empty error text or exact Go refusal prose. Empty raw input and an empty decoded object are accepted. JSON null, malformed JSON and non-object values are refused. Non-empty objects are refused with sorted distinct decoded keys. Its JSON callback must follow encoding/json, including last duplicate key, and its sorting callback must follow Go UTF-8 byte order, which differs from JavaScript UTF-16 order for some keys. Inputs here are string configurations representable as valid UTF-8; arbitrary invalid raw bytes are not represented by this API. Those byte inputs require a byte adapter before integration. No regex matcher is introduced.

The private Go overlay calls actual UtilityEvaluator.compile, UtilityEvaluator.Compile and react.DecodeCompilerRuleOptions. It observes cloneNodes, the real walk and removeNodes separately to provide callback inputs and outputs. The Adamic driver compares postconditions and selected identity sets, dependency order, fresh state, unchanged definitions, and canonical Go node JSON. It does not implement a native CSS walker or remover. Consumer options are re-encoded already-decoded Go option objects, not the original configuration file bytes.

Run after sourcing `/workspace/adamic-tools/env.sh`:

```
python3 stage1/cohere/lint/helpers/slot04_wave22/testdata/regenerate.py
python3 stage1/cohere/lint/helpers/slot04_wave22/testdata/capture.py
go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave22
```

Redirect each command to its own log. Capture uses temporary Go overlays and a temporary Tailwind 4.3.3 install; it edits no shared harness or upstream file. REPORT.md records validation, mutants, consumers and limits. readiness.json subtracts only these three symbols from the frozen inventory.
