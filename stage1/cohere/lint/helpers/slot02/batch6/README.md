# Slot 02 sixth helper batch

| File | Go contract |
|---|---|
| stylesheet_load_file.a | stylesheetCollector.loadFile: absolute path, visiting-stack cycle check, read, parse, stylesheet append and ingest in Go order; preserve nested error causes and remove the visiting mark after returned errors. |
| stylesheet_resolve_import.a | stylesheetCollector.resolveImport: trim/segment params, strip all edge quote characters, accept only empty or source(...) modifiers, resolve against the parent directory, and preserve exact errors and causes. |
| ingest_custom_variant.a | stylesheetCollector.ingestCustomVariant: Go TrimSpace, bracket-aware segment on an ASCII space, trim the first segment, remove one trailing -*, skip an empty name, and set its map value to true. |

These are three helpers with six consumers each. RULES.md lists the six rules and readiness.json records their residual blockers. Eighteen dependency entries are removed; no rule loses its final blocker in this batch. The ports do not implement whole rules.

Loading receives an initialized visiting map and stylesheet list. The visiting set is an active stack, not a permanent seen set: completed loads can be repeated, while a currently true mark is a cycle. A false existing entry permits loading and is removed afterward. Resolve failures precede visiting writes; read and parse failures wrap their original cause and clean up the mark. Successful parsing appends the absolute path before invoking ingest, and its returned error is passed through unchanged. The loader does not roll back stylesheet append on an ingest error, because Go does not.

Path resolution, filesystem read, CSS parsing and ingestion are explicit dependencies. Node and error handles refer to externally owned arenas; -1 denotes no error. describe reads an existing error; makeError allocates a new record retaining the original cause handle. This preserves Go's %w cause chain rather than just its final text. The parser supplies a handle for the exact parsed node list, passed unchanged to ingest. The caller owns ingestion side effects and recursive import loading. Dependencies return normally with error handles; panicking adapters and nil receivers/maps are outside this typed projection. The visiting cleanup comparison concerns returned errors, not panic unwinding.

Import resolution delegates Go TrimSpace, bracket-aware segment, filepath.Dir, quote formatting and the resolver to explicit dependencies. It owns quote removal, modifier iteration, source-prefix acceptance, call order and exact diagnostic assembly. Quotes at either edge are removed repeatedly regardless of whether they match, as strings.Trim does. Whitespace-only segments are skipped. source(unfinished is accepted as a prefix, while source without the parenthesis is refused. Refusals stop before invoking the resolver. Unsupported modifier errors have no cause; resolver errors retain the original cause handle. Dense string lists are required and malformed adapter elements are refused.

The custom-variant helper receives the separately owned segment function and calls it once, even on empty params. It records only the name, as Go does, without interpreting selector bodies. Strings must be valid Unicode text; Go's White_Space set includes NEL and excludes BOM and zero-width space. The callback returns a dense list of strings. The first-element check reports a malformed adapter instead of proceeding with a missing element. Arbitrary invalid UTF-8 bytes and unpaired UTF-16 surrogates are outside this projection.

The owned oracle overlays export the actual private Go helpers. Only dependency call sites are wrapped: filesystem/path/parser/ingest operations and segmentation. Actual Go implementations remain unchanged. Ordered calls, the active visiting mark and stylesheet append at ingestion, unchanged parsed-list identity, and nested error chains are recorded. Temporary real files exercise missing files, parse and ingest failures, successful loads, active cycles and initially false marks. The corpus supplies independent dependency results; expected Want is removed before Adamic reads it. No Node implementation decides the Go verdict.

157 actual rule/file/source captures cover every consumer. Their decoded TypeScript literal/template text, supplementary Go theme/variant/utility fixture literals, real temporary stylesheet graphs, import modifiers, whitespace boundaries, visiting-stack state and nested error causes feed the real helpers. Coverage is bounded and does not claim whole-rule findings/fixes parity. The capture exits 1 on eight known unavailable external Tailwind/corpus gates; these are separately retained, not passing rule gates. Unexpected failures or missing consumers reject regeneration.

Baseline and semantic mutants run on source Node, emitted JavaScript and ASan/UBSan native. All three must finish successfully and agree before comparison to real Go can credit a semantic mutant. Compile failures, crashes and sanitizer failures are not mutant kills.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch6/testdata/regenerate.py > /tmp/slot02-batch6-capture.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/slot02-batch6-helpers.log 2>&1
```

Ownership: NewUtilityEvaluator and addRepositoryFunctionalRoots were withdrawn to slot 05 after a claim race; neither duplicate port is delivered. The retained loader/import claims were pushed in 7767668 before implementation.
