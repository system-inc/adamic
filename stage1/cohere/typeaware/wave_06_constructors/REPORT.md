Built: native Adamic ports of no-new-func, no-new-native-nonconstructor and no-new-wrappers.
Commits: earlier six rules completed in d7f911f4; claim bde4cbde463a07459639f184ed87dd4373cbd00a pushed before this implementation.
Checks: owned validator compares full findings, fixes and suggestions on controls and the frozen compiler/repository populations.
Mutants: one semantic mutation and one diagnostic identifier mutation per rule, plus a declaration-file fact mutation; independent Go bytes must reject each.
Limits: default configurations and finite populations, without a full root test gate or exhaustive upstream fixture matrix.

These were the first three remaining entries after fetching all origin heads and checking 341 origin refs, 33 unique Markdown claim blobs, 197 ranked rules, 117 claimed rules and 25 ranked base/main ports. All three have zero recorded compiler and repository volume. Ranking is descending combined compiler/repository count, lexical on ties. The six earlier ports were already tested and pushed before this claim; no additional rules are reserved.

Each rule lives in its own directory and all new Adamic files use `.a`. `global.a` reads the existing raw `wave06-declarations` symbol metadata, checking the first declaration's declaration-file flag exactly as Go's `resolvesToAGlobal` does. It does not unwrap import aliases or require every declaration to be ambient. No new bridge question was necessary. Existing shared registration, harness, bridge and compiler source files are untouched.

`no-new-func` follows parenthesized callees, recognizes direct calls/construction and static apply/bind/call methods, and reports the invoking expression. `no-new-native-nonconstructor` recognizes only unparenthesized Symbol/BigInt identifier callees and reports the identifier. `no-new-wrappers` follows parentheses for String/Number/Boolean and reports the whole new expression. Preserving these differences matters for byte agreement with the production Go rules. Native decisions do not call Go lint implementations.

The first control comparison caught a missing finding for `(Function?.call)(...)`: the native parser carries an optional-chain token before the property name. The owned rule now reads the last member child. Added optional computed members and cooked identifier/string escapes also match. An initial config containing only `.a` files failed TypeScript config discovery; the owned validator now supplies an explicit empty declaration root, as the existing suites do. Both failures and their corrections are distinct from final passing evidence.

Reproduce with the configured toolchain, from the repository:

```
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_constructors/validate.py --scratch /workspace/wave-06-constructors-final --compiler /workspace/wave-06-typescript > /tmp/wave-06-constructors-final.log 2>&1
```

The owned Go oracle is built through a virtual-main overlay within pinned cohere and calls these three unmodified production rules, with independent loading and AST traversal. The native runner and Go driver serialize every diagnostic field, fix and suggestion and sort whole records while retaining duplicates. All three production rules return no fixes or suggestions. Sanitizer runs compare those same complete streams, not just totals.

Controls exercise direct/indirect invocation, optional members, computed strings/templates, parentheses, local and imported shadows, nested declarations, aliases, nonidentifier receivers, uncalled members, multiple findings, Unicode and CRLF. The frozen corpus runs are useful for avoiding unintended findings, but their zero totals alone cannot demonstrate positive behavior.

The compiler checkout is TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`, cohere at `715ba94f3608a6500086b1076ce5cb7e51b836db`, and typescript-go at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Input manifests reuse the original 77 compiler and 287 repository roots. No submodule pins changed. Setup reused the original wave's successful 77-second setup: `nproc` 5, CPU quota 4, Go 1.27.1, clang 20.1.8 and Node 24.19.0.

Each diagnostic-ID mutant exits 0 with empty stderr and is caught only by the independent byte comparison. The raw fact mutant flips the declaration-file flag without modifying the oracle; it must likewise exit normally and differ in findings. Querying the existing declaration question after program release must exit 70 with `invalid or released checker handle`. The previously completed wave's released-registry mutation and full bridge regressions remain recorded in `../wave_06_next/validation/`; this continuation does not change that bridge.

Complete compressed command stdout/stderr, run commands/exits/times, input manifests and hashes are saved under `validation/`. Timings are single fresh-process, warm-filesystem wall times including parsing/loading and excluding builds, measured before concurrent regression commands. They do not establish universal equivalence or a performance improvement. No shared-harness gap blocks these rules in the owned native runner.

Final byte comparison passed:

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Controls | 109 | 51 | 23011 |
| Compiler | 77 | 0 | 5318 |
| Repository | 287 | 0 | 18485 |

Positive control findings by rule: `no-new-func` 18, `no-new-native-nonconstructor` 9, `no-new-wrappers` 24. Each population also matched under ASan/UBSan with empty sanitizer stderr. Released-handle panic 70 and the normally exiting fact mutant were caught.

| Timed process | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.833564 | 0.350764 | 5.227 |
| repository | 0.272956 | 0.136781 | 1.996 |

The additional semantic mutants all finished normally with empty stderr and differed from Go: removing `no-new-func`'s shadow discrimination, anchoring the nonconstructor finding on the whole new expression, and removing wrapper parenthesis traversal. Their streams and commands are retained separately as `*-semantic-*` and `semantic-runs.json`. The reusable validator now runs these semantic mutants. The earlier successful diagnostic-ID mutations remain preserved as additional evidence.

`TMPDIR=/workspace go -C cohere test ./internal/lint/rules/core -run '^TestNoNew(Func|NativeNonconstructor|Wrappers)' -count=1 -timeout=10m` passed. Its output is in `validation/go-tests.txt`. The new Go driver was formatted with gofmt, the validator parses successfully, and `git diff --check` passed. Existing shared bridge regressions were not rerun because this continuation changes no bridge source. All test output went directly to files.
