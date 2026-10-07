Built the CSS quote sub-printer and an exact file-to-stdout driver in `stage1/cohere/cssstrings`.
Go cohere, Prettier 3.9.6, native Adamic, Node source and the JavaScript backend agree on 18,674 cases.
All three output mutants were caught; native ASan, UBSan and leak checks passed.
One stage 0 gap is recorded with a proving program: multi-value Array.push.
Full gate passed; isolated throughput: native 357k, Go 555k, Node 124k, Prettier 114k texts/s.

## Choice and scope

I read CLAUDE.md, the language and memory documents, the whole GraphQL slice,
and the whole JSON slice on `origin/codex/stage1-json-format`, including their
layouts, tests, gap records and proving programs before implementing this slice.
The prior slices supply the overlay oracle, sanitizer, Node and backend pattern.

The formatter survey found YAML's CST/compose tree (51 files) and Markdown's
micromark/mdast pipeline (112 files). No TOML or ignore-file formatter was found
under `cohere/internal/format`; the existing stage 1 gitignore slice already covers
ignore matching. CSS has independently callable sub-printers. I chose its quote
printer because the whole dependency chain is four small string functions, with
no AST, TypeScript parser, document-layout engine or general regex runtime.
Fixed quote/escape scans implement the original patterns directly.

Ported functions: adjustStrings, printString, getPreferredQuote and makeString.
The driver formats a file with this sub-printer, preserving surrounding text.
It does not implement full CSS formatting. No compiler-owned files were edited.
`GAPS.md` records the contract, reproduction commands, workaround and limitations.
The copied execution helpers retain the earlier GraphQL slice's timeout/process
cleanup, lowering, sanitizer and leak checks; the new corpus and assertions are
independent.

## Pins and corpus

- Base: origin/main `fe3b9f236e0672e968bcb7ad53c1882cdde18f29`.
- Branch: `codex/stage1-formatter-slice`, created after fetching origin/main.
- Go cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- Its TypeScript submodule: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
- Original JS: npm `prettier@3.9.6`, installed in
  `/tmp/adamic-cssstrings-prettier`; the library driver verifies the version.

The recursive walk includes every case-insensitive `.css`, `.scss` and `.less`
file in the checkout and initialized submodules, skipping only `.git`.
There is one: cohere's
`internal/lint/rules/tailwind/tools/generate_descriptor_base/testdata/independent_theme.css`.
There are no physical SCSS or Less files in this checkout. There are 9,336
generated texts: every length zero through five over six characters (letter,
single quote, double quote, backslash, newline, emoji), plus CR/tab/Unicode-line
separator, 10,000 repeated quoted strings, unmatched terminal escape, preserved
unnecessary escape, and NUL cases. Both preferences yield 18,674 comparisons.

The retained escaped corpus `/tmp/cssstrings-cases.txt` has SHA-256
`838b44c76f6f0a519c3fecee67ecf046931c769a5a667190acfd080fdfb03be5`.
The `.files.json` sidecar names the walked files. The tests regenerate both.
All results are compared byte for byte, including each timed result.
Raw file stdout is separately checked on every repository file and on empty
input, no final newline, LF, CRLF/multiple newlines, Unicode and escapes, with
both preferences. The native raw runs enable leak detection too.

## Setup and commands

`bash cloud/setup.sh > /tmp/stage1-formatter-setup.log 2>&1` exited 0.
The environment file printed/created was `/workspace/adamic-tools/env.sh`.
Each toolchain command sources it. `nproc` printed **5**; the cgroup quota is four CPUs.

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
v24.19.0
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (16s)
setup: done in 16s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`npm install --prefix /tmp/adamic-cssstrings-prettier --save-exact prettier@3.9.6`
exited 0: `added 1 package in 771ms` (log `/tmp/cssstrings-npm.log`).

The first slice run failed with the deliberate, recoverable lowering refusal
`stage 0 can't lower push with other than one value yet`. The port was changed to
single-value pushes; `gaps/1_multi_push.ts` retains the minimal example and
`TestMultiPushGap` holds the exact NotYet against Node's `ab` output. This was an
implementation gap, not an observed formatter output discrepancy.

The standalone driver build:
`go run ./cmd/adamic build stage1/cohere/cssstrings/main.ts -o /tmp/cssstrings-driver`
exited 0, with no output (`/tmp/cssstrings-driver-build.log`). On the input
`[title='hello'] {}` plus LF it writes `[title="hello"] {}` plus LF.

