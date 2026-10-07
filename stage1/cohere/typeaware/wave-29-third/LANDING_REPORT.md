Built: wave 29 rebased onto current main and rechecked; local RegExp gap assertion now checks the working feature against Node.
Commits: old tip 5a73f6b7, main e8ba3d5d, rebased source tip 8f790c11; final landing-evidence commit is reported in the handoff.
Commands: setup 134s, nproc 5; completed six-rule comparisons, sanitizers, handles, inherited coverage and filtered Node oracle pass again.
Mutants: all six rule faults, raw provenance/rune/shorthand, VM context, retained registries, two kernel faults and inherited guards are caught.
Not covered: the three full React ports still lack native SSA lowering/capture/control analysis; full repository gate and full emitted-JavaScript lint comparison were not run.

The only branch pushed by this worker is `codex/typeaware-wave-29`. It was not
an ancestor of main. `git rebase origin/main` replayed 14 commits cleanly onto
`e8ba3d5d81de4d3773c723914fccd4c76248b965`; source tip afterward was
`8f790c11d7ee98b5a24f1e9fe5fead57a84bf1c6`. The inherited ten-rule baseline
commit remains included. A remote check immediately before the handoff still
shows that same main tip and old branch tip
`5a73f6b73ee7ada5c5d024177b9231c09a3ed6ce`. The push uses an explicit lease on
that old branch tip, targets only this worker's branch, and preserves concurrent
remote updates by refusing to overwrite an unexpected tip. No main or area
branch is a push target. No new rule claim was taken.

All tests used rebuilt stage 0 and rebuilt bridge archives from this rebased
source. Tools remain Go 1.27.1, clang 20.1.8, Node 24.19.0. Fresh setup completed
in 134 seconds: Go/clang/Node/submodules ready 0s and cache warm 134s. nproc is 5,
CPU quota is 400000/100000, and memory is 17.6 GB. Independent suites ran
concurrently; timings below are observations under that load, not benchmarks.

| Check | Observed result |
| --- | --- |
| Configured id-denylist/id-match | 272 inputs, 209 findings, 70068 equal bytes; instrumented native agrees |
| Default original batch | 41 controls, 15 findings, 9881 equal bytes; compiler 77 roots/5241 bytes, repository 287 roots/18485 bytes; both zero findings; instrumented native agrees; Go test PASS 161.700s |
| Independent regexp VM | 1050 decisions equal to Go regexp; instrumented native agrees |
| Second batch | 400 controls, 252 findings, 120087 equal bytes; same 77 compiler and 287 repository roots agree; all seven profile groups and both corpora agree under instrumented native |
| Static-components supplied-graph kernel | 34 graphs, 27 findings, 14419 equal bytes; instrumented native agrees; not a full rule port |
| Inherited ten-rule coverage and guards | controls 53 findings/24994 equal bytes, instrumented native, ten rule faults, released handles and refusal/decoder controls PASS; package 192.981s |
| Checker Go package | PASS 0.480s |
| Filtered Node oracle | nine selected fixtures plus one-byte detection check PASS 15.861s |
| Production Go rules | selected core PASS 0.208s; React PASS 0.103s; nexus PASS 0.057s |
| Vet | empty output, exit 0 |

All measured rule fixes and suggestions are zero, matching unchanged Go
production. Both native programs and their C bridge archives are instrumented
where the check uses the bridge. The independent static kernel uses no bridge.
Native programs that should finish have empty stderr. The complete streams and
commands are preserved under `validation/landing/` and the local scratch trees
named in those command records.

Current main now supports RegExp construction. The first configured rerun
correctly exposed an obsolete expected refusal after all rule comparisons had
passed. Only this worker's `wave-29-configured/check.py` changed: the formerly
refused control now compares the source on Node, native, emitted JavaScript on
Node, and sanitized native. The first JavaScript attempt lacked the runtime
loader; using the repository's existing `oracle/node.mjs` resolves its `adamic`
import. The final full rerun passes, printing `match` identically in all four
runs. Both initial failure logs are retained. No shared harness or generator was
edited and the Go RE2 instruction VM remains the id-match implementation.

Every mutant rerun is accounted for below. Compiling rule/core mutations finish
normally with exit 0 and empty stderr, and independent bytes catch them. They
are not credited for compilation or sanitizer failures.

- Naming configuration: denylist drops source-reference exemption; match drops
  option eligibility. Both change Go/native findings.
- Default naming: wrong findings for empty denylist or empty pattern; byte
  comparison catches each at offset 51.
- Concurrency rule: report end increased by one; byte comparison offset 57.
- Provenance facts: source module erased; byte comparison offset 51.
- Regexp VM: empty-width context ignored; independent regexp decisions differ.
- Regexp raw facts: rune 95 incremented; independent regexp decisions differ.
- Globals: source-shadow resolution inverted; configured group 00 differs.
- Setter: getter recognized in place of setter; configured group 00 differs.
- Restricted-name shadow: all undefined declarations exempted; group 00 differs.
- Shorthand raw facts: ordinary accessor replaces value accessor; group 00 differs.
- Released registries: original/provenance, regexp and second-batch queries all
  reject a released handle with panic 70. Retaining the registry makes each
  request exit 0 and the required panic check catches it. The second-batch
  rejection is also checked under instrumentation with no sanitizer diagnostic.
