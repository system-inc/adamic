Built: native no-throw-literal, no-useless-backreference and prefer-arrow-callback, with exact callback fixes.
Commits: earlier nine rules complete through 7e889b14; claim df75b525 preceded implementation and push.
Checks: owned independent Go byte comparison, sanitizer and released-handle checks over controls and frozen corpora.
Mutants: inverted throw judgment, reversed forward-reference test, corrupted callback arrow insertion, and flipped declaration-file fact.
Limits: default configurations and measured source populations; no full root gate or exhaustive upstream fixture/options matrix.

The next three entries were selected after an all-heads fetch and scan of 347 origin refs, 33 unique claim documents, 126 claimed ranked rules and 25 ranked ports on main or the bridge base. All three recorded volumes were zero. Ranking is descending compiler plus repository count, lexical on ties. The earlier nine rules were fully tested and pushed before this claim. The intervening core rules were already claimed elsewhere; no additional rules are reserved here.

Each native rule lives in its own directory. New Adamic source uses `.a`; the runner, regex helpers and validation script are also owned files. Existing shared registration, test harness and compiler files are unchanged. The existing raw `wave06-declarations` question supplies symbol identity and declaration metadata. No new question or dispatcher edit was necessary.

`no-throw-literal` mirrors the syntactic couldBeError cases and the separate global-undefined test. It preserves Go's distinctions between arithmetic assignments, logical assignments, comma and conditional branches, and its treatment of explicit parenthesis nodes. `prefer-arrow-callback` tracks function ownership of this, super, new.target and arguments, resolves self references, follows callback and bind shapes, preserves fix declines, and emits fixes in Go's order. `no-useless-backreference` scans regex group/reference paths, alternation, lookaround direction, negative assertions, duplicate named groups, character classes and escapes; it traces global RegExp aliases and constant patterns natively. Go supplies compiler facts rather than lint judgments.

The controls caught a crash in this wave's existing declaration helper: `Node.Text()` cannot read a destructuring name. `bridge/tsgo/checker/wave06_declarations.go` now limits that raw name projection to text-bearing AST kinds, preserving an empty name for binding patterns and computed names. The existing question and ABI framing are unchanged. Native compilation also required explicit nullable guards and digit arithmetic instead of a Number call; those were corrected in owned native files without compiler edits.

The 210 controls include positive and clean throws, shadowed undefined, recursive callbacks and shadowed self names, implicit versus declared arguments, nested ownership, generators, bind chains and comments, callback repair declines, Unicode/CRLF fix spans, all five backreference diagnoses, duplicate named groups, u/v-mode classes, invalid references, constructor aliases, destructuring, constant bindings and modified globals. They produce 136 findings with byte agreement including callback repair arrays. These positive controls matter because both recorded corpus volumes are zero.

Reproduce from the repository with the configured toolchain:

```
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_callbacks/validate.py --scratch /workspace/wave-06-callbacks-final --compiler /workspace/wave-06-typescript > /tmp/wave-06-callbacks-final.log 2>&1
```

The owned Go driver is built through a virtual-main overlay in pinned cohere, using unmodified production rules with independent loading and AST traversal. Full diagnostic records are compared, including ordered fixes and suggestions, retaining duplicate findings. These default rules produce no suggestions; prefer-arrow-callback emits fixes and the other two emit findings only. Neither oracle rule source nor shared harness was edited.

TypeScript is v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`, cohere at `715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Manifests reuse the original 77 compiler and 287 repository roots. Setup reused the original successful 77-second setup: nproc 5, CPU quota 4, Go 1.27.1, clang 20.1.8 and Node 24.19.0. No pins changed.

Each semantic rule mutant and the raw declaration-file fact mutant must compile, exit 0 with empty stderr, and be rejected by independent Go diagnostic bytes. Inspecting the declaration question after program release must panic with exit 70. All test output goes directly to logs. Complete compressed streams, commands, exits, elapsed times, manifests and hashes are under `validation/`.

Timings measure one fresh process on a warm filesystem, including loading/parsing but excluding builds, before concurrent regression commands. They support only the measured populations, not universal semantic equivalence or a performance improvement. Default configurations are covered; nondefault prefer-arrow-callback flags and a full upstream fixture matrix were not exercised. No shared-harness gap blocks these rules in the owned native runner.

Final results:

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Controls | 210 | 136 | 50914 |
| Compiler | 77 | 0 | 5318 |
| Repository | 287 | 0 | 18485 |

Control findings: `no-throw-literal` 16, `prefer-arrow-callback` 36, `no-useless-backreference` 84. Full records include 71 repair entries and 0 suggestions. All populations also match under ASan/UBSan with empty sanitizer stderr. All semantic mutants exit 0 and differ from Go; released handle exits 70 with the expected panic.

| Timed process | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.943704 | 0.419421 | 7.018 |
| repository | 0.391958 | 0.143270 | 2.736 |

Final regression commands passed, with their full logs in validation:

- `TMPDIR=/workspace go test ./bridge/tsgo/... -count=1 -timeout=15m -v`
- `TMPDIR=/workspace go -C cohere test ./internal/lint/rules/core -run '^Test(NoThrowLiteral|NoUselessBackreference|PreferArrowCallback)' -count=1 -timeout=10m` (0.580s).
- `TMPDIR=/workspace go vet ./...` and `gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware` (empty logs).
- Earlier owned `wave_06_next/validate.py` and `wave_06_constructors/validate.py`, each with a fresh scratch directory and `--initial`, still match independent Go on their 84/42 and 109/51 controls/findings. Their path-dependent streams contain 29368 and 24101 identical bytes respectively.
- `git diff --check` and Python source compilation passed.

The root test gate, a new filtered Node run and every upstream fixture/options configuration were not run. The compiler and runtime are unchanged; the bridge's existing foundation, ownership, sanitizer and mutant checks passed in the bridge regression.
