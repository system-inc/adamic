Built native default-mode jsx-fragments, jsx-no-undef and no-array-index-key with numeric listener dispatch.
Implementation 02df7420 rebased onto main c01907a7; claims 1902550c/cd559736 were pushed before code.
New oracle suite PASS: 229 parse-clean controls, 305 findings, 287 repository and 77 compiler roots, normal and sanitized.
Three rule mutants and one raw-fact mutant run normally and fail only the Go-byte comparison; released-registry mutant fails panic-70 requirement.
Limits: four React analysis claims parked; non-default options and emitted-JavaScript lint execution uncovered; full snapshot decoding remains slow.

The three active reservations are complete at production defaults. All implementations
are .a and live in this worker's rule directories. Numeric rule.json kinds are
verified against unchanged Go cohere registrations. The isolated kind-indexed driver
passes decoded nodes directly to relevant listeners. It dispatches the index-key
SourceFile listener once; that listener owns the same recursive enter/exit iterator
stack as the production Go rule. No kind is read or compared as a string in a rule.

One raw bridge question, numeric-syntax-bindings, lives in a new checker file and a
new native decoder file. The shared checker dispatch change is one registration
line. The raw wire carries numeric parser nodes, byte positions, token starts,
structural roles, and symbol declaration syntax. No React predicates, diagnostic
verdicts or lint fixes are implemented in Go. New JSX controls are parsed by the
pinned checker, so these ports require neither the unfinished native JSX parser
nor the parked source-to-HIR/SSA/capture pipeline. Shared generators, harnesses,
parser and protected compiler files are untouched.

Original claim commits 1902550c and cd559736 were pushed before any implementation.
The context-value claim was subsequently parked on native return evaluation,
alias escape and captured-reference classification. A fresh all-origin audit
found no-adjacent-inline-elements reserved elsewhere and skipped it; index-key
was the first eligible replacement. Selection snapshots and reconstruction are
in selection/. The four parked claims are not counted as native ports.

The independent oracle imports unchanged Go production rules inside the pinned
cohere module through a Go build overlay. Its own loader and numeric listener
walk do not import bridge code. Sources are ten focused controls plus 219 literal
sources extracted from Go fixture tables; configured options in those tables
are not copied, and production defaults are passed on both sides. All 229 sources
parse cleanly. Native and Go agree on every serialized field: rule, message id,
message text, byte range, fixes and suggestions, including duplicated key findings
and Unicode/CRLF ranges. Counts are fragments 34, undefined tags 186, index keys 85.
All three production-default rules emit zero fixes and suggestions; this is tested
as output, not omitted serialization. Both frozen corpora have zero findings.

Normal and ASAN/UBSAN/LSAN builds agree on all controls and both corpora without
sanitizer diagnostics. The released-handle query panics with exit 70 and the exact
invalid-or-released-handle message. Retaining released registry entries lets the
probe exit 0, which the required ownership failure catches. All five new mutants
were compiled and run:

| Mutant | Change | What caught it |
| --- | --- | --- |
| fragment-attributes | attributes >0 to >1 | Exit 0, empty stderr; independent Go bytes differ at 961 |
| undefined-intrinsic | ASCII lower bound 97 to 98 | Exit 0, empty stderr; independent Go bytes differ at 1303 |
| index-name | compare index text with added ! | Exit 0, empty stderr; independent Go bytes differ at 2576 |
| raw-foreign-binding | replace declaration source path with queried file | Exit 0, empty stderr; independent Go bytes differ at 48 |
| released-registry | omit live-handle deletion | Probe exits 0; required exit-70 ownership check fails |

Two initial test-mutation setups were corrected before the final green run:
the raw-fact substitution initially left an unused Go local, and the released
probe initially appended an unsupported question argument. Neither is counted
as a successful semantic mutant. Final evidence is the c019 run.

Native versus Go standalone process wall times, same corpora and production
defaults, measured before concurrent legacy revalidation:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Repository, 287 roots | 0.7366s | 0.1259s | 5.85x |
| TypeScript compiler, 77 roots | 5.7986s | 0.3414s | 16.99x |

