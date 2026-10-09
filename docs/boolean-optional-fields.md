# Boolean optional fields

Built native boolean | undefined fields toward roadmap step 17, including omitted-slot writes, deletion, spreads, unknown reads, static inheritance, and bounded JSON.stringify.
Delivery branch compiler/boolean-optional-fields depends on compiler/optional-presence a774d31656ebbb183d259e445b56135a71d5a1fb; no other unlanded worker branch was merged.
Validation covers build, internal vet, changed .a checks, named stage1 probes, stage3, Linux counts, and Node comparisons in both backends with sanitizers.
Mutants catch dropped domain and presence checks, each retired port's restored refusal, restored static refusal, omitted JSON metadata/origin checks, and an omitted prototype-setter boundary.
Not covered: React IR hir-01 is absent from this base and fetched main; checked-view writes remain acceptance dependency: compiler/views-rehearsal 2b6c032a.

## Representation and admission

The existing tagged byte now works consistently for instance and static object fields: false is 0, true is 1, and undefined is 2. Values outside that domain stop with `invalid boolean field domain`. This value-domain tag is separate from the object's own-property insertion rank and readiness byte. An absent field and an own field containing undefined remain different objects.

Class admission previously refused this representation, and zeroValue reserved a reference slot for it. Both instance and static layouts now reserve the tagged scalar representation before constructor writes. Contextually omitted optional boolean fields reserve an absent slot through optional presence's existing mechanism. Copies preserve the representation, presence, and readiness tails. Unknown reads unpack the scalar tag and box only the observed true/false value.

JSON scalar and literal boolean optionals use the same byte. For object references and direct spreads, this unit admits only bounded const/literal origins whose complete actual fields are boolean or undefined data. It follows the initializer, including hidden fields and spread sources, rather than trusting a structural annotation. Getters, toJSON, unknown origins, and general containers retain their existing boundary. The runtime enumerates actual own keys and ranks, validates physical tags, omits undefined, and preserves reinsertion and replacer order. This also prevents a static JSON field list from resurrecting deleted keys.

Conservative assumption: optional plain-object __proto__ slots are not admitted by the new boolean reservation path because assignment uses JavaScript's inherited setter rather than creating an own boolean field. Classes define their fields through their existing class machinery. No prototype mutation or checked-view admission is added.

No changes were made to internal/native/emit.go, internal/lower/lower.go, internal/native/native.go, or internal/oracle/oracle_test.go. The metadata-selection hook is in internal/native/emit_objects.go's dynamicProperties function. Lower hooks are in class.go's zeroValue, class_inheritance.go, class_static.go, optional_fields.go, and the JSON lowerer.

## Ports, before and after

| Port | Before | After |
| --- | --- | --- |
| Selector gap 3 | quoted boolean plus quotedPresent presence bit; boolean \| undefined declaration refused | quoted directly holds boolean \| undefined; quotedPresent removed; renderer checks undefined |
| Values gap 2 | flag encoded inline/quoted, with shape encoding boolean presence and isHex/isColor initialized false | Separate optional inline, quoted, isHex, isColor fields; boolean key presence comes from values |
| React IR hir-01 | Requested workaround cannot be located in this checkout | Not removed or claimed passing |

The selector and values old refusal cases are removed from their gaps_test.go tables. Their minimized programs move to boolean_optional_selector.a and boolean_optional_values.a in the ordinary oracle inventory. Only existing .ts files were edited; new Adamic sources are .a.

Values still uses shape for nonboolean fields and container kind, and the ports retain their unrelated workarounds. This is removal of their recorded boolean workaround, not a claim that the ports' entire arena representation became the original JavaScript class hierarchy.

The actual port differential tests pass with their default generated corpora: **26,778 selectors** and **8,208 value parses** (5,598 trees and 2,610 refusals). They agree byte for byte with Go cohere, source Node, and the JavaScript backend; native uses ASan, UBSan, and a separate leak run. All existing port mutants also pass their catcher tests. Optional npm library comparisons skip because ADAMIC_SELECTOR_LIBRARY and ADAMIC_VALUES_LIBRARY are unset; no original npm-library run is claimed.