## Mutants

Each mutant is made in a fresh scratch copy, lowers, builds with sanitizers,
finishes with exit 0 and no stderr, and is caught solely by its output against
Go cohere. Each is checked natively and on Node source. The originals are never
mutated on disk.

| Mutant | Change | Witness |
| --- | --- | --- |
| quote tie | choose alternate on `>=` instead of `>` | case 0, byte 1213, repository CSS file, double preference |
| escaped closing quote | stop skipping the character following a backslash | case 142, byte 2, both native and Node differ |
| preserve original escape | reprint even when the original enclosing quote wins | case 818, byte 2, both native and Node differ |

A post-format run initially failed the two whitespace-sensitive mutant-site guards,
not the byte comparisons. The anchors now name expressions rather than `if` spacing;
all three mutants passed their output-only checks in the subsequent run. An initial
`prefer-template` lint finding was also fixed before that run.

The test log records the differing output for each. No compile failure or crash
counts as a caught mutant.

## Verification and performance

The full gate exited 0. It began during implementation; after the final edits,
the complete slice was checked again against the pinned library and vet was run
again. No filtered oracle or reduced package gate was substituted.

```sh
git fetch origin && git checkout -b codex/stage1-formatter-slice origin/main
git fetch origin codex/stage1-json-format:refs/remotes/origin/codex/stage1-json-format
(cd cohere && go build -o /tmp/cssstrings-cohere ./command/cohere) > /tmp/cssstrings-cohere-build.log 2>&1
/tmp/cssstrings-cohere --no-cache --format-only stage1/cohere/cssstrings/main.ts stage1/cohere/cssstrings/strings.ts > /tmp/cssstrings-format.log 2>&1
/tmp/cssstrings-cohere --no-cache --no-fix --no-format stage1/cohere/cssstrings/main.ts stage1/cohere/cssstrings/strings.ts > /tmp/cssstrings-lint.log 2>&1
/tmp/cssstrings-cohere --no-cache --format-only --no-fix stage1/cohere/cssstrings/main.ts stage1/cohere/cssstrings/strings.ts > /tmp/cssstrings-format-check.log 2>&1
gofmt -l cmd internal stage1/cohere/cssstrings > /tmp/cssstrings-gofmt.log
go vet ./... > /tmp/cssstrings-vet.log 2>&1
go test -count=1 -timeout=30m ./... > /tmp/cssstrings-full-gate.log 2>&1
ADAMIC_CSSSTRINGS_LIBRARY=/tmp/adamic-cssstrings-prettier \
  ADAMIC_CSSSTRINGS_KEEP=/tmp/cssstrings-cases.txt \
  go test -count=1 -v -timeout=30m ./stage1/cohere/cssstrings > /tmp/cssstrings-final3-test.log 2>&1
```

Branch creation, the extra JSON branch fetch, both builds, final formatting,
type/lint check, formatting check, gofmt and vet all exited 0. Build, vet and
gofmt logs are empty. The final lint output reports 276 rules, 2 checked files,
and 100% Adamic-ready. Its warning that only 2 of 81 project files were checked
records the deliberately scoped source check. Both submodules are clean.

Complete full-gate output (the oracle took 919.582 seconds on this run):

```text
?   	github.com/system-inc/adamic/bench	[no test files]
?   	github.com/system-inc/adamic/cmd/adamic	[no test files]
?   	github.com/system-inc/adamic/cmd/adamic-fuzz	[no test files]
ok  	github.com/system-inc/adamic/internal/flow	223.985s
ok  	github.com/system-inc/adamic/internal/fresh	66.256s
ok  	github.com/system-inc/adamic/internal/fuzz	26.399s
?   	github.com/system-inc/adamic/internal/ir	[no test files]
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/internal/load	3.278s
ok  	github.com/system-inc/adamic/internal/lower	13.008s
ok  	github.com/system-inc/adamic/internal/native	423.586s
ok  	github.com/system-inc/adamic/internal/oracle	919.582s
ok  	github.com/system-inc/adamic/stage1/cohere/cssstrings	181.619s
ok  	github.com/system-inc/adamic/stage1/cohere/formatfiles	140.410s
ok  	github.com/system-inc/adamic/stage1/cohere/gitignore	218.888s
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	248.628s
ok  	github.com/system-inc/adamic/stage1/cohere/mediaquery	136.977s
ok  	github.com/system-inc/adamic/stage1/cohere/suppression	163.516s
ok  	github.com/system-inc/adamic/stage1/cohere/values	129.101s
```

