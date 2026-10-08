compiler/stage3-front-3 evidence checkpoint: b7e19a97972a40ed496dab9650c3c78f99d9281d
Combined stage3 compiler merges from exact main base 48c05d091f0a43c31cbe051b1d6578d99eeedf19.
Each landed item passed uncached lower/native/ir/flow/oracle/stage3 fixtures, vet, gofmt and whitespace checks before push.
Mutants are checked by Node output/exit comparisons, pinned refusal/IR assertions and ASan/UBSan/LSan; evidence is listed below.
Held/skipped items are excluded; full TypeScript compiler acceptance and performance parity are not claimed.

| Item | Branch | Exact pin | Result | Branch SHA after item | Failure / resolution |
| --- | --- | --- | --- | --- | --- |
| 01 | codex/nested-functions-scanner | 8b7b8db9af2a2c6a0458974e8e9618221d764d27 | merged | ad44eafd92010b101aea6f5a49c1e35c042965c9 |  |
| 02 | codex/scanner-nested-references | e35ee2d390e87e22369b4cb8c198aec057d863f3 | merged | 2db01c2fe919a4cad282b7d6fe59bf8c000c8cb2 |  |
| 03 | codex/imported-const-case | c9d885d79e10c6e0b404a9b6e9f3529f9e9b3412 | fixed | 28a101229a0cdd8812e5de15494392097673368d | Initial counts: five case-function prepass NotYet failures and two cyclic-star-export refusals; resolved on branch. |
| 04 | codex/host-blockers | cd220dd58420c25c884e92ef37d789973126ebce | held | 28a101229a0cdd8812e5de15494392097673368d | Held by user: 32 conflicts, ABI/narrowing/break design reconciliation pending. Counts initially failed shorthand panic and case restriction; stage3 native panics in objects/08_loop_state, objects/12_decorator_descriptor, taste/11_build_info_pending. |
| 04a | codex/host-blockers | b788e96e4e4d0a3f211b6dd1cc28504b5d3935b1 | held | 28a101229a0cdd8812e5de15494392097673368d | Held by user pending closure-convention replacement. |
| 05 | codex/scanner-generic-optional-result | 6ffb8bb89c020d81d800cb139b2c5768cc552ff6 | merged | 1bf3f59df068a60df71c94c76e66cd6634c25d38 |  |
| 06 | codex/scanner-expressions | 998fb3eb3fc89fbeaebc5453d328ed9f2f222fb7 | merged | ac2836a16ad858617fdc76999bd03c0c4a2e4dd5 |  |
| 07 | codex/scanner-string-destructuring | cf9fa276c9180f63c78e9697760599aba919832f | merged | 45c2e05235e1c40fa0ed7bff748ba52e7cc3e268 |  |
| 08 | codex/scanner-value-increments | 85cade98e3bd0bee1f7b10f2b20c2c905d05f4f2 | merged | 022fdaa60e018b91ac6478ffdef3fc6d685bb4e9 |  |
| 09 | codex/scanner-generic-property | 75b107a0a486844a537255f6d58a354776b6c0d6 | fixed | 34d3c38456bc27dece97faca061b1c33eb497600 | Counts initially failed nested-generic-sibling-call.a:5:29, nested generic call without concrete type arguments; preserve existing registered specialization. |
| 10 | codex/non-null-narrowed-number | 3a6c8231e8ecf8bd695fbf1df80003f2348ab5f5 | fixed | 8051eedc9160e8aa6b5f43463139bd260804c085 | First full gate: stale non-null policy expectations in TestOmittedOriginalProbePolicy and two TestRealNestedBlockers cases. Update remaining diagnostic expectations; compose declaration/readiness semantics. |
| 11 | codex/parser-generic-empty-array | 4e022aaef7d24085fd0ded0c11b539843022fc47 | merged | 21243ee4b7299cc17ede9dcd763ba2b2f3a15c0e |  |
| 12 | codex/census-small-families-3 | 24a2b28709aeaa80615e9b2760e534829b018b80 | fixed | e0b679cb38eadfc7e19d8989127c52d535b27514 | Initial gate: stale nested rest/checker diagnostics, lost unsupported union refusal, and inherited checked static field panic. Preserve refusal and checked boxed views; metadata index derives from actual slot. |
| 12a | codex/proven-predicates | a58ba402b428280d4932afe6c98c1f47a5d3c495 | fixed | b144826210541e4c5e1b8e0c51f2f8b1e8216d79 | Compose early context restoration with nested-frame specialization; no gate regressions. |
| 12b | codex/proven-predicates-2 | b131a36dc50163896dbbdd0ef6d304223454fd0a | fixed | 3b7e996f80c87ec8b67b7c7328c6465ca6f13a44 | Retain closed-caller refusal for uncalled predicate callback; incoming test now pins it and keeps original Node source observation. |
| 13 | codex/views-arrays-callables-parser | c898009bfb8520b03c5a366ac23c308e1e56c882 | skipped | 3b7e996f80c87ec8b67b7c7328c6465ca6f13a44 | TestCountsAreRecorded: 19 carried filesystem fixtures refuse Error to ErrnoException optional errno widening, e.g. node_fs_file_rm.a:5:42. Preserve predicates-2 optional-field proof; resolution saved in /tmp/stage3-front-3-13-skipped-resolution.patch. |
| 14 | codex/records-lowering | 456c981b1c108abae5f2c8d6ae665cfd6b92e2fe | merged | 0059e65c1a3674692fe905c03abf9c2159c5c017 |  |
| 15 | codex/namespaces-tsc | 47a6fabec9eebd21f1e6320e7668216f18729680 | fixed | 567564331b7b1bb520bf872d312694508caefabf | Initial counts: parameter-property AST-kind panic and namespace repeated assertion-var refusal; fixed with kind guard and namespace-only exemption. |
| 04-replacement | codex/closure-convention | 22fd701a3a4dee5c11cffec3d6b750cfe28ce399 | skipped | 567564331b7b1bb520bf872d312694508caefabf | TestCountsAreRecorded fails optional errno/encoding/mode widening in carried filesystem fixtures, Buffer representation, host24 namespace reachability and taste21 checked objectType. Non-null resolution and proving fixtures saved in /tmp/stage3-front-3-04-replacement-skipped-resolution.patch. |
| 15a | codex/enum-init-reach | 2152fc3b5c74cd5455887639184f88608f1f6060 | fixed | 282874ebcf1d11a19daa78f06a53ef2be37095a2 | First full gate: three premature-read fixtures use incompatible inserted-check registration (native70, Node70). Keep dedicated source Node/exact backend stop pins, register seven success fixtures only; corrected full gate passes. |
| 16 | codex/phantom-brands | d90994daa68bdba300f22acdad849a22ad5192a2 | fixed | 58270ec655f020bc7e3d20e37469fc3b942af56b | Rest overload parameter comparison, stale diagnostics, and guarded call-target read repaired; first gate lower diagnostic and TestCallTargetReaders failures. |
| 17 | codex/library-array-holes | f05aec3cee91d2ce3bfced6d341865933ec28da1 | held | 7a1fd0a95cee3059533fa7a7f5adbe9a7787c014 | Counts94.129s fail22 carried host fixtures,17 optional-errno refusals and5 namespace-readiness stops. Exact fixtures and incoming recorded counts to unmeasurable outcomes in stage3/front-3/item17/REPORT.md; candidate patch committed as artifact only. |
| 18 | codex/library-error-value | 6469f37bc02f5ce41b5735a966e68e974377c388 | skipped | b7e19a97972a40ed496dab9650c3c78f99d9281d | STOP: unmodified candidate Error.captureStackTrace(frozen target), Node caught/exit0; native and JavaScript completed/exit0. Candidate patch and exact probe preserved in stage3/front-3/item18. |
| 19 | codex/module-init-order | 28e366fa9ffa8ad5a44e2ee8a0d97a38faf2c916 | pending | not attempted | Stopped at item18 exit-zero disagreement. |
| 20 | codex/method-presence-test | 16cb9b105127f28eb5a6d1d0af7757c6083e2d27 | pending | not attempted | Stopped at item18 exit-zero disagreement. |
| 21 | codex/enum-tag-narrowing | 41231d514cf0090d1a7ec103326858b7eb34b6e3 | held | b7e19a97972a40ed496dab9650c3c78f99d9281d | User hold: carries views-integration e555d67e; optional-boolean storage clearance not received. |
| 22 | codex/debugger-statement | 7d2cbc895a1e690c18ceefa908da44618370b769 | pending | not attempted | Stopped at item18 exit-zero disagreement. |
| landing | origin/main | not freshly fetched | pending | not attempted | Stopped before current-main landing. |

