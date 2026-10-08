Built primitive phantom brands, erased casts, catchable undefined reads, and primitive-property collision refusals.
Commits: current-main merge `0f6c5f1`; earlier review `03d6b83`; implementation SHA is reported with the feature-branch push.
Commands: touched packages and all oracle tests passed; the audited 78-root census measured brand value refusals 66 before, 45 after.
Mutants: undefined-yields, prototype-accepted, nonvoid-accepted, runtime-cast, and object-representation all failed their intended checks.
Not covered: the complete `./...` gate; standalone Node's uncaught stack/exit-1 protocol; bigint/symbol brands and two-argument substr.

No commit was pushed to main or an area branch, and no merge into either was performed. Only `codex/phantom-brands` is the push destination. `origin/main` was fetched explicitly again and merged before the final report; it remained `b6b1538b0cebc4ba6741ac34f1aedb60293c1d06` and the merge reported "Already up to date."

## Implemented ruling

`internal/lower/phantom_brands.go` holds the rule and its small hooks. String, number and boolean brands retain their primitive representation. A void or undefined brand uses the canonical undefined representation, with no object allocation. A union such as tsc's `__String` retains its primitive arms and undefined arm. Casts between proven compatible primitive representations introduce no IR operation. Literal constraints remain enforced.

A completed brand-member read returns undefined. Ordinary reads through undefined create a TypeError and unwind through the existing exception mechanism; optional reads return undefined. Receivers run once, including their side effects. The stale-narrowing fixture verifies that the ordinary read remains catchable after a function changes a narrowed value to undefined.

Names on the actual primitive's own properties or complete prototype chain are refused, including inherited Object names and canonical string indices. The inventories are checked independently against Node's boxed primitives. Boolean `length` and number `slice` remain valid absent names. Non-void members remain refused, even in unused declarations. The refusal pins the member and the fix in tests. The original `__proto__` review program is refused with:

```
main.a:1:14: Adamic 0.1 refuses a primitive brand member __proto__ that exists on the primitive; use a member name the primitive does not have on its own properties or prototype chain
```

The actual tsc Path declaration remains refused with:

```
main.a:2:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or optional and typed undefined) so the brand is phantom
```

The fixture preserves tsc 6.0.3 `utilitiesPublic.ts:842-854`'s escape/unescape bodies. The original unescape body needs `substr(1)`, so the string-library hook maps zero/one-argument substr to slice, whose relative-start semantics are identical. Calls supplying a length stay refused.

## Oracles and limits

All new sources are `.a`. `phantom_brands.a` exercises string operations, Map key equality, cast in/out, and the three-arm `__String`, including InternalSymbolName and an undefined value. `phantom_catch.a` exercises catchability, optional reads, numeric/boolean brands, receiver evaluation, and stale narrowing. `phantom_undefined.a` preserves the review read with String only to satisfy Adamic's one-string console signature.

`TestPhantomUndefinedReview` also executes the original, unchanged `review/phantom-brands/undefined-read.a` as the Node source oracle. A load overlay wraps only the console argument in String for Adamic: the property read throws before that conversion can run. Source Node, JavaScript backend, sanitized native and release native agree byte for byte on empty stdout, the TypeError message and exit 70 under the repository oracle runtime.

Standalone Node has a different uncaught-error protocol: exit 1 and a source/stack trace. `oracle/adamic.mjs` normalizes uncaught Node errors to Adamic's single-line panic and exit 70. This unit preserves that existing protocol; it does not claim standalone uncaught stack/exit-1 parity. The catchable-error fixture finishes with exit 0 and prints the same TypeError name/message as the source on Node.

## Census observations

The README and method of the pinned `70456b7` census were read before measurement. Findings are measured on a checker-rejected program, deduplicated by `(kind, where, reason, text)`. Each attempted unit reports only its first lowering error; these numbers are observations of blockers, not successful compilation of tsc.

The archived cumulative report's **222** is the sum of every exact `NotYet: a value of type ...` reason. It includes non-brand types such as `any`, type parameters and AST object intersections. Filtering its exact primitive-brand value reasons yields **77**, not 222. This distinction is reproducible from the archived report and the exact reason list in `evidence/census-summary.json`.

The same tooling was rerun with current main before this production change and this feature implementation as the after configuration. Scratch overlays and binaries alone admitted checker-rejected source; production loading and lowering were untouched by the measurement. The adapted TypeScript 6.0.3 source uses the pinned pipeline's adaptations 10 and 20. All **78** source SHA-256 hashes equal the original archived manifest. Both configurations produce **2,165** checker diagnostics. Current main contains subsequent integrations, so its starting counts differ from the archived cumulative configuration.

