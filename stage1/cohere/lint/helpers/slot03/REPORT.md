Built three .a helpers: React base names, Tailwind whitespace, and Tailwind listener kinds; 38 dependencies removed for 26 rules, zero final blockers removed alone.
Commits: 121a6c3 (base names), c8bfb9c (whitespace), 61d7f55 (listener kinds and handoff), on codex/lint-helpers-03.
Commands: setup succeeded in 106s, nproc=5; all helper packages passed (152.830s and 48.475s); vet, formatting and six-fixture uncached input oracle passed.
Mutants: dropping PureComponent, dropping vertical tab, replacing VariableDeclaration, sharing the listener array, and dropping a consumer were caught; all four original helper mutants were re-killed.
Not covered: full repository gate, complete lint-rule findings/fixes, live Tailwind installation/corpora, and integration into existing rule entry points.

## Delivered behavior and rule handoff

| File | Go symbol | Remaining consumers | Additional fully helper-ready rules |
|---|---|---:|---:|
| component_base_name.a | react.isComponentBaseName | 14 | 0 |
| tailwind_space.a | tailwind.isSpace | 12 | 0 |
| listener_kinds.a | tailwind.ListenerKinds | 12 | 0 |

The two Tailwind helpers share the same twelve consumers. [README.md](README.md) names all fourteen React/structure rules and all twelve Tailwind rules, and documents each API. [readiness.json](readiness.json) records every removed and residual dependency. These are observed helper comparisons and a dependency calculation, not a claim that any whole rule is implemented. Other slots' helpers are deliberately not counted in this ledger.

Base: origin/codex/lint-helpers at 29990b4, containing the four original helpers at 5d13f5b. The clone fetched only main initially, so an explicit wildcard refspec fetched every origin/codex/lint-helpers* branch. Every remote claim was read before each selection. Concurrent reservations collided repeatedly; earlier commit times determined ownership. This slot withdrew identifier matching, FileContextFor and the class predicate, and delivers none of their implementations. The final active ownership is in claims/03.md.

Claims preceded implementation: 66e752b for the React name leaf, 6cdb955 for whitespace, c8bfb9c for listener kinds. Slot 03 retained these against later duplicate claims. No compiler file or cohere worktree was changed. All new Adamic source files are .a.

## Independent evidence

The Go overlay exports the actual private helpers and calls the actual public ListenerKinds. The source capture observes actual runtime fixture arguments, including dynamically assembled inputs, for all 26 recorded consumer rules. The pinned cohere commit is 715ba94f3608a6500086b1076ce5cb7e51b836db. Coverage and pin drift fail the differential harness.

Capture recorded 1,519 rule/file/source inputs. Names come from the Go AST's decoded Identifier, StringLiteral and PrivateIdentifier text, augmented by exact-case, suffix, whitespace, NUL and Unicode controls. Whitespace is compared on source code points, signed rune extremes and every value from 0 through 0x10ffff, including surrogate values. Listener values, order, independent mutation across calls and membership for every node kind seen in the captured inputs are compared. Node executes the unchanged .a source; native uses ASan/UBSan with Linux leak checking. Both must exit 0 and write no stderr before output is compared.

Independent captures before and after the listener claim produced identical compressed source bytes. evidence/reproducibility.log records the SHA-256. The expanded million-line observation is generated in scratch space and is not committed.

## Commands and observed output

All test output was written directly to log files, never piped. `source /workspace/adamic-tools/env.sh` preceded setup-dependent commands. Setup ran `bash cloud/setup.sh > /tmp/lint-helpers-03-setup.log 2>&1`; evidence/setup.log preserves its output:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (106s)
setup: done in 106s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed 5. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

- `python3 stage1/cohere/lint/helpers/slot03/testdata/regenerate.py > .../evidence/regenerate.log 2>&1`: exit 0, 1,519 captured inputs, all 26 consumers. The capture invokes Go rule tests which exit 1 for the eight known unavailable-Tailwind/live-corpus failures; see the limitations below. It does not label that upstream gate as passed.
- `go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > .../evidence/three-helpers.log 2>&1`: PASS, 18.903s, 1,115,085 matching output lines, four compiling semantic mutants caught. This preceded the separate-stderr and metadata checks added for the final regression run.
- `go vet ./... > .../evidence/vet.log 2>&1`: exit 0, empty log.
- `gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > .../evidence/gofmt.log`: exit 0, empty log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > .../evidence/oracle.log 2>&1`: PASS, 28.701s, all six fixtures, six probe misses, zero cache hits.

`go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=20m > .../evidence/helpers-final.log 2>&1`: PASS for the original helper package in 152.830s and slot 03 in 48.475s. The final harness compared 1,115,085 identical output lines, required empty stderr and exit 0, re-killed all four source mutants, and caught the missing-consumer mutant. The original four-helper package also passed its 22,412-line comparison, ten message refusal cases, known-gap checks and four original source mutants.

## Mutants

Every source variant is built from temporary copies. Compile failure, panic, sanitizer output or nonempty stderr is rejected by the harness and never credited as a semantic mutant caught. Successful native execution must differ on a real output line from Go's independent result.

| Mutation | Witness |
|---|---|
| Base-name helper omits PureComponent | line 159: false instead of true |
| Whitespace helper omits vertical tab | line 847: false instead of true |
| Listener helper substitutes StringLiteral for VariableDeclaration | line 1,114,948: wrong ordered list |
| Listener helper returns a shared global array | line 1,114,949: mutation from the preceding caller is visible |
| Coverage omits better-tailwindcss/no-unknown-classes | missing-consumer error, rather than silently accepting skipped external fixtures |

The original regression suite also re-killed its existing mutants: JSON permitting a raw newline (line 15,157, Go invalid vs mutant valid); schema accepting multiple oneOf matches (line 15,166); strict target ignoring an unknown field (line 15,178); and policy rendering omitting value interpolation (line 22,166). The schema/target mutants accepted inputs Go refused; policy output retained the placeholder where Go supplied the value.

The coverage mutant is a metadata verdict check, not an Adamic source mutant. The four source mutants prove all three delivered helpers' semantic comparisons can fail.

## Limits and findings

Observed: Go's Tailwind rule package requires an installed tailwindcss and live corpora under hardcoded /Users/kirkouimet/Projects/ahra paths. Eight existing rule-package guard tests fail here. Actual fixture source arguments are captured before engine-dependent skips, so all twelve affected helper consumers are represented, but their full findings and fixes were not executed against a live design system. The generator permits only those named existing guard failures and aborts on unknown failures or missing consumer coverage. React and structure rule-package captures pass. See evidence/capture.log for every failure.

The first discarded identifier oracle showed Go's private isIdentifierNamed panics on nil through SkipParentheses. It was not counted as a passing parity run. A discarded file-context runner initially treated readTextFile's result as a string; the final runner unwraps the tagged result explicitly. Both helpers were then yielded because another slot held an earlier claim. A locally passing class helper was likewise discarded after ownership refresh; its mutant dropping ClassExpression was caught at line 2,068 but is not counted as a delivered helper. No duplicate helper or failed attempt is counted in this delivery.

Not covered: arbitrary invalid UTF-8 Go strings, every possible signed 32-bit rune value outside the tested Unicode interval and sampled extremes, full lint-rule replay/fixes, cache/reader construction, rule integration, concurrent API use, or the full repository gate. Integration of other slots' adapters and dependency predicates remains the rule workers' responsibility. No rule's stage-1 implementation status was rewritten.