No landed Compiles outcome became Refused or NotYet. Existing status files have all bytes outside stage0 verified against pre-merge HEAD, including Node observations. Counts are generated by the Linux oracle; none are hand-written.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. nproc=5, CPU quota=4. Node24.19.0, Go1.27.1, clang20.1.8. Setup timing: Node0.073s, Go0.081s, clang0.629s, markdown1.151s, submodules213.199s, Go build456.773s, deferred test binaries456.892s, cache456.894s, total456.998s.

| Item | Gate package seconds (lower, native, ir, flow, oracle, fixtures) | Evidence prefix |
| --- | --- | --- |
| 01 | 87.067s, 469.633s, 16.730s, 256.960s, 459.939s, 59.723s | /tmp/stage3-front-3-01 |
| 02 | 80.739s, 464.352s, 17.872s, 278.811s, 456.332s, 55.598s | /tmp/stage3-front-3-02 |
| 03 | 83.919s, 423.175s, 31.549s, 324.261s, 431.337s, 56.225s | /tmp/stage3-front-3-03 |
| 05 | 90.264s, 479.507s, 15.644s, 321.697s, 467.262s, 71.611s | /tmp/stage3-front-3-05 |
| 06 | 82.412s, 423.739s, 14.202s, 300.481s, 465.305s, 54.897s | /tmp/stage3-front-3-06 |
| 07 | 81.955s, 449.774s, 14.680s, 305.725s, 461.510s, 60.348s | /tmp/stage3-front-3-07 |
| 08 | 89.848s, 464.995s, 16.033s, 314.538s, 482.453s, 63.450s | /tmp/stage3-front-3-08 |
| 09 | 100.492s, 552.959s, 21.448s, 328.492s, 562.824s, 70.276s | /tmp/stage3-front-3-09 |
| 10 | 111.929s, 638.778s, 7.486s, 373.712s, 687.707s, 71.501s | /tmp/stage3-front-3-10 |
| 11 | 97.660s, 533.853s, 2.560s, 318.043s, 588.218s, 60.236s | /tmp/stage3-front-3-11 |
| 12 | 110.501s, 504.699s, 14.771s, 304.090s, 586.589s, 60.613s | /tmp/stage3-front-3-12 |
| 12a | 126.386s, 462.157s, 17.381s, 363.148s, 626.772s, 70.887s | /tmp/stage3-front-3-12a |
| 12b | 129.084s, 534.445s, 3.554s, 359.464s, 636.297s, 72.962s | /tmp/stage3-front-3-12b |
| 14 | 151.065s, 579.481s, 21.686s, 421.040s, 707.817s, 90.988s | /tmp/stage3-front-3-14 |
| 15 | 155.995s, 612.967s, 35.229s, 468.017s, 775.226s, 107.788s | /tmp/stage3-front-3-15 |
| 15a | 141.897s, 491.005s, 2.749s, 359.607s, 637.503s, 66.294s | /tmp/stage3-front-3-15a |
| 16 | 154.813s, 516.288s, 16.502s, 393.560s, 658.716s, 67.867s | /tmp/stage3-front-3-16 |
| 18 |  | /tmp/stage3-front-3-18 |


