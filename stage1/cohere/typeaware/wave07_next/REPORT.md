Built: prefer-rest-params, prefer-promise-reject-errors and prefer-regex-literals in native .a; regex production dispatch registration remains pending.
Commits: claim 94a733f6046a2fbeda3416b6eed0e20f671e04a4; implementation 8323db46398a99b2137a173f3841d90fefc6ecaf; evidence accompanies this report.
Commands and outputs: three byte gates, ASan/UBSan/LSan, checker/scanner tests, vet and 11-file source lint pass; controls have 9/22/41 findings.
Mutants: all three rule-message mutants and the raw regex-character mutant caught only by Go bytes; registry-retention mutant caught by released-handle refusal.
Not covered: regex production registration, full repository gate, all upstream options/JSX fixtures, emitted JavaScript comparison and previous-wave gate reruns.

# Second continuation

The earlier six wave-07 ports and their evidence were pushed before taking this
trio. A fresh all-head fetch inspected 356 origin refs. Before the new claim,
33 distinct Markdown claim blobs named 132 ranked rules. Selection combines the
197 compiler/repository checker-dependent counts linked from VOLUME_REPORT.md,
orders by descending total and lexical ties, and excludes existing ports and
claims. Original port inventory comes from the production oracle suites, including
abbreviated native filenames and registry aliases. Its 26 ports include
method-signature-style, which is outside the 197 checker-dependent rows.
The claim was pushed before code. The reproducible read-only audit is
[selection.py](selection.py); [selection evidence](evidence/selection.json.gz)
contains fetched ref SHAs, blob IDs, ranking and exclusions, including this claim.
No additional rules were claimed in this turn.

## Implementation

Each rule has a separate .a file and native driver. Rest parameters checks the
implicit binding's live symbol identity and zero declarations, exempting dotted
property reads. Promise rejection checks global Promise origins, callback symbol
identity, duplicate parameter names, nested closures, parameter defaults and
syntax that could yield an Error, with TypeScript wrappers kept transparent.
Both use existing node-symbol-details frames and run against the normal archive.

Regex literals follows global references through aliases, destructuring, logical
and conditional expressions, assignments, parentheses and global objects. It
recognizes cooked strings, templates and genuine String.raw, then builds complete
suggestions in Adamic. Native checks comments, preceding tokens, flags, printable
patterns and group/quantifier balance, and constructs character escapes and
adjacency padding. Only default options are implemented, including the default
false disallowRedundantWrapping and allowEmptyReject.

The one new bridge question exposes generic regex scanner character spans and
escape/class boundaries. It does not return findings, rewritability or edits.
The scanner helpers and their tests are attributed copies of pinned cohere's
regexpattern and regexsyntax packages with their upstream MIT license. The
independent finding oracle invokes production Run methods unchanged and imports
no bridge implementation. Its scanner lineage is shared with the raw helper, so
this is an independent rule oracle, not an independent regex grammar oracle.

The shared dispatcher and existing shared harness/generator remain untouched.
Private overlays register the regex question for validation. Normal archive
refusal is measured: an unregistered regex request exits 70 before output.
Integration requires the one case shown in [QUESTIONS.md](QUESTIONS.md).
Consequently the regex implementation is validated in isolation and ready for
that integration, rather than registered in the production archive. The first
continuation's three pending routes remain documented in its own report.

## Gates and observations

All shell commands source /workspace/adamic-tools/env.sh and redirect output to
files. Exact build/test commands and exits are preserved in compressed
commands.json files. The rule gates are:

```sh
python3 stage1/cohere/typeaware/wave07_next/validate_rest.py /workspace/wave-07-next-rest --compiler /workspace/wave-07-typescript
python3 stage1/cohere/typeaware/wave07_next/validate_promise.py /workspace/wave-07-next-promise --compiler /workspace/wave-07-typescript
python3 stage1/cohere/typeaware/wave07_next/validate_regex.py /workspace/wave-07-next-regex --compiler /workspace/wave-07-typescript
python3 stage1/cohere/typeaware/wave07_next/validate_regex_question.py /workspace/wave-07-next-questions
go test ./bridge/tsgo/checker ./bridge/tsgo/wave07_regex/... -count=1 -v
go vet ./bridge/tsgo/checker ./bridge/tsgo/wave07_regex/...
```

