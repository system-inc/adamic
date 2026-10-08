Merged lazy checked views and lowered enum-tagged structured, optional and undefined payloads through their read contracts.
Implementation: branch codex/enum-tag-narrowing from b2be5147, first merging views-integration d1937df8 in 2fb22170, then newest e555d67e in a9bfbc6f and main f4efdd23 in a36d2960.
Verification on the final merge: uncached enum/checked-view oracle PASS 76.009s; twelve fixture counts PASS 1.682s; enum and optional-view lowering pins PASS 1.598s; paired census 90 to 0 in each input, 79/79 files, zero errors or panics.
Mutants: transitive read erasure, undefined admission in both backends, callable/binding name guards, packed-false admission, nullable receiver bypass, and existing open-tag/remainder mutants are caught.
Not covered: full compiler compilation or a green repository gate; merged upstream has reproduced lower/native failures and the full counts update fails.

## Payload census before implementation

The accepted b2be5147 latent census has 90 unique sites in each input for
`checked enum object views with incompatible structured or optional payload fields`.
These are checker-rejected programs, measured with output disabled. Removing a
row does not establish that those 90 programs compile or execute.

[The shape ledger](enum-tag-views-shapes.json) records all 90 main locations,
checker payload types and optionality. Area has the same field histogram.

| Payload family | Sites | Details |
| --- | ---: | --- |
| undefined | 39 | 19 required and 20 optional; largest identical payload type |
| structured object | 40 | 11 distinct checker shapes, 38 required and 2 optional |
| NodeArray | 5 | Four Statement arrays, one Expression array |
| branded __String | 3 | Existing view family handles representation |
| optional boolean | 1 | Packed native optional storage is decoded explicitly |
| never | 2 | Lazy admission; named refusal if read |

The structured family is larger in aggregate; it is not one identical type.
Implementation first covers undefined, then uses the merged structured and
optional families. No cohere source was copied.

## Lowering and runtime behavior

Narrowing no longer rejects an entire target because one payload field lacks a
view family. For structured, optional or nullish payloads, enumTagViewAs uses the
shared lazy view. Origin schemas survive nested reads, aliases and helper calls.
Unused fields are admitted. Unsupported families become Refused at the read,
including never and unsupported nested array consumers; pins check the family
and consumer location. Existing eager scalar enum payload checks retain their
messages and old mutant coverage.

Pure undefined fields have an explicit view contract. Native validation checks
actual storage before reading it, distinguishing undefined from null, false and
number values. Packed optional booleans use their byte representation rather
than pointer reinterpretation. JavaScript performs the corresponding contract
check. Once a receiver is narrowed to an object, discriminant detection excludes
nullish union variants and rechecks that the remaining type is a union.

Twelve .a fixtures run under Node for source behavior, sanitized native,
release native and the JavaScript backend. Bad views are compared with a pinned
exit 70 and exact panic text rather than Node's erased cast. Successful witnesses
match Node and pass leak checks. Wrong nested boolean, required/optional undefined,
null, false, optional boolean and nullable receiver cases stop at their reads.
The unread unsupported-payload fixture finishes successfully.

The inventory also exposed a BindingPattern panic in callable implementation
scanning and a ComputedPropertyName panic in binding scanning. The first now
looks only at actual callable implementation names; the second becomes a named
refusal. Reduced direct schema/binding pins and restored-guard mutants cover both.

## Mutants and verification

| Mutant | Independent failure |
| --- | --- |
| Existing open_tag eager check erasure | TestEnumTagPayloadMutant |
| Existing open remainder erased to never | TestEnumTagOpenRemainderMutant |
| Nested ready-field view removed | Both backends finish incorrectly; structured-wrong pinned stop catches it |
| Native undefined guard accepts any value | Required and optional wrong-value exit assertions fail |
| JavaScript undefined guard accepts any value | Required and optional wrong-value exit assertions fail |
| Callable inventory uses unguarded Name.Text | BindingPattern panic in direct schema pin |
| Computed binding guard removed | ComputedPropertyName panic in named-read pin |
| Packed false treated as an undefined reference | undefined-false finishes incorrectly; pinned stop fails |
| Nullish receiver normalization removed | nullable-structured-wrong exits 0 instead of 70 |

All temporary mutants were restored. Logs are under /tmp/enum-views-*:
current-oracle.log, current-lower.log, current-counts.log, native-undefined-mutant.log,
js-undefined-mutant.log, callable-name-mutant.log, binding-name-mutant.log,
packed-false-mutant.log and nullable-mutant.log. The transitive mutant is permanent
in the fixture oracle. Twelve allocation rows were measured with counted via a
scratch Go overlay and recorded in internal/oracle/counts.md.

Additional checks: lower excluding the seven reproduced upstream failing roots
PASS 97.850s at the first checkpoint and 49.467s on the final merge; JavaScript and IR packages PASS 2.048s and 30.932s; native Field|View|Object
PASS 40.868s; touched package vet passes. These initial broader checks preceded the final
nullable adjustment. Final merged native Field|View|Object passes 9.075s, full
JavaScript 0.985s and full IR 18.155s; final vet passes. The final enum
lowering/oracle pins cover the nullable adjustment.

