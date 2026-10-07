Built: native .a ports of no-throw-literal, no-useless-backreference and prefer-arrow-callback.
Commits: pushed claim dd9c6958 before implementation b2a0bc7231961d3707dfb60188b60e36672a73e4; evidence accompanies this report.
Commands and outputs: differential PASS 126.843s; checker PASS 0.179s; filtered Node oracle PASS 8.834s; go vet ./... clean; setup 31s, nproc 5.
Mutants: throw, backreference containment, arrow fix whitespace and RegExp constructor tracking caught solely by Go byte comparison; binding and identity retained-registry mutants caught by required released-handle panic; Node one-byte mutant caught.
Not covered: full repository gate, shared unknown-question mutation anchor and production JSX parsing; no new rule is blocked by those gaps.

The preceding batches were already pushed at 101aff2f. A fresh fetch inspected
348 origin refs, 33 unique Markdown claim blobs and both requested base branches.
The first eligible entries were global ranks 152, 153 and 154, each zero-volume.
Earlier entries were skipped as ported or claimed. The complete selection audit
is [selection.json](validation-wave-17-fourth/selection.json). Claim dd9c6958
was pushed before any implementation was written. All changes in this batch
are inside stage1/cohere/typeaware; no shared registration, harness or checker
file was edited. Existing raw symbol-identities and global-binding-facts suffice.

Each rule has its own .a file. Throw checks preserve production syntax recursion
and undefined-shadowing behavior. Backreference checks implement native capture,
reference and disjunction tracking, named duplicates, all five finding classes,
RegExp aliases and global-object members, constant patterns and flags, and Go's
64-bit numeric-reference wrap behavior. Arrow callbacks preserve scope ownership,
self-reference and arguments checks, bound-this handling, both options and exact
comment-sensitive fixes. Decisions run natively; the independent Go oracle loads
unchanged production rules and serializes findings, fixes and suggestions.

The final differential test accepts 335 upstream and targeted controls, reporting
197 findings and 72000 identical bytes: 22 throw, 111 backreference and 64 arrow.
Arrow findings contain 124 fix edits; all suggestion counts are zero in these
controls and in the production rules. Both ordinary and ASan/UBSan native builds
match the complete Go stream. Option streams match at 69480 bytes (allowNamed),
74369 (disallowUnbound), and 71586 (both), also under sanitizers. The frozen
77-root TypeScript compiler and 287-root repository manifests match at 5318 and
18485 bytes respectively, with zero findings, in both builds. Source hashes and
compressed complete streams are preserved in validation-wave-17-fourth.
Four syntactically invalid extracted candidates are excluded by the independent
Go parse-validity filter; they are not claimed as successful rule controls.

All four native mutants exit zero with empty sanitizer stderr. Independent Go
comparison catches throw's unconditional report at byte 16897, disabled capture
containment at byte 1187, extra arrow-fix whitespace at byte 326 and disabled
RegExp constructor tracking at byte 14477. Both existing raw binding and symbol
identity questions reject released programs with panic 70. Registry-retention
mutants finish cleanly and fail the required-panic assertion. The filtered Node
oracle additionally proves its one-byte comparator can fail.

An initial stage0 refusal for an inferred array of never was resolved with typed
empty arrays inside these new files. An early tracker mutant survived because
its consuming rule still imported the original helper; the test rejected that
run, and corrected mutation wiring makes both constructor and consuming rule
import the mutated helper. The failed run is preserved as pre-final.log.
Additional escaped-pattern controls initially parsed as legacy octal; their
fixture escaping was corrected before the final successful run. These were
validation-construction mistakes, not findings ignored by the comparison.

Quiet three-round alternating whole-process medians, with output equality on
every round:

| Corpus | Native | Go | Native / Go | Native bridge queries |
| --- | ---: | ---: | ---: | ---: |
| compiler | 2.394282s | 0.366925s | 6.53 | 452 |
| repository | 0.277371s | 0.117416s | 2.36 | 24 |

These are process measurements, including load, parsing, rule execution and
serialization. The measured native implementation is slower. Phase timings,
individual rounds and byte hashes are in timing/measurements.json.

Commands (all test output redirected to files):

```
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_FOURTH_ARTIFACTS=/workspace/wave-17-fourth ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^TestWave17FourthAgreementAndMutants$' -count=1 -v -timeout=20m
go test ./bridge/tsgo/checker -count=1 -v
go vet ./...
go test ./internal/oracle -run 'TestNativeAgreesWithNode/(functions|closures|generic_functions|method_closures)|TestTheOracleCatchesOneByte' -count=1 -v
python3 stage1/cohere/typeaware/validation-wave-17-fourth/measure.py /workspace/wave-17-fourth stage1/cohere/typeaware/validation-wave-17-fourth/timing /workspace/wave-17-typescript
```

Setup reported Go, clang, Node and submodules ready in 0s each, cache warm in
31s and total 31s on five visible processors (four CPU cgroup quota). TypeScript
is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8; cohere at
715ba94f3608a6500086b1076ce5cb7e51b836db.

The full shared gate was not run. Its existing unknown-request mutation test
hardcodes the former facts.go refusal line and cannot construct that mutant
after the prior dedicated-question fallback registration. Actual unknown-request
refusal and dedicated bridge tests were validated previously; this batch does
not change their code. Production JSX still refuses in the native parser; prior
Head decision validation uses the documented independent AST projection. Those
prior integration gaps are unchanged and are not represented as passing here.