hir-01 belongs to the unlanded React Compiler IR port on hir/land-08 and hir/unit-* branches. Its workaround closes when that port lands; no HIR branch is imported by this unit.

## Node fixtures and pending acceptance

Five ordinary fixtures cover field operations, the two retired minimized gaps, unknown views, and static inheritance. boolean_optional_fields.a covers false and true writes into {}, deletion, spread of present and deleted fields, undefined reinsertion, JSON literal fields, actual boolean object references, direct JSON spreads, insertion order, and replacer order. The standard oracle runs source Node, backend Node, release native, sanitized native, and leak checks.

boolean_optional_view_pending.a observes `true` on Node, then verifies Adamic's NotYet for a checked-view boolean field. TestBooleanOptionalViewPending explicitly skips with **acceptance dependency: compiler/views-rehearsal 2b6c032a**. It is excluded from semantic fixture inventory and counts, never marked passing. The inherited optional-presence checked-view pending fixture also stays skipped.

## Mutants actually run

| Mutant | Catcher |
| --- | --- |
| Remove packed > 2 domain check | TestBooleanOptionalDomainMutant: identical stray byte 255 stops at exit 70 in production and escapes at exit 0 without the guard |
| Ignore optional presence during reads | TestBooleanOptionalPresenceMutant: exit 0 with wrong stdout; after deletion Node prints `undefined|` and `{}`, mutant prints `true|` and `{"flag":true}` |
| Restore old instance-field refusal, selector | Selector's own TestThePortParsesAsGoCohereDoes fails at SelectorNode.quoted with `a field of type boolean \| undefined` |
| Restore old instance-field refusal, values | Values' own TestThePortParsesAsGoCohereDoes fails at ValueNode.inline with the same refusal |
| Restore old static-field refusal | Static fixture's standard oracle fails with `a static field without a native slot` |
| Drop JSON physical-tag guard | TestBooleanJSONMetadataMutant: identical invalid tag 6 stops in production and escapes without the guard |
| Replace complete JSON origin proof with an identifier test | TestBooleanJSONRequiresCompleteDataOrigin fails because an unknown parameter is admitted |
| Drop optional __proto__ boolean boundary | TestBooleanOptionalPrototypeSlotStaysNotYet fails because the setter-shaped source is admitted |

Runtime-domain and metadata tests compile the production helper implementation under private names, changing only their guard. They do not count compiler warnings as kills. Refusal and admission mutants use scratch Go overlays, leaving production files intact. The inherited optional-presence reservation, copy-state, readiness, overlap, and static-key mutants were rerun successfully.

Reproduce the two port refusal overlays by mapping internal/lower/class_inheritance.go to a scratch copy that changes `if slotless(of) && of != ir.MaybeBoolean` back to `if slotless(of)`. Run each port command below with `-overlay /tmp/boolean-optional-refusal-overlay.json`; both must fail. The static overlay makes the same replacement in class_static.go. The JSON origin overlay replaces the depth-zero jsonBooleanObject condition with ast.IsIdentifier(ast.SkipParentheses(node)). The prototype overlay removes only the three-line __proto__ reservation guard.

## Validation commands and observations