Closure candidate: 22fd701a3a4dee5c11cffec3d6b750cfe28ce399 carries the requested host, census and arguments-length pins and neither held e555d67e nor skipped c898009b. The 28 conflicts are resolved in /tmp/stage3-front-3-04-replacement-skipped-resolution.patch. It retains representation-changing Narrow and all checked Call nodes; only same-representation references may be unwrapped. Added Box|null fixture and five direct IR cases pass. Two direct narrowing overlays and checked-call-erasure are caught; the original direct-Narrow mutation survived because source fixtures used checked helper calls, so no proof is claimed for that attempt. The full candidate is not green: /tmp/stage3-front-3-04-replacement-counts.log records filesystem optional errno/encoding/mode refusals, Buffer representation gaps, host24 namespace reachability and taste21 checked objectType. Existing refusal and Node contracts are preserved by leaving the candidate out.

Parser-views c898009b was left out after carried filesystem Error to ErrnoException contracts failed the existing optional-field proof. Candidate patch: /tmp/stage3-front-3-13-skipped-resolution.patch. Its candidate-only mutants are not claimed as landed coverage.

The top finding is item18: the unmodified candidate prints completed with exit0 for frozen-target Error.captureStackTrace, while Node prints caught with exit0. Both native and JavaScript IR disagree. The unit was stopped as instructed. Intentional wrong-output mutants are identified separately as proving controls. Build failures, no-test matches, interrupted attempts and survived controls are excluded from successful mutant evidence.

The compiler source at this checkpoint is identical to last green integration 58270ec655f020bc7e3d20e37469fc3b942af56b; preservation and report commits add artifacts only. The final branch SHA, including the documentation commit, is supplied in the final response. The working tree is cleared and only compiler/stage3-front-3 is pushed.

Item17 preservation commit: 7a1fd0a95cee3059533fa7a7f5adbe9a7787c014. Its full saved resolution patch is stage3/front-3/item17/resolution.patch. The exact 22 fixtures, six incoming recorded counts each, and failed candidate outcomes are listed in item17/REPORT.md. These were lowering failures, so no candidate numerical counts were measured; unavailable is not zero. This distinguishes the supplied incoming baseline from an invented before/after count delta.

Item18 evidence commit: b7e19a97972a40ed496dab9650c3c78f99d9281d. item18/REPORT.md explains the wrong-output reproduction, preserved candidate fixes, actual partial gates, disk/cache failure and interrupted rerun. Its compiler changes are excluded. Node stdout was caught\n, native and JavaScript stdout completed\n; all three exits were0, all stderr empty. The Go overlay changed only test code. The direct source is item18/capture-frozen.a; the raw result is item18/capture-frozen.log.

Items19,20,22 and current-main landing remain pending because the explicit exit-zero stopping condition fired. Item21 remains held for views-integration clearance. Full TypeScript parser/scanner/compiler acceptance, performance parity, macOS and a fresh Test262 sweep were not run by this integration unit.
