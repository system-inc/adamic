Built: all nine algorithms rebased onto lint integration d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06; no new claims.
Commits: prior own tip c9e6a598711764ef500636a9248dcd5851e3b83a; rebased implementation ab8576fa74fce143ed3fb755126e2906389d4df5; this report commit is pushed only to codex/typeaware-wave-19.
Commands and outputs: current compiler rebuilt all nine rules; 50 normal/sanitized comparisons passed, 801774 canonical Go bytes; checker 0.483s, registry 0.396s, metadata 0.014s, vet PASS. Node/typeof mutants PASS 23.935s; released handles PASS 60.420s; production refusal PASS 16.575s.
Mutants: six latest-rule mutations freshly compiled and caught only by Go bytes at 56/2365/8388/66/3334/628; retained-registry mutation caught by all four panic checks; Node caught constructor, string, null and missing-slot classifier mutations. Earlier six rule mutation receipts remain in prior reports.
Uncovered: shared checker-context/bridge-link integration, lint emitted-JavaScript comparison and full gate including 17 external correctness checks. No selected check skipped or relaxed; no shared source was edited by this worker.

The 28-commit rebase completed without conflicts and retained integration changes. The new base changes native typeof classification and slot presence, so all owned native runners were rebuilt normally and under ASAN/UBSAN/LSAN. Bridge/question sources, owned rules and pinned Go cohere are unchanged; checker archives and independent production Go oracle binaries were reused. Live comparisons cover current texts in the frozen 77-root TypeScript compiler and 287-root repository populations, positive controls, isolated/index settings, strict typeof options, three require-await contract sources and four ASI sources. Findings, fixes and suggestions agree byte for byte. Native sanitizer stderr is empty.

The six latest-rule mutations are wrong Symbol callee, invalid typeof string accepted, await search ignored, wrong union flag, strict option ignored and incorrect initializer metadata. All compile and exit zero with empty stderr before independent Go bytes catch them. The retained-handle mutant also exits zero cleanly and fails four required-panic checks. Normal/sanitized releases pass. The production archive still refuses all four unregistered questions with panic 70. The shared options guard was not bypassed; the owned strict typeof oracle supplies ValidTypeofOptions explicitly.

The new integration tests independently compare constructors and strings, five null observations and slot presence against Node. Their mutants finish cleanly and only stdout differs. The filtered baseline includes the seven new typeof fixtures, inherited static fields and the existing one-byte oracle mutant. No protected compiler or oracle file was changed by this worker.

Shared RuleContext still has no checker program/query handle, and the shared harness native build does not link the bridge archive. Seven previously prepared registration lines remain unapplied under the shared-file restriction. Latest three descriptors declare ast.Kind names; private handed-node adapters retain pinned internal numeric IDs. Earlier six retain their documented legacy driver limitation. This is algorithm parity with pending shared production integration.

Fresh single-round original-three complete-process timing, measured after tests finished and after checking identical Go/native bytes:

| Corpus | Go seconds | Native seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 0.556916 | 2.153681 | 3.87x |
| Repository | 0.238838 | 0.380306 | 1.59x |

Native remains slower. These are single observations including loading and parsing, not medians or isolated classifier costs. Earlier latest-rule timing rounds remain historical. Same-unit cloud setup previously completed in 195s; nproc 5. Setup was not rerun for this rebase.

Commands sourced /workspace/adamic-tools/env.sh and set TMPDIR=/workspace/wave19-f801-scratch. All test streams were redirected to files:

```sh
python3 stage1/cohere/typeaware/wave_19_third/rebuild_landing.py "$TMPDIR" /workspace/wave19-typescript
python3 stage1/cohere/typeaware/wave_19_third/prove_native_mutants.py "$TMPDIR"
go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1
go test ./stage1/cohere/typeaware -run '^TestWave19ThirdNumericMetadata$' -count=1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestTypeOf(Constructor|StringLiteral|Null|NullSlotPresence)Mutant|TestNativeAgreesWithNode/internal/oracle/testdata/(typeof_.*|inherited_static_field_read)\.a$' -count=1 -timeout=15m -v
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS="$TMPDIR/third-release" go test ./stage1/cohere/typeaware -run '^TestWave19ThirdReleasedHandles$' -count=1 -v
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS="$TMPDIR/third-release" go test ./stage1/cohere/typeaware -run '^TestWave19ThirdProductionPending$' -count=1 -v
```

Logs: /tmp/wave19-typeof-base-parity.log, /tmp/wave19-typeof-base-mutants.log, /tmp/wave19-typeof-base-packages.log, /tmp/wave19-typeof-base-metadata.log, /tmp/wave19-typeof-base-vet.log, /tmp/wave19-typeof-base-node.log, /tmp/wave19-typeof-base-release.log, /tmp/wave19-typeof-base-pending.log and /tmp/wave19-typeof-base-bench.log. Raw rebuilt runner streams/results occupy the historical local directory $TMPDIR/landing-b8; that name does not describe the compiler base. New timing streams are in $TMPDIR/typeof-base-bench. Evidence stays local; only report/claim updates are committed after the rebase.

Explicit all-head fetch refreshed 654 origin refs. The 197 ranked rules, Markdown reservations on all origin branches and baseline ports leave zero available rules. No new claim is made. Remote main and area bases were checked before pushing and remain b6b1538b0 and d3a37422c. The full repository gate was not run; no claim is made that its 17 external correctness checks passed. None was skipped, relaxed or deleted to make the selected checks green.
