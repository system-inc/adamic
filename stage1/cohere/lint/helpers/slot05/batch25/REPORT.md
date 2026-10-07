Built ParseRegexFlags, RegexFlags.UV and PatternAndFlags in separate .a files; nine blocked dependency occurrences removed across three rules.
Reservation 56d10c693dba9faaf6899908a8d688af036261b6 pushed before source; final implementation SHA appears in the final response.
Batch25 PASS 23.227s, seven uncached Node probes PASS 1.300s, vet/format clean; setup 25.842s, nproc 5.
All twelve compiling native semantic mutants caught by actual Go outputs; substitutions and first differences in evidence/mutants.json.
Not covered: full repository gate, seventeen required stage 1 external-input checks, whole-rule findings, regex engines or arbitrary byte-array values outside 0..255.

## Contracts and readiness

parse_regex_flags.a takes original Go string bytes and detects ASCII u and v independently. It accepts duplicate flags and unknown bytes exactly as Go does; it does not validate a regex flag set. regex_flags_uv.a returns the OR of the two fields, including the both-true case. pattern_and_flags.a requires at least two bytes and leading slash, splits at the last later slash, and retains everything after the opening slash if no closer exists. UTF-8 bytes, including invalid sequences, are preserved verbatim; no finding positions are built here. flags.a is the data interface, main.a the test driver. Callers must supply original UTF-8 byte arrays (integer values 0..255). No matcher, RegExp engine or UTF-16 span conversion is introduced.

Each helper removes one prerequisite for no-control-regex, no-regex-spaces and no-useless-escape. No final blocker is removed by this trio alone and no rule is ported. Cumulative slot05 readiness: 71 helpers, 374 prerequisite occurrences, 73 unique consumers and 50 helper-ready rules under the frozen common AST adapter assumption. readiness.json records the exact mapping and calculation.

## Oracle and consumer coverage

The oracle overlay adds only a new executable inside pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and calls the actual three exported helpers unchanged. It reads every Go string literal in every inventory consumer's test files, including already-ported consumers outside the frozen blocked set. It then projects each literal to its last-slash suffix and includes raw strings as flag inputs. Prose/options/full sources are deliberately retained: this is helper projection coverage, not observed finding parity or AST parsing.

Flags and UV each capture 2206 distinct strings from eight consumers; pattern splitting captures 2316 from nine. Final baseline calls: flags 3505, UV 3509, pattern 3615, total 10629. Every byte occurs alone, embedded in flags, after an opening slash, inside slash-delimited text and in suffixes. Controls cover empty/open-only literals, absent closing slash, escaped slash, multiple slash, astral text, invalid UTF-8 and duplicate/case-sensitive flags. UV also tests all four possible Boolean field combinations directly. All consumer counts and source paths are retained in the per-mode coverage JSON.

The complete same-source baseline compares actual Go output byte-for-byte with source Node, emitted-JavaScript Node and ASan/UBSan native. The driver renders byte outputs as hex so neither malformed UTF-8 nor delimiters are lost. A variant must compile, exit zero and produce empty stderr before a Go-output mismatch earns credit. Native compile failures, panics or sanitizer-only failures cannot satisfy the mutation guard. Four ParseRegexFlags variants remove either flag or inspect only its first position; three UV variants replace OR with AND or either field alone; five PatternAndFlags variants remove the leading-slash guard, split at the first slash, discard the unterminated pattern, retain the opening slash or retain the closing slash in flags. All twelve are caught with exact independent witnesses. The initial frozen-only corpus also passed 14.177s; its log is preserved separately, superseded by the broader final run.

## Commands and limits

Each build shell sources /workspace/adamic-tools/env.sh. Test output is redirected directly to logs, never piped.

```
bash cloud/setup.sh > /tmp/lint05-batch25-setup.log 2>&1
ADAMIC_SLOT05_BATCH25_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch25/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch25 -count=1 -v -timeout=15m > /tmp/lint05-batch25-helpers.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch25-node.log 2>&1
go vet ./... > /tmp/lint05-batch25-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 > /tmp/lint05-batch25-format.log
```

Go 1.27.1, clang 20.1.8, Node 24.19.0, five processors (cgroup quota four cores). Exact setup timing lines are retained in evidence/setup.log. All selected tests PASS; seven uncached Node probe misses, empty static check logs. No shared registration, harness, compiler or upstream source was edited. Main 71d7e491 and area b28757f3 remained ancestors at final fetch; all twenty origin helper claim trees were reread and selected ownership was unique to slot05. Only codex/lint-helpers-05 is published; no main/area push or PR.

The previous 68 helpers remain complete with unchanged source. Their retained complete landing proof covers all 24 previous packages and 290 semantic variants; this batch adds twelve, not a rerun claim for the old packages. Full repository and seventeen required external-input tests were not run or credited; no skip or relaxation. No arbitrary corpus exhaustive proof, whole-rule diagnostic/fix parity, flags validation or new rule adapter is claimed.

## Current lint-area landing

See LANDING26.md for area e667e3e1, unchanged compiler/helper input proof and twelve fresh compiling semantic mutant witnesses. No new reservation.