The full touched-package run has seven lower failing roots and five native graph
failing roots. The exact failures reproduce in an unmodified d1937df8 snapshot.
The full counts update fails 45 fixture subtests and does not rewrite its table;
its unmodified upstream comparison aborts in the callable-name panic. Therefore
not all 45 count failures are independently established as inherited. The full
repository gate was not run. No pass-to-fail claim is made for the whole meter.

Setup: Node 0.025s, Go 0.027s, submodules 0.072s, markdown 0.083s, clang 0.159s,
Go build 24.048s, deferred test binaries 24.201s, warm 24.202s, total 24.232s.
nproc=5; CPU quota 4. GOPROXY=https://proxy.golang.org|direct; environment sourced
from /workspace/adamic-tools/env.sh. Node declarations installed from the pinned
stage3/api lockfile. Test output went to log files.

## Scope

Manual integration/implementation edits (the incoming merge imports additional files):
internal/ir/ir.go;
internal/javascript/javascript.go, readiness.go;
internal/lower/cast.go, cast_proof.go, invariance.go, expression.go, enums.go,
enum_tag_views.go, interface_cast.go, object.go, view_callables.go, view_arrays.go;
internal/native/emit_expressions.go, emit_objects.go, view_fields.go, runtime/object.c.
New pins: internal/lower/enums_views_test.go, internal/oracle/enums_views_test.go,
and twelve stage3/fixtures/enum-views/*.a files. Records: internal/oracle/counts.md,
docs/enums.md and this report with its shape/meter artifacts.
The final main merge preserves the views branch versions of internal/lower/optional_widening_census_test.go, internal/lower/optional_widening_test.go, internal/oracle/counts.md and internal/oracle/testdata/field_access_paths.a to retain lazy admission and its census evidence. These are merge resolutions, not new widening implementation.
The four prohibited files were not manually edited; their incoming integration
changes are part of the requested merge. cast_proof.go and invariance.go conflicts
were comment integration only. No pull request is opened.

## Meter reproduction

Inputs are the pinned adapted main and area trees used by the preceding unit,
79 files each. Exact input revisions and hashes are in enum-census-panics.json.
Use stage3/meter/report.py latent_summary to validate all file coverage and count
unique (kind, where, reason, text) sites.

The merged census overlay generator rewrites return true in the predicate
preflight as though it belonged to the main refusal visitor, producing undefined
found/visit symbols. The scratch generator preserves the preflight and applies
that rewrite only after `var found error`. It makes no production loader or
output changes. The correction is archived alongside this report; build with
its generated overlay and run stage3/census/latent/tool on each src/compiler tree.

Commands for the pinned final compiler:

```bash
source /workspace/adamic-tools/env.sh
python3 docs/verification/enum-tag-views-overlay.py "$PWD" /tmp/enum-views-landing-overlay
go build -overlay /tmp/enum-views-landing-overlay/overlay.json -o /tmp/enum-views-landing-meter ./stage3/census/latent/tool
/tmp/enum-views-landing-meter /tmp/adamic-gate/stage3-meter.RLIiIU/main-adapted/src/compiler /tmp/enum-views-landing-run/main/latent.jsonl
/tmp/enum-views-landing-meter /tmp/adamic-gate/stage3-meter.RLIiIU/area-adapted/src/compiler /tmp/enum-views-landing-run/area/latent.jsonl
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestEnumTag|TestCheckedView' -count=1
go test ./internal/lower -run 'TestEnumTag|TestOptionalWidening' -count=1
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/oracle
```

Every run above redirects stdout/stderr to the corresponding /tmp/enum-views-landing-*.log in the actual execution. Binary measurement is equivalent to go run of the overlaid tool; no normal loader is relaxed. Final census comes after the newest views/main merges. Earlier incomplete runs are excluded.

## Final census and push

[The complete meter ledger](enum-tag-views-meter.json) records the final compiler,
views/main tips, input hashes, binary hash, unique sites and every reason row.
The display row groups raw reasons by the prefix
`a checked numeric enum object view with an incompatible structured or optional payload field `;
the final word is the field name.

| Input | Payload NotYet before | After | Files | NotYet total after | Refused total after | Skipped dependencies | Errors | Panics |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| main | 90 | 0 | 79/79 | 1050 | 708 | 4 | 0 | 0 |
| area | 90 | 0 | 79/79 | 1052 | 706 | 4 | 0 | 0 |

Coverage and row assertions pass via stage3/meter/report.py latent_summary.
Other totals include the effects of all merged view families and are not
attributed solely to enum payload lowering. This is one push after implementation,
latest views/main integration and the final report: the row is 0 on both inputs
at that push. The implementation merge is 2fb22170, newest views merge a9bfbc6f,
and main merge a36d2960; the report commit is the pushed branch tip.