The source lint command uses the isolated .a-capable cohere CLI with --no-cache
--no-fix over wave07_next/*.a. It reports 276 rules, 11 checked, zero findings.
Go formatting and vet logs are empty. Scanner tests include all copied table and
fuzz seed cases; no timed fuzz campaign was run.

| Rule | Controls | Findings | Bytes | Fixes | Suggestions |
| --- | ---: | ---: | ---: | ---: | ---: |
| rest | 18 | 9 | 3431 | 0 | 0 |
| promise | 40 | 22 | 8175 | 0 | 0 |
| regex | 53 | 41 | 24341 | 0 | 30 |

Each rule also agrees over the frozen 287-source repository corpus (18485 bytes)
and TypeScript's 77 src/compiler sources (5318 bytes), with zero findings in both.
The zero corpus counts are backed by positive and negative controls, not treated
as sufficient evidence alone. Every comparison preserves duplicate findings,
messages, byte spans, fixes and complete suggestions. DOM controls agree too.
Controls include shadowing, nested functions, shorthand reads, optional access,
callback identity, TypeScript assertions, Unicode/CRLF offsets, alias cycles,
comments, escapes, flags and invalid patterns. A merged String namespace control
pins Go's first-declaration origin test for String.raw; Promise uses all origins.

Normal and sanitized runs agree on controls and both corpora. Native/C archive
builds use address and undefined-behavior sanitizers with LeakSanitizer, and all
native stderr is empty. Go-managed heap internals are not C-sanitizer instrumented.
The initial class-inheritance build was refused by Adamic 0.1; composition replaced
it. Differential controls caught a spurious finding on `error as Error`, and
wrapper handling was corrected before final validation. These initial failures
are not counted as successful checks or mutants.

Each rule-message mutant compiles and exits 0 with empty stderr; only production
Go bytes catch it. The raw scanner mutant drops character spans, preserves valid
frames and exits normally; suggestion bytes differ first at byte 5854. Both the
new regex question and existing symbol question succeed live, then reject a
released handle with exit 70 and exactly `adamic: panic: invalid or released
checker handle`. A registry-retention mutant makes a post-release query exit 0;
the required panic assertion catches it. A scratch log filename containing `/`
initially prevented that gate from running; safe filenames fixed it and the final
gate passes. No compile failure is counted as a mutant kill.

## Timings and environment

Three final alternating rounds compare complete outputs, with no concurrent
builds or tests. The original full_bench.py records load/query/run phases and
process times; committed records preserve every round. Process medians in seconds:

| Rule | Repository native / Go | Compiler native / Go |
| --- | --- | --- |
| rest | 0.247672 / 0.130115 | 1.819633 / 0.346379 |
| promise | 0.228299 / 0.118278 | 1.605191 / 0.294662 |
| regex | 0.229748 / 0.125629 | 1.825777 / 0.329351 |

Native is slower on these zero-hit corpora. These measurements include program
loading and traversal and do not measure positive-control grammar query costs.
The setup from the initial wave remains in use: 81 seconds, Go 1.27.1, clang
20.1.8, Node 24.19.0. nproc again reports 5, with the existing four-core quota.
Pinned oracle cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db, its TypeScript-go
is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a and the compiler corpus is
050880ce59e30b356b686bd3144efe24f875ebc8. No pins changed. The newer CLI is used
only for .a implementation-source lint.

Evidence contains stdout/stderr streams, SHA-256 hashes, fixture snapshots,
manifests, command records, selection audit and final timing rounds. Absolute
paths are retained, so byte hashes describe this run rather than relocated runs.
No full repository gate, exhaustive upstream options/JSX matrix, emitted-JavaScript
comparison, full sanitizer buffer-mutant rerun or earlier-wave gate rerun was
performed. Previous reports retain those earlier measurements. Regex production
registration remains blocked by the shared-file restriction.
