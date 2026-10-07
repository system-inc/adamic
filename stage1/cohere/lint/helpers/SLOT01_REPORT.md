Built: FileContextFor, IsEs6ComponentClass and DefaultClassLiteralSettings in three .a files.  
Implementation commits: feb52ac, 6c2c75b, de05d72; claims were pushed before each implementation.  
Commands/results: complete helper package PASS, 65.010s; go vet ./... exit 0; filtered uncached input oracle PASS, 22.563s.  
Mutants: all six delivered-helper mutants compiled and were caught by Go output comparisons; inherited and withdrawn runs are recorded below.  
Not covered: whole-rule findings/fixes, dynamic fixture generation, cross-slot adapter integration and the full repository test gate.  

# Slot 01 lint helper report

Branch: `codex/lint-helpers-01`, from `origin/codex/lint-helpers` at `29990b47913aa77d542045194b512484af995d29`, which contains the first four helpers at `5d13f5b`. Go cohere is pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`. No compiler ownership files, existing rule entry points, submodule source or frozen readiness ledger were edited. No pull request is opened.

## Delivered behavior and readiness

| Helper | Consumers | Additional final blockers removed |
|---|---:|---:|
| FileContextFor | 18 | 3 |
| IsEs6ComponentClass | 14 | 0 |
| DefaultClassLiteralSettings | 12 | 0 |

The three helpers remove dependencies for 44 distinct rules. Under the inventory's common AST adapter assumption, the only additional fully helper-ready rules are `structure/network-no-direct-fetch`, `structure/react-hook-require-result-naming` and `structure/storage-no-direct-local-storage`. The class API requires the separately owned Go-equivalent `isComponentBase` callback. These counts describe helper readiness, not implemented rules or observed findings parity. `slot01_readiness.json` publishes every residual dependency rather than guessing what other slots will deliver.

File context preserves all nine flags, including lexical substring exemptions and Windows path normalization. The class predicate preserves declarations, expressions, empty/nil heritage, named type fields and Go's unfiltered heritage token behavior. Tailwind defaults preserve all three ordered lists and return fresh independently mutable arrays. `SLOT01_README.md` documents the APIs and adapter contract.

## Commands and observations

All test stdout went directly to log files. Baselines compare real Go helpers, Node running the same `.a` source through `oracle/node.mjs`, and ASan/UBSan native builds with default Linux leak detection. Successful baseline and semantic mutant processes must exit 0 with empty stderr; compilation failure, a panic or sanitizer output cannot count as a killed semantic mutant.

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` printed **5**. `source /workspace/adamic-tools/env.sh` was used in each build shell.

`bash cloud/setup.sh > /tmp/lint-helpers-01-setup-retry.log 2>&1` completed after the branch settled:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (75s)
setup: done in 75s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The first setup attempt overlapped checkout because the narrow remote fetch initially lacked the helper branch. Cache warming observed removed files `internal/lower/definite_assignment_test.go`, `internal/oracle/literal_optional_test.go`, `internal/native/borrow_consumes_test.go`, and stale `stage1/cohere/lint` references to `volumeGenerated` and `checkRecoveryRefusal`. It printed FAIL. This was a moving-tree run, not accepted evidence; the settled-tree retry succeeded. Both logs are retained.

