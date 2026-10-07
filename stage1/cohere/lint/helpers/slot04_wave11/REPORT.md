Built: raw-byte isDigit, isBracketed and hasAnyPrefix, one .a file per helper.
Commits: claim bd21674 pushed before code; based on current main e8ba3d5, after green pushed tip 2b43ded.
Checks: new helper package PASS 18.239s; six filtered uncached Node probes PASS 1.413s; vet and formatting clean.
Mutants: six compiling semantic mutants caught on source Node, sanitized native and emitted JavaScript; all four consumer omissions caught.
Not covered: full repository gate, whole-rule integration, invalid numeric byte adapters or arbitrary long prefix/content fuzzing; no additional final blockers removed.

## Behavior and consumers

The digit predicate recognizes exactly bytes 48 through 57. The bracket predicate
requires two bytes and literal first/last square brackets, without interpreting
interior content. Prefix matching compares raw bytes and includes Go's empty-
prefix behavior. This preserves invalid UTF-8 and embedded NUL without a decoded
JavaScript-string approximation. All input views are readonly. README.md states
the exact adapter boundary and nil/empty representation.

Each helper removes one prerequisite for each of these four consumers:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

This removes 12 frozen dependency entries and zero final blockers. Residual
helpers are recorded in local readiness.json; the original ledger and stage1
rule statuses remain unchanged. This is isolated helper behavior, not four
completed rule ports or findings parity.

## Claims and landing

All 18 origin codex/lint-helpers* branches were fetched and every claim path
inspected before reservation. These three tie the highest remaining unclaimed
concrete count, four each. All greater counts are reserved, including the
original comment bundle in the base worker's HELPERS.md. The branch was green,
pushed and based on current main before this claim; main remains e8ba3d5 after
the delivery refresh. No main or area branch is pushed.

Final refresh found two later overlapping reservations on slot 03, isDigit and
hasAnyPrefix: 4f5c75f at 04:07:41 UTC. Our bd21674 is 04:07:40 UTC. Under the
established earliest-claim rule, slot 04 retains both. isBracketed has no competing
claim. These identities are recorded for integration to reconcile duplicate work.

## Commands and observations

With /workspace/adamic-tools/env.sh sourced, test output went directly to logs.
Earlier setup in this session passed: Go 0s, clang 1s, Node 1s, submodules 2s,
warm 165s, total 165s; nproc 5. Setup was not repeated.

```
python3 stage1/cohere/lint/helpers/slot04_wave11/testdata/capture.py
go test ./stage1/cohere/lint/helpers/slot04_wave11 -count=1 -v -timeout=15m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m
go vet ./stage1/cohere/lint/helpers/slot04_wave11/...
gofmt -l stage1/cohere/lint/helpers/slot04_wave11
git diff --check
```

The real Go Tailwind rule suite passed using Tailwind 4.3.3 and a temporary
capture overlay. It recorded 112 unique asserted sources from all four rules.
Source bytes and actual decoded parser string/template fields provide 476
observations, with empty-prefix and ordinary-prefix lists. The separate helper
entry capture observed 18 digit calls, three distinct single-byte inputs. It
observed zero isBracketed or hasAnyPrefix calls. The capture is limited evidence;
no claim is made that these two predicates executed in the consumer suites.

The sweep checks all 65,793 byte inputs of lengths zero through two. Each input
is tested against five prefix lists: none, empty, [, [], and 0/255. It produces
328,965 observation lines. Fourteen shape controls cover empty data and prefixes,
exact and too-long prefixes, later matching prefixes, mismatches, long inputs,
Unicode bytes, embedded NUL/255 and nested bracket interiors. The two digit
observations on each nonempty case inspect the first and last byte.

All observations match real private Go helpers on source Node, ASan/UBSan native
and emitted JavaScript. Every successful process exits zero with no stderr;
Linux leak checking remains enabled by default. The package passes in 18.239s.
The filtered oracle bypasses probe cache, zero hits and six misses; exact timing
is in evidence/oracle.log. Vet and format logs are empty. No superseded failures
occurred, and no compiler fixes were needed.

## Every mutant and what caught it

Each mutant compiles and runs successfully on all three Adamic execution paths,
then differs from Go on a semantic witness. Compiler errors, stderr or sanitizer
failures are not credited as mutant catches.

| File | Mutation | Witness |
| --- | --- | --- |
| digit.a | Exclude lower endpoint 48 | Leading ASCII 0 in digit-run control |
| digit.a | Exclude upper endpoint 57 | Trailing ASCII 9 in digit-run control |
| bracketed.a | Require length at least three | Go accepts [] |
| bracketed.a | Require ] as opening byte | Go accepts [] and longer bracketed words |
| any_prefix.a | Skip equal-length prefixes | Empty prefix on empty value or exact whole-value prefix |
| any_prefix.a | Accept mismatches instead of matches | Exact prefix and nonmatching prefix controls diverge |

The readiness-based coverage test removes all fixtures for each of the four
consumers separately; all four omissions are caught. Semantic mutants operate
on temporary copies and leave production source unchanged.

## Limits and inference

Observed: the byte-domain helper outputs match Go for the bounded sweep and
captured/source/shape cases. Inference: these prerequisite entries can be removed
once callers adapt their data to the stated byte views. Whole engine and rule
integration remain separate. Actual runtime capture covers only isDigit; source
and controls supply the other predicates' evidence. Arbitrarily long arrays and
prefix lists, corrupt byte views, fractional numbers and whole-rule reports,
fixes or suggestions are not tested. The full repository gate was not run.
Previously retained helper code and its landing evidence are unchanged.