Native bridge queries are one per file: 287/77. Repository bridge load/query/run
are 0.0691/0.2798/0.6602s; compiler 0.2522/2.8056/5.5166s. Go repository
load/rule/run are 0.0691/0.00670/0.04543s; compiler 0.2309/0.05921/0.09214s.
These are observations of this standalone driver, not a speedup claim. Numeric
dispatch removes every-rule-per-node work, while generating and decoding a full
parser/binding snapshot remains expensive and needs later shared API work.

Final-base command for the new suite (all output logged, never piped):

```
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE_11_SEVENTH_ARTIFACTS=/tmp/wave-11-c019-seventh ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test ./stage1/cohere/typeaware -run '^TestWave11SeventhAgreementAndMutants$' -count=1 -timeout 15m -v
```

PASS in 90.799s. Logs, complete comparison outputs, mutation outputs and a compact
results.json are retained in validation-c019/. Setup reused the successful ready
0s/cache warm 86s/total 86s run; nproc is 5. The branch was cleanly rebased after
origin/main advanced from f8013f0b to c01907a7; its older fifteen rule ports and
metadata are revalidated separately on that final base before publication.

Production-default coverage is complete for the measured corpora and controls.
Fragment element mode and allowGlobals=true are not exposed by this standalone
entrypoint. The .cjs suffix exception is implemented but lacks a dedicated runtime
control; fixture sources are deliberately compared as .tsx under default settings. No end-to-end emitted-JavaScript execution of these lint rules is
claimed: that belongs to the shared lint harness worker. Existing older fifteen
rules retain their previously reported string-kind execution limitation pending
the shared handed-node API. Four analysis-dependent claims remain parked for
#dnv6f2c. There was no new shared Diagnostic commit named or landed in this base.


Final landing revalidation passed after the clean rebase onto c01907a7. All six
previous worker suites reran unchanged, including 641 positive control findings,
normal/sanitized corpus comparisons, all 35 existing worker mutant observations
and released-handle checks. Together with the three new rules this branch has
18 native production-default ports and 946 positive control findings. Four
analysis claims remain parked. No new claims were made after this batch.

| Earlier suite | Process seconds | Existing mutant witnesses |
| --- | ---: | --- |
| First | 111.273 | regexp, array, branches, signature, index, syntax; released registry |
| Next | 295.999 | collection, outcome, pure, lineage, awaited; released registry |
| Third | 105.300 | eval, extend, assign, global, anchor; released registry |
| Fourth | 91.677 | func, nonconstructor, wrappers, global; released registry |
| Fifth | 143.328 | throw, backreference, arrow-fix, provenance, regex-path, self-resolution, symbol provenance; released registry |
| Listeners | 34.344 | missing JSON kinds, wrong numeric kind, missing listener |

Every earlier semantic mutant exits 0 with empty stderr and is caught by complete
Go output bytes; released-registry mutants are caught by required panic 70.
Listener mutants are caught by independent production Go registration bytes.
Names and first differing bytes are in the retained c019-<suite>.log files.
All 40 worker mutant observations passed on this base; the filtered compiler
oracle separately proves its one-byte comparison can fail.

Full bridge tests passed in 81.269s, including ownership/sanitizer checks.
Package vet passed in 0.197s. The filtered original-Node/native/emitted-JavaScript
compiler oracle passed in 15.050s; it actually ran 55 fixtures (the Go hierarchical
filter admits substring matches) plus its one-byte mutant. This compiler coverage
does not claim emitted-JavaScript execution of the lint rules. Full argv vectors,
zero exits and elapsed times are in validation-c019/landing-results.json; its
revalidate.py launches independent tests with at most three concurrent jobs.
These legacy timings were under concurrency and are not isolated performance
benchmarks. The new-rule timings above were measured before that concurrency.

Only this worker's own branch is published. No push to main or area/* and no PR.