- Static kernel: wrong creator at store binding or phi; each compiles and exits
  normally and changes one diagnostic. These are kernel tests only.
- Inherited coverage: before, cast, methods, coercion, caller, parameter,
  invariant, optional, alias and unused mutations all exit 0; Go bytes catch
  offsets 83, 1664, 2392, 1243, 10626, 395, 13272, 16143, 16673 and 17571.
  Its retained-registry mutation also violates the required panic 70.
- Inspect guards: wrong-kind and unknown-question guard mutations exit 0 and
  violate the refusal expectation; unmutated requests panic 70.
- Node oracle: the existing one-byte mutant check passes, proving comparison
  rejects its changed output.

The decoder's malformed wire controls are individually caught by panic 70:

- empty: panic 70, invalid checker facts frame.
- length-negative: panic 70, invalid checker facts length.
- length-short: panic 70, invalid checker facts length.
- trailing: panic 70, trailing checker facts.
- version: panic 70, unsupported checker facts schema.
- question: panic 70, unsupported checker facts schema.
- not-integer: panic 70, invalid checker facts integer.
- rounded-integer: panic 70, invalid checker facts integer.
- noncanonical-integer: panic 70, invalid checker facts integer.
- negative-natural: panic 70, negative checker fact.
- bad-boolean: panic 70, invalid checker facts boolean.
- zero-id: panic 70, invalid checker type record.
- negative-tuple: panic 70, invalid checker type record.
- missing-root: panic 70, missing checker type identity.
- missing-constraint: panic 70, missing checker type identity.
- missing-element: panic 70, missing checker type identity.
- duplicate-id: panic 70, invalid checker type record.
- missing-link-17: panic 70, missing checker type identity.
- missing-link-18: panic 70, missing checker type identity.

Whole-process measurements on the frozen corpora, including startup:

| Batch / corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Original / compiler | 2.113407713s | 0.356960001s | 5.92x |
| Original / repository | 0.308986099s | 0.142899671s | 2.16x |
| Second / compiler | 2.093302137s | 0.338014929s | 6.19x |
| Second / repository | 0.348427754s | 0.194516898s | 1.79x |

The three React rules remain owned and blocked: set-state-in-effect,
set-state-in-render and static-components. Their positive Go controls again
produce three JSX findings plus two hook-only findings. Native paired JSX is
still parsed as TypeAssertionExpression; native `<C/>` still panics at slash.
Published JSX work on the separate parser branch was already verified in
scratch in the earlier report, but is not part of this main. More fundamentally,
source-to-SSA lowering, captures, compilation-unit gates, manual memoization
handling, and control/post-dominance are absent from this branch. An updated
448-origin-ref search finds only the same wave08 supplied-HIR partial kernel;
this is supporting evidence, not an exhaustive proof of absence. The source
pipeline cannot be replaced by zero default findings or a supplied-graph test.
No full three-rule findings/fixes/suggestions agreement, full-rule mutant, new
bridge lifetime coverage or native/Go lint timing is claimed for that batch.

The full repository gate, inherited older-sixteen-rule suites, inherited coverage
corpora, and full lint comparison through emitted JavaScript were not rerun.
The requested completed wave-29 rules were rechecked on their configured
controls, frozen compiler/repository corpora and native sanitizer oracles.
The landing unit stops here rather than adding shared pipeline code or new
claims while the current React ports remain blocked.

Exact commands are in the JSON records and logs. The top-level checks used:

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-configured/check.py /workspace/wave29-regex-controls > /tmp/wave29-landing-configured.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-configured/regex_check.py /workspace/wave29-regex-vm > /tmp/wave29-landing-regex-vm.log 2>&1
ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-regex-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v > /tmp/wave29-landing-default.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-configured/regex_mutants.py /workspace/wave29-landing-regex-facts > /tmp/wave29-landing-regex-facts.log 2>&1
go build -buildmode=c-archive -o /workspace/wave29-next-checker.a ./bridge/tsgo/archive > /tmp/wave29-landing-next-archive.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-landing-next > /tmp/wave29-landing-next.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-next/fact_checks.py /workspace/wave29-landing-next-facts /workspace/wave29-landing-next > /tmp/wave29-landing-next-facts.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-third/check_static_core.py /workspace/wave29-landing-static-core > /tmp/wave29-landing-static-core.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-third/prove_gaps.py /workspace/wave29-landing-react-gaps > /tmp/wave29-landing-react-gaps.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$|^TestFactsDecoderGuards$|^TestInspectRequestRefusals$|^TestPinnedTypeFlags$' -count=1 -timeout=30m -v > /tmp/wave29-landing-coverage.log 2>&1
go test ./bridge/tsgo/checker -count=1 -timeout=10m > /tmp/wave29-landing-checker.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave29-landing-vet.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /tmp/wave29-landing-node-oracle.log 2>&1
# From cohere/: selected production core, React and nexus suites as recorded in their logs.
```