| Exact diagnostic family | Archived cumulative | Current-main before | Feature after |
| --- | ---: | ---: | ---: |
| Broad archived named-value reason filter | 222 | 196 | 175 |
| Primitive-brand value reason filter | 77 | 66 | 45 |
| `__String` value, intersection and expanded union | 30 | 21 | 0 |
| `a function returning __String` | 10 | 10 | 0 |
| `a function returning __String \| undefined` | 4 | 4 | 0 |
| Path | 20 | 19 | 19 |
| ResolvedConfigFileName, including unions | 9 | 8 | 8 |
| ResolvedConfigFilePath | 18 | 18 | 18 |

Thus this family removes **35 __String representation refusals**. The 45 non-void Path/config value refusals are unchanged. There are also eight distinct new declaration refusals identifying non-void members, including tsc's any-typed Path and never-typed resolved-config brands. The new prototype-name restriction produces **zero** collision refusals on this corpus. Total unique NotYet/Refused/Panic sites change from 3,742 to 3,731 as later blockers become visible. All named-value reasons, including a newly visible reason outside the archived filter, total 196 before and 176 after.

The census configurations, binary/overlay hashes, complete compressed raw ledgers, exact reason filters, family locations/messages, and matching source manifest are committed under `evidence/`. The original historical scratch conflict resolver failed on duplicate braces while preparing its cumulative tree; that scratch-only conflict was repaired. The headline rerun uses current main, as the unit requests, rather than presenting those historical reconstruction attempts as measurements.

Measurement commands, all with output redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/phantom-after-overlay
# gofmt the generated scratch overlay files
# The before binary was built from main's production source before any edits.
go build -buildvcs=false -overlay=/tmp/phantom-after-overlay/overlay.json -o /tmp/phantom-after-census ./stage3/census/latent/tool
LATENT_ASSERT_NO_OUTPUT=1 /tmp/phantom-before-census /tmp/phantom-tsc/src/compiler /tmp/phantom-before-exact.jsonl
LATENT_ASSERT_NO_OUTPUT=1 /tmp/phantom-after-census /tmp/phantom-tsc/src/compiler /tmp/phantom-after-pinned.jsonl
python3 stage3/census/latent/audit.py /tmp/phantom-after-census
```

The census audit passed continuation, all refusal sites, no IR output, disabled production loader, body skipping, signature eligibility and same-line ranges. Its planted-function, signature/body-range and misattribution mutants were caught. Earlier runs with incompletely adapted inputs were discarded.

## Validation and mutants

`bash cloud/setup.sh` succeeded: Go 1.27.1 0s, clang 20.1.8 0s, Node 24.19.0 0s, submodules 0s, build cache 114s, total 114s. `nproc` prints 5; cgroup cpu.max is `400000 100000`. The environment used `/workspace/adamic-tools/env.sh`.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/flow ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle -count=1
```

Passed: flow 173.461s, ir 16.667s, lower 51.000s, native 279.686s, oracle 265.705s; JavaScript has no separate package tests. The full oracle package includes Node, both backends, ASan/UBSan, release builds and leak checks. `TestCountsAreRecorded -args -update-counts` passed in 56.916s and added precisely the three fixture rows. The final uncached focused run passed oracle 0.898s and lower 0.408s; the added real-Path pinned-message test also passed (lower 0.249s). After the final merge check, the focused run passed lower 0.503s and oracle 0.903s. No full `./...` gate was run.

| Production mutant run and restored | Check that caught it |
| --- | --- |
| Skip the native undefined receiver check | Original-review oracle: exit 70 became 0; catch fixture: TypeError output became undefined. Sanitized and release failures recorded. |
| Skip the primitive-name refusal | Pinned __proto__ review diagnostic and unused prototype-brand declaration both fail. |
| Skip the non-void member refusal | The non-void brand declaration is accepted; the pinned refusal test fails. |
| Return ir.Defined from a phantom cast | Full IR equality against the unbranded program fails; runtime cast operation appears in both functions. |
| Represent a phantom string as ir.Object | The same IR test detects object return/parameter representations in place of string. |

These mutants failed for the intended behavioral/IR assertions, not compiler warnings. Sources were restored, and the final focused run was green. Logs are in `evidence/`; no mutant is present in production source.
