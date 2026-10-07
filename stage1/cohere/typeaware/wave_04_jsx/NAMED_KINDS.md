Corrected all twelve owned listener manifests and module declarations to ast.Kind names.
Commits: builds on 325cf6e8e; source and evidence are committed together on codex/typeaware-wave-04.
Checks: nine named listener declarations match production Go, all three JSX manifests deserialize, and JSX helper/reporting comparisons pass.
Mutants: nine compiled listener mutations, twelve in-memory manifest mutations, eight reference byte mutants and six reporting/refusal mutants caught.
Not covered: full shared registry discovery, full JSX source ports/corpus parity, source throughput and full repository gate.

The latest user instruction supersedes the previous numeric listener requirement. All twelve `rule.json` kinds and corresponding `listenerKinds` exports now use upstream enum names without the `Kind` prefix, exactly as registry.Discover validates them. The production listener oracle enumerates the actual pinned typescript-go enum instead of a maintained subset table, parses Go listener maps, and independently derives the expected names. Rule code does not add string comparisons to decide dispatch relevance.

The old Descriptor.kinds type mismatch is resolved. The actual registry probe now prints `<nil>` for all three JSX manifests. This observation covers deserialization only: these partial manifests are not yet full registry descriptors, lack completed factories/source adapters/oracle witnesses, and remain outside the shared rule discovery tree. No shared registration generator or harness was edited. The private numeric AST snapshots used by existing helper controls remain an isolated testing representation, not an assertion that the shared ParseNode has numeric kinds or that a driver adapter exists.

## Checks run

All commands source /workspace/adamic-tools/env.sh and send output directly to retained logs:

- `python3 stage1/cohere/typeaware/wave_04_react/validate_rule_json.py /workspace/typeaware-wave-04-named/json`: PASS nine production-Go/module/manifest comparisons and nine in-memory Unknown-kind mutants.
- `python3 stage1/cohere/typeaware/wave_04_react/validate_listeners.py /workspace/typeaware-wave-04-named/listeners --adamic /workspace/typeaware-wave-04-harness-ab70/adamic`: PASS nine declarations, 658 identical Go/native/sanitized/source Node bytes. Each of nine mutated module exports compiles and exits zero with empty stderr; only byte comparison catches it. Emitted JavaScript remains unavailable for these checker-importing modules with the current stage-0 JS command.
- `python3 stage1/cohere/typeaware/wave_04_jsx/validate_partial.py /workspace/typeaware-wave-04-named/jsx-reporting-final --adamic /workspace/typeaware-wave-04-harness-ab70/adamic`: PASS 56 records/4941 bytes, Go/native/sanitized/source Node/emitted JavaScript, three byte mutants and three source-refusal mutants. Three JSX module/manifest comparisons use listener names derived by the production Go oracle and reject three in-memory Unknown-kind mutants.
- `python3 stage1/cohere/typeaware/wave_04_jsx/validate_references.py /workspace/typeaware-wave-04-named/jsx-references --adamic /workspace/typeaware-wave-04-harness-ab70/adamic`: PASS 64 snapshot nodes/302 records/3801 bytes, five backends and eight clean sanitized comparison-only mutants.
- `go run stage1/cohere/typeaware/wave_04_jsx/testdata/registry_probe.go` with the three JSX manifest paths: all print `<nil>`. This is an observation probe, not a full discovery test.

Current main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, already an ancestor of the owned branch. All nine listener modules were freshly compiled and compared; no judgment implementation changed in this correction. Previous six completed-rule corpus findings/fixes/suggestions, released-handle, sanitizer and timing evidence is at HARNESS_AB70.md. No new source-rule throughput measurement was made.

Remaining JSX source work is owned implementation work: full descriptors/factories/adapters, fragment alias/attribute/options judgment, undef bindings/declaration-file facts, context provider/component resolution, recursive construction and memo stability/escape analysis, and final source ranges. The former numeric registry compatibility blocker must no longer be cited. HIR-dependent React claims remain parked. No new claims were taken, and the three JSX source rules remain explicitly incomplete.