Final slice output, including all mutant witnesses (the two lines for each are
native and Node respectively):

```text
=== RUN   TestMultiPushGap
=== PAUSE TestMultiPushGap
=== RUN   TestCSSStrings
=== PAUSE TestCSSStrings
=== CONT  TestMultiPushGap
=== CONT  TestCSSStrings
--- PASS: TestMultiPushGap (0.11s)
=== RUN   TestCSSStrings/quote_tie
    port_test.go:313: caught by output case 0, byte 1213: got "rpus-root <that repo's source tree> --json independent.json\\" want "rpus-root <that repo\"s source tree> --json independent.json\\"
    port_test.go:313: caught by output case 0, byte 1213: got "rpus-root <that repo's source tree> --json independent.json\\" want "rpus-root <that repo\"s source tree> --json independent.json\\"
=== RUN   TestCSSStrings/escaped_closing_quote
    port_test.go:313: caught by output case 142, byte 2: got "aa\"\\\\\"" want "aa'\\\\'"
    port_test.go:313: caught by output case 142, byte 2: got "aa\"\\\\\"" want "aa'\\\\'"
=== RUN   TestCSSStrings/preserve_original_escape
    port_test.go:313: caught by output case 818, byte 2: got "a'\"'" want "a'\\\\\"'"
    port_test.go:313: caught by output case 818, byte 2: got "a'\"'" want "a'\\\\\"'"
=== NAME  TestCSSStrings
    port_test.go:347: throughput Go 555329.4 texts/s, 3 process runs 0.100881s (startup, input, escaping, output included; build excluded)
    port_test.go:347: throughput native 356919.5 texts/s, 3 process runs 0.156960s (startup, input, escaping, output included; build excluded)
    port_test.go:347: throughput Node 124307.6 texts/s, 3 process runs 0.450672s (startup, input, escaping, output included; build excluded)
    port_test.go:347: throughput Prettier 113719.9 texts/s, 3 process runs 0.492632s (startup, input, escaping, output included; build excluded)
    port_test.go:349: corpus: 1 repository/submodule files, 9336 generated texts, 18674 preference cases; all byte-identical
--- PASS: TestCSSStrings (23.11s)
    --- PASS: TestCSSStrings/quote_tie (4.11s)
    --- PASS: TestCSSStrings/escaped_closing_quote (4.33s)
    --- PASS: TestCSSStrings/preserve_original_escape (4.16s)
PASS
ok  	github.com/system-inc/adamic/stage1/cohere/cssstrings	23.113s
```

| Side | Texts/s | Total seconds, 3 process runs |
| --- | ---: | ---: |
| Go cohere | 555,329.4 | 0.100881 |
| Adamic native, clang -O2, no sanitizers | 356,919.5 | 0.156960 |
| Adamic source on Node 24 | 124,307.6 | 0.450672 |
| Original Prettier 3.9.6 printer on Node 24 | 113,719.9 | 0.492632 |

Observed ratios: native is 0.64 times Go throughput, 2.87 times Node source,
and 3.14 times the original library. The final measurement ran after the full
gate finished. Earlier runs made during the gate showed substantial CPU
contention; they are not used for this table. Timing covers the identical 18,674
preference cases on each side, with no format refusals or output substitutions.
Three rounds are too few to claim a stable benchmark distribution.

## Commits and delivery

Implementation: `9fae91b62dcf1754781bf0d1fc064357baacf7a6`,
`Port CSS quote printing to Adamic`.
The report is a separate follow-up commit whose SHA is given in the final response.
`git push -u origin codex/stage1-formatter-slice` exited 0 and created the branch
on `https://github.com/system-inc/adamic.git`. The completed report is pushed to
the same branch. No pull request was opened.


## Not covered

Full CSS parsing/layout, number and unit printers, attribute flags, YAML,
Markdown, TOML, invalid UTF-8 exact bytes, and non-Linux raw stdout behavior were
not covered. The raw driver uses `/dev/stdout`, since the prelude has no raw
stdout primitive. No expected upstream differences were found on this corpus;
this is an observation, not a proof for all possible Unicode strings or CSS ASTs.
Whole-process mixed short-text throughput includes startup, file I/O, escaping
and stdout collection and excludes builds. It does not measure full-formatting
or steady-state printer throughput.
