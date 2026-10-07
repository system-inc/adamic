# Wave 17 duplicate reservation withdrawal

Built: no new production ports; three experimental native handlers were withdrawn after a competing claim was found.
Commits: claim 14713630; competing wave-11 claim 667e6a47; existing landing tip fdbab035 on main c01907a7.
Checks: experimental selected oracle PASS 108.263s, 309 controls, both frozen corpora, sanitizers and released handles.
Mutants: namespace byte 7794, danger byte 46, multi-component byte 2382; retained handle registry caught by required panic 70.
Limits: projected syntax only, no production parser integration; experimental sources retained outside the branch.

The explicit all-heads fetch inspected 583 origin refs and 33 distinct Markdown
claim blobs. Wave 11 claims react/no-danger-with-children, react/no-multi-comp
and react/no-namespace in 667e6a47 (commit timestamp 05:34:09 UTC), while our
claim 14713630 has timestamp 05:43:56 UTC. The first snapshot used for our
claim did not expose that competing reservation. These timestamps alone do not
prove push visibility ordering; the refreshed conflict is sufficient reason
to withdraw our duplicate reservation in favor of wave 11. No other worker's
branch or claim was edited, and no duplicate implementation is committed.

The current audit excludes ports on main/library and names in all origin
Markdown reservation files. It leaves zero unclaimed checker-dependent rules
among the 197 ranked entries. Selection reconstructs the combined descending
compiler/repository totals with lexical ties from VOLUME_REPORT.md's preserved
counts. The complete refs, claims and exclusions are in selection.json.gz.
There is no new claim after this correction.

Current main remains c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Existing wave-17
sources are unchanged from the previously validated landing: all eight
selected suites PASS 909.575s and the changed fifth-batch module PASS 171.661s.
See WAVE_17_LANDING3_REPORT.md. Shared model ab70f38d4 is visible on
origin/lint-rules/harness but is not yet an ancestor of main. This correction
does not change shared generators, harnesses, parsers, compiler code or the
other completed/parked scopes.

## Experimental evidence, not a production-port claim

Before discovering the collision, native .a handlers were implemented in owned
rule directories, with numeric SyntaxKind declarations and cached handed-node
callbacks. Their independent Go oracle calls the three unchanged production
rules and separately serializes raw numeric AST syntax into fixtures. Native
Adamic makes the decisions. The ordinary parser's string-kind model and absent
JSX support were not changed. The fixture includes token boundaries and named
AST fields; that comparison does not establish native parsing/position decoding
or general outside-file declaration traversal parity.

Final experiments, with fixture files confined to scratch storage:

| Population | Identical bytes | Findings | Native whole process | Go whole process |
| --- | ---: | ---: | ---: | ---: |
| 309 controls | 50136 | 193 | 0.052162s | 0.077088s |
| Controls, sanitizer | 50136 | 193 | 0.152393s | 0.082804s |
| Repository, 287 roots | 18485 | 0 | 0.424307s | 0.139102s |
| Repository, sanitizer | 18485 | 0 | 1.619082s | 0.144769s |
| Compiler, 77 roots | 5318 | 0 | 2.650576s | 0.337355s |
| Compiler, sanitizer | 5318 | 0 | 11.058715s | 0.365360s |

These are single process observations. Native timings exclude the external
AST fixture generation and cannot establish a production parsing speedup.
Native still takes 7.86 times Go on the projected compiler population. No
isolated phase benchmark or production-wide performance claim is made.

Findings retain complete message IDs/text, byte ranges and zero fix/suggestion
counts, as the three Go rules specify. The ignoreStateless option has 39859
identical bytes under sanitizers. Controls cover the upstream tests, Unicode
and CRLF, bare/imported/shadowed createElement, receiver parentheses, static
versus computed callees, cyclic prop spreads, first-child newline behavior,
class/factory/wrapper ordering, curried functions, null versus JSX returns,
assignment capitalization and the deliberately narrow return-statement walk.

All three mutants compile and finish with exit 0 and empty sanitizer stderr:
namespace suppresses reports (first difference 7794), danger suppresses reports
(46), and multi-comp wrongly exempts a second component (2382). Only the Go
byte comparison catches them. The live bridge's binding-declarations rejects a
released checker with exit 70 and the required panic text. A registry mutant
retaining the released program finishes cleanly, so the required panic catches
that mutant. Normal and sanitizer outputs agree on controls and both corpora.
The complete repository gate was not run.

The initial experiment passed in 126.122s; a receiver-parenthesis edge case
and scratch-only fixture transport were added, and the final suite passed in
108.263s. Unsupported namespace imports, imported constant switch cases and
an expression-position postfix increment were replaced inside owned files;
no compiler workaround was added. These failed builds are not counted as
valid mutants.

Sources and the dedicated test were removed from the working tree after being
saved as /workspace/wave-17-sixth/withdrawn-implementation.patch and under
/workspace/wave-17-sixth/withdrawn-sources. The patch SHA-256 and exact file list
are in status.json. Raw command stdout/stderr, fixture inputs and binaries are
also retained in that scratch directory. The committed evidence contains the
run logs and audit, not a competing production implementation.

Commands run, with test output redirected directly to log files:

```sh
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
python3 /workspace/wave-17-sixth/select.py
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript ADAMIC_WAVE17_SIXTH_ARTIFACTS=/workspace/wave-17-sixth go test ./stage1/cohere/typeaware -run '^TestWave17SixthAgreementAndMutants$' -count=1 -timeout 30m -v > /workspace/wave-17-sixth/test-final.log 2>&1
# PASS 108.263s, experimental files subsequently withdrawn.
```

The last setup for the landing succeeded: Go/clang/Node ready 0s, submodules
0s, cache warm 46s, done 46s; nproc 5 (cgroup quota 4 CPUs). This continuation
reused that setup and sourced the printed tool environment. No new setup
failure occurred. Only codex/typeaware-wave-17 is eligible for the push; main
and every area branch remain untouched.