- `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01/final.log 2>&1`: PASS, 65.010s, exit 0. All seven top-level tests ran.
- `go vet ./... > stage1/cohere/lint/helpers/evidence/slot01/vet.log 2>&1`: exit 0, empty log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01/oracle.log 2>&1`: PASS, 22.563s; all six fixtures, zero probe hits and six misses.
- `gofmt -l cmd internal stage1/cohere/lint/helpers`: empty output. `git diff --check`: empty output.

The first complete helper-package run took 119.037s and failed only because inherited `TestHelpersMatchCohere` treats any stderr as failure: its Go build printed downloads of `github.com/dlclark/regexp2/v2 v2.5.2` and `github.com/dlclark/regexp2 v1.11.5`. Every slot 01 test and the other inherited tests passed in that run. The downloaded dependencies were cached, then the same full package command was rerun without weakening the stderr condition. See `package-initial.log`.

Final slot 01 corpus totals: 1,815 file-context answers from 895 extracted fixture strings plus controls; 62,020 class answers over 3,548 source batches, from 2,628 fixture strings plus controls; 5,808 defaults/freshness answers over 1,936 call batches, from 1,016 fixture strings plus controls. Every consumer listed below contributes a required fixture file. The corpus extracts statically evaluable Go string expressions, including complete constant concatenations, rather than claiming to run the whole rule harness. Labels and message strings are included as additional controls.

## Mutants run

| Delivered file | Compiling semantic mutation | Observation that caught it |
|---|---|---|
| structure_file_context.a | remove backslash normalization | output line 7: Windows page path loses special/page flags |
| react_es6_component_class.a | decline ClassExpression nodes | output line 13400: Go true, mutant false |
| tailwind_default_class_literal_settings.a | omit class from attribute defaults | output line 1: ordered list differs |
| tailwind_default_class_literal_settings.a | share attributeNames between calls | output line 3: later defaults inherit mutated attribute |
| tailwind_default_class_literal_settings.a | share calleeNames between calls | output line 3: later defaults inherit mutated callee |
| tailwind_default_class_literal_settings.a | share variablePatterns between calls | output line 3: later defaults inherit mutated pattern |

The full package also ran the four inherited mutants: JSON raw-control acceptance (line 15157), overlapping oneOf acceptance (15166), unknown target-field acceptance (15178), and omitted message interpolation (22166). Each compiled and was caught. The package also ran the ten inherited message refusal comparisons and explicit unsupported-feature checks.

A withdrawn duplicate `isComponentBaseName` experiment passed 3,547 Go/Node/native answers and caught its compiling mutant that omitted PureComponent. Its file was removed and is neither delivered nor counted. The earlier withdrawn identifier experiment stopped at Go's nil-input panic in `ast.SkipParentheses`; no identifier semantic mutant ran and no parity result was claimed. Only these three final source helpers are delivered.

## Claims and concurrency

The repository initially fetched only main. Explicit wildcard fetches loaded every `origin/codex/lint-helpers*` branch before selections and refreshes. Comments were excluded because their earlier claim lives in `HELPERS.md`, even before claims/ existed. Multiple workers selected identical leaves between refreshes; timestamps resolved precedence, and later duplicates were yielded.

- Identifier: slot 01 claim `86d7292` yielded to the other workers and delivered no code.
- File context: slot 01 `751526e` at 00:23:54 UTC precedes slot 03 `f600f2c` at 00:24:05. Implementation `feb52ac`.
- Component base name: slot 01 `2050553` at 00:26:12 yielded to slot 03 `66e752b` at 00:26:01. No duplicate delivered.
- ES6 class: slot 01 `bf6e048` at 00:27:33 precedes slot 03 `121a6c3` at 00:27:52. Implementation `6c2c75b`.
- Tailwind defaults: claim `ccbb639` pushed before code. Implementation `de05d72`.

Each replacement was the highest remaining unclaimed consumer count at the last remote refresh. Ties selected prerequisites before the larger reader/class family. Claim history and final ownership are in `claims/01.md`.

## Per-helper consumer rules

### structure.FileContextFor

- `structure/boundary-no-project-theme-value`
- `structure/network-no-direct-fetch`
- `structure/network-no-forbidden-import`
- `structure/next-no-page-state`
- `structure/next-require-api-parameter-name`
- `structure/next-require-page-default-export`
- `structure/react-component-no-const-assignment`
- `structure/react-component-no-destructuring`
- `structure/react-component-no-separate-named-export`
- `structure/react-component-require-named-export`
- `structure/react-component-require-properties-parameter`
- `structure/react-component-require-properties-type-suffix`
- `structure/react-element-no-anchor`
- `structure/react-element-no-horizontal-rule`
- `structure/react-hook-no-properties-in-dependencies`
- `structure/react-hook-require-effect-comment`
- `structure/react-hook-require-result-naming`
- `structure/storage-no-direct-local-storage`

### react.IsEs6ComponentClass

- `react/no-arrow-function-lifecycle`
- `react/no-did-mount-set-state`
- `react/no-did-update-set-state`
- `react/no-direct-mutation-state`
- `react/no-string-refs`
- `react/no-this-in-sfc`
- `react/no-typos`
- `react/no-unused-class-component-methods`
- `react/no-will-update-set-state`
- `react/prefer-stateless-function`
- `react/require-optimization`
- `react/state-in-constructor`
- `structure/consistency-require-matching-file-name`
- `structure/react-component-no-display-name`

### tailwind.DefaultClassLiteralSettings

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-concatenated-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

## Not covered

The full repository `go test ./...` gate was not run. This unit ran the entire touched helper package, repository-wide vet, formatting checks and a filtered uncached external oracle. Whole rule listeners, defaults/custom settings conversion, findings/spans/fixes, arbitrary or dynamically generated rule inputs, production AST adaptation and cross-slot callback wiring are not covered. The default-settings helper does not compile regexes or implement the Tailwind reader/cache/key/listeners. Other workers' symbols remain their own deliverables. No inference that a rule is helper-ready is presented as a rule executing natively.
