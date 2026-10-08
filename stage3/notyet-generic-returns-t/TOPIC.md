Built primitive branded-string return signatures, including undefined through the existing NULL sentinel.
Base: origin/main 6998ebc24ae353193cb1495d3d51308131a4b5c7; branch codex/notyet-generic-returns-t-topic.
Focused oracle tests pass (12.239s), lower tests pass (0.629s), counts refresh passes (23.727s).
Both new semantic mutants disagree with Node in native and JavaScript; three production-rule mutants fail their intended tests.
Not covered: uninstantiated generic declarations, branded value slots, checked casts/views, and six reserved clock kinds.

Only our nine non-merge commits were carried forward:
2f92582e -> 135e4f40
fd0e1225 -> 6778a6dd
b647eebb -> 6c3eac89
fe2bd0e0 -> b2228e2e
a53c8958 -> 5743b2f9
3063d1a9 -> 816cb936
af7f71ab -> 1677d489
d61b40d3 -> a9129300
2ec964a8 -> 4a33e8df
No area, census, or checked-non-null merge commit was imported. Counts regeneration removes seven stale checked-non-null rows whose fixtures are absent on this base.

Morning census e8c283b5, stage3/notyet-table/rerun-0730/after/roots.csv: filtering kind=NotYet and counting rows by exact reason gives 16 __String | undefined roots and 15 __String roots. These 31 signature roots have a representation; this is not a claim that 31 complete compiler functions lower. The larger T (100) and T | undefined (75) groups remain declarations without concrete instantiation; carried concrete-call fixtures pass without erasing T. The next new signature cases handled here are the two branded-string groups. Explicit any remains an adaptation refusal.

The signature hook uses the existing Map-key proof: a primitive string with only void phantom markers, with no conflicting primitive members, callable/indexed brand, never field, or runtime field. It does not change expression.go typeOf, the IR, either backend, or runtime C. Existing string undefined/null sentinel behavior is reused. Ordinary branded value representations remain with their owner. A checked cast needed by binder.ts:663:21 is left refused; no checked view is added.

Fixtures, each registered from its own _test.go:
internal/oracle/testdata/notyet_signature_escaped_string.a
Node prints:
plain-brand: __call __new
string 3
internal/oracle/testdata/notyet_signature_optional_escaped_string.a
Node prints:
optional-brand: __call undefined
string undefined 4
Run Node with: node --disable-warning=ExperimentalWarning oracle/node.mjs <absolute fixture path>
Each agrees in release native, ASan/UBSan native with leak checks, and JavaScript.

Replays use the prior authorized census replay worker as an external scratch overlay, not tracked code or a merge. Worker built with:
go build -buildvcs=false -overlay=/tmp/notyet-topic-replay-after/overlay/overlay.json -o /tmp/notyet-topic-replay-after/worker ./stage3/census/latent/tool
For each row below run:
/tmp/notyet-topic-replay-after/worker -project /tmp/notyet-generic-adapted/src/tsc/tsc.ts -where /tmp/notyet-generic-adapted/src/compiler/<site> -kind NotYet -reason '<original reason>'
Before, all four reproduced their exact original function-signature reason. After, each exits 1 with its next stop:
binder.ts:661:14 (__String | undefined): Refused, binder.ts:663:21, a cast the runtime can't check.
checker.ts:19210:14 (__String | undefined): NotYet, checker.ts:19213:13, a BinaryExpression with a value and a boolean.
checker.ts:2436:14 (__String): Refused, checker.ts:2439:17, a value as a condition.
checker.ts:38352:14 (__String): NotYet, checker.ts:38355:23, a value of type __String.
Before/after JSON and logs: /tmp/notyet-topic-replay-{before,after}/{1,2,3,4}.{json,log}.

New mutants:
TestEscapedStringReturnMutant replaces the enum return with "wrong": native and JavaScript stdout disagree with Node, clean exits and no leaks.
TestOptionalEscapedStringReturnMutant replaces undefined with "wrong": same checks catch it.
Production overlay optional-undefined removes the undefined-member rule: optional fixture fails at its original signature NotYet.
Production overlay plain-signature removes the signature hook: plain fixture fails at its original signature NotYet.
Production overlay unproven-brand accepts every member as string: TestStringBrandSignatureGuards rejects the mutant. Guards cover runtime fields, never fields, primitive-member collisions, callable brands and existing indexed-brand refusal.
Logs and overlays: /tmp/notyet-topic-rule-mutants/.

Exact final commands (source /workspace/adamic-tools/env.sh first; output redirected to the named logs):
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_|TestNotYetGenericReturnsMutants|TestExplicitGenericReturnMutants|TestExplicitNullableArgumentsHaveDistinctInstances|TestUndefinedSignatureMutant|TestVoidUnion(SignatureMutant|RepresentationMutants)|TestObjectSignatureMutant|TestMap(KeyMutants|ExplicitAnyRefused)|TestPatternDefaultMutants|Test(Optional)?EscapedStringReturnMutant' -count=1 -v
/tmp/notyet-topic-final-oracles.log: PASS, 12.239s.
go test ./internal/lower -run 'TestStringBrandSignatureGuards|TestMapKeyBrandGuards|TestGenericFunctionPolymorphicRecursionIsRefused|TestGenericUnionFixtureHasSeparateInstances|TestGenericJSONUnionArrayIsNotYet|TestInheritanceGeneric|TestInheritanceKeepsNominalTupleDestructuring' -count=1 -v
/tmp/notyet-topic-final-lower.log: PASS, 0.629s.
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
/tmp/notyet-topic-counts.log: PASS, 23.727s.
No full package suite or full gate ran.

Setup: export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh. Total 46.328s; Node ready .036s, Go .040s, markdown ready .128s, submodules .133s, clang .303s, build ready 46.138s, tests deferred 46.291s, cache warm 46.293s. nproc=5 (cgroup quota=4). Setup passed without a workaround.