Setup: `export GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Reported timing lines: Go ready 0.023s; Node ready 0.024s; submodules ready 0.067s; markdown skipped step-duration 0.007s, ready 0.075s; clang ready 0.169s; go build ready 32.253s; test binaries deferred 32.404s; cache warm 32.406s; done 32.434s. `nproc` is 5, with cgroup cpu.max 400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup log: /tmp/boolean-optional-setup.log.

All test output goes directly to log files. Commands:

```
go build ./...
go vet ./internal/...
go test ./stage1/cohere/selector ./stage1/cohere/values -run 'TestThePortParsesAsGoCohereDoes|TestEachGapStandsWhereGapsMdSaysItDoes' -count=1 -v -timeout 30m
go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 30m
go test ./internal/lower -run 'TestBoolean|TestOptionalDeletion|TestOptionalAssign|TestObject|TestUnknown' -count=1 -timeout 20m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestBooleanOptional|TestBooleanJSONMetadata|TestOptionalField|TestNativeAgreesWithNode/internal/oracle/testdata/(boolean_optional|json_stringify|optional_field)' -count=1 -v -timeout 30m
go test ./stage3/fixtures -count=1 -timeout 30m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m
python3 /tmp/optional-presence-acheck.py
```

The a-check harness uses devtools/fast-gate's aCheck operation, checking only changed/new .a paths against origin/main. Its `checked` result is the source check; it does not turn either pending checked-view fixture into a semantic pass. It includes dependency fixtures because origin/main does not yet contain optional presence.

Mutant commands:

```
go test -overlay /tmp/boolean-optional-refusal-overlay.json ./stage1/cohere/selector -run 'TestThePortParsesAsGoCohereDoes/natively_and_on_Node' -count=1 -v -timeout 30m
go test -overlay /tmp/boolean-optional-refusal-overlay.json ./stage1/cohere/values -run 'TestThePortParsesAsGoCohereDoes/natively_and_on_Node' -count=1 -v -timeout 30m
go test -overlay /tmp/boolean-optional-static-refusal-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/boolean_optional_static' -count=1 -v -timeout 20m
go test -overlay /tmp/boolean-optional-json-origin-overlay.json ./internal/lower -run TestBooleanJSONRequiresCompleteDataOrigin -count=1
go test -overlay /tmp/boolean-optional-prototype-overlay.json ./internal/lower -run TestBooleanOptionalPrototypeSlotStaysNotYet -count=1
```

Each exits 1 at the intended test assertion. Logs use /tmp/boolean-optional- with suffixes ports.log, gaps.log, oracle-final.log, lower-final.log, build-final.log, vet-final.log, stage3-final.log, counts-update-final.log, counts-final.log, acheck-final.log, selector-refusal-mutant.log, values-refusal-mutant.log, static-refusal-mutant.log, json-origin-mutant.log, prototype-mutant.log, and json-metadata-mutant.log.

Counts are regenerated on Linux: five new rows relative to a774d316, with no existing-row changes. The pending fixture has no row.

There are **no Compiles-to-Refused or Compiles-to-NotYet regressions**. Only two new stage3 records change: taste/11_build_info_pending.a from NotYet to Compiles, and objects/06_watch_close.a from NotYet to the existing cycle-capable Refused diagnostic. The update harness regenerates the compiling fixture only after its native and Node observations agree. Its updater cannot write noncompiling outcomes because nativeAgrees stays false; the watch-close diagnostic was recorded separately from the observed compiler diagnostic after the fixture's Node check passed. A byte audit confirms all bytes outside stage0, including all Node observations, are unchanged. Log: /tmp/boolean-optional-stage3-byte-audit.log.

No full gate, performance measurement, Darwin, WASI, original npm-library run, or unavailable hir-01 differential test is claimed. General JSON objects/accessors/toJSON, mixed boolean unions, and checked-view optional writes remain outside this unit.

## Landing on c2

The authorized merge of cloud/land-stack-c2-c79c7572 c79c7572 carries a774d316 as an ancestor. The delivery branch lacked c2, so c2 was merged without conflicts. Its packed slot cache requires deriving the spread field presence index from the actual returned slot; emit_objects.go and reuse.go now do so. Copy-state and insertion-order mutant probes use that same API, including c2's shape registration initializer. Pending view assertions track the current optional/conversion refusal and remain skipped.

The landing checks use /tmp/boolean-optional-c2- logs: build-green, vet-green, acheck-final, stage3-green, oracle-complete, counts-green, and the selector, values, static, JSON-origin and prototype mutant logs. The a-check operation covers the six .a files changed against c2. The oracle reruns all Boolean and inherited optional-presence fixtures and mutants uncached, including release and sanitizer builds. Its final native fixture filter additionally includes reuse, spread_undefined, and class_instance_key_consume to exercise the reused-spread hook.

No Compiles regressions occur. Two c2 stage3 real fixtures, structural-method-statics/main.a and generic-optional-return/main.a, change from NotYet to Compiles after the updater verifies recorded Node, current Node and native agreement. All Node observations and all bytes outside stage0 remain byte for byte unchanged. HIR remains with its owning unlanded port.
