# Fourth slot 08 batch

Pushed previous work through 86546bbf, fetched 330 origin refs, then claimed default-case-last, default-param-last and for-direction in 98377d42 before implementation. Original helper-ready 46 are reserved; these are the first free syntax-ready inventory entries, positions 13-15. selection.json records the fetched ref SHAs and claim exclusions, with main source checks as well as descriptors.

Three independent directories contain .a rules/messages, descriptors, upstream Go adapters, witnesses and mutants. These rules have no repairs or suggestions in Go. Core default-param-last reports rest parameters before required parameters, unlike the namespaced sibling. for-direction mirrors Go's literal-only static sign analysis rather than adding scope resolution.

Observed parity, uncached: 234 compiler/stage1 files and 13,076,927 serialized bytes matched upstream Go on source Node, emitted JavaScript and ASan/UBSan native. TypeScript compiler pin: 050880ce59e30b356b686bd3144efe24f875ebc8. All 212 captured upstream cases matched: default-case-last 38, default-param-last 128, for-direction 46; 58,751 bytes. No batch cases omitted. Owned witnesses matched 21,076 bytes. go vet with overlay and filtered TestTheOracleCatchesOneByte passed. Full repository gate was not run.

Best of five rotating process timings, compiler's 77 files plus 1,000 positive examples. Native release build, Node source, upstream Go; startup included. Counts agree on every timing run:

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| default-case-last | 1000 | 669.05 | 833.71 | 4177.66 |
| default-param-last | 1024 | 690.30 | 871.87 | 4170.18 |
| for-direction | 1000 | 652.14 | 986.54 | 4158.56 |

Setup: Go ready 0s, clang/Node/submodules ready 1s, cache warm 26s, total 26s, nproc 5. Evidence logs contain exact observations; validate.py reproduces the owned scratch comparison setup using the earlier owned compatibility.patch. Shared production files were not changed. The now-available origin/codex/lint-harness-dot-a branch was not merged into this owned-directory unit. This checkout still needs the scratch overlay for .a registration, emitted comparison and the shared profile compile fix, so default integration is not claimed. Earlier batch JSX and parser-recovery gaps remain recorded in no-useless-computed-key/REPORT.md.

Mutants default_clause_suppressed, core_default_suppressed and loop_direction_reversed all compiled and ran with exit 0 and empty stderr on source Node, emitted JavaScript and sanitized native. Only serialized-output comparison caught each. Mutant gate passed in 69.01s; registry tests passed in 0.064s.
