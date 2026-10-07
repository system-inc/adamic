Built: the existing wave 08 branch rebased onto main e8ba3d5d; six complete ports and three tested React partials preserved.
Commits: tested rebased implementation tip f4b81ff7; this report accompanies the landing evidence commit.
Checks: all six production oracle/mutant gates, bridge sanitizers/ownership, React supported subset and kernels green on the new compiler.
Mutants: every completed rule, raw fact/state and ownership mutant caught again; Globals and both partial kernel mutants caught again.
Uncovered: JSX parser and native React HIR/SSA still block full React parity; no new claims, no full repository gate.

Only codex/typeaware-wave-08 was pushed by this unit. It was not on main and
was not based on current main, so landing preparation took precedence. The first
conflict-free rebase used e011f8f6. During its oracle run, main advanced with
compiler changes; a second conflict-free rebase used
`e8ba3d5d81de4d3773c723914fccd4c76248b965`. All native profiles were rebuilt and
validated again with that compiler. Bridge/rule/cohere sources were unchanged
between those main tips, so the new checker archives built during the first
rebase were reused. The final bridge gate rebuilt its own native probes against
the second compiler. The report preserves the exact tested base rather than
claiming a continuously moving branch head.

## Final gates

The original `go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v
-timeout 30m` passed in 298.153 s with corpus environment variables set. Controls
have 31 findings / 17,816 bytes; compiler 77 has 28 / 20,884 bytes; repository
287 has two / 19,558 bytes. Normal and ASan/UBSan output matches unmodified Go.
All three rule mutants compile, exit zero with empty stderr, and differ only
under byte comparison: loop first byte 61, require 7483, cycle 12425. The retained
released registry mutant exits zero instead of the expected panic 70. Ambient
symbol flags, type-node classification and module-target mutations compile and
fail direct checker comparisons.

`validate_process.py` passed all 100 upstream programs (129 process-exit and 21
blocking findings), compiler 77 and repository 287, normal and ASan/UBSan. The
callee mutant is caught in six programs, first byte 164; blocking order in 19,
first byte 1690. Both compile and exit zero with empty stderr. Every new raw
question rejects released handles before output, exit 70. `validate.py` passed
race controls 22 / 13 findings / 8,394 bytes and both corpora normal/sanitized;
its compiling lost-timeout mutant is caught by byte comparison. Paths make
control-stream lengths and first differing bytes differ from earlier runs.

`wave08-react/validate.py` passed Globals controls 35 / 34 findings / 23,175
bytes and both corpora normal/sanitized. Its compiling ancestry mutant is caught
at byte 9252. The JSX witness still receives a Go finding and native parser
refusal before output. `validate_cores.py` passed 52 sanitized results against
unmodified Go kernels: 36 lattice joins, six mutation decisions and ten derived
effects. Join and dependency-count mutants compile and exit zero, then fail Go
kernel byte comparisons. Both unsupported production entry points refuse with
exit 70. These two are kernel proofs, not completed production rule mutants.

`validate_pending.py` passed raw returned-flags, resolved-target and ancestor-name
mutants against direct checker witnesses, native event expectations under
sanitizers, and compiling write/blocking-state mutants caught by Go expectations.
`GOFLAGS=-overlay=/workspace/wave08-landing/bridge/overlay.json go test
./bridge/tsgo ./bridge/tsgo/checker -count=1 -v -timeout 30m` passed: bridge
177.283 s, checker 0.412 s. This includes 100 ABI queries, retained outputs,
stale/zero handles, 162-position oracle and ASan/UBSan/LSan. Input/output-length
mutants trigger ASan; retained stale handle triggers assertion; wrong position
fails byte 6; removed link guard fails refusal; omitted C free and incorrect
region ownership trigger LeakSanitizer. The owned Go raw package, vet and
formatting checks passed. Filtered `TestTheOracleCatchesOneByte` passed in
0.788 s using a fresh native/Node cache miss.

## Timing and scope

After concurrent validation completed, three alternating fresh-process rounds
measured process profile compiler native 4.231385 s versus Go 0.483251 s
(8.756x), repository native 0.634040 s versus Go 0.198621 s (3.192x). Native
remains slower for these end-to-end runs. Setup timing remains the earlier
132 s, nproc 5 with a four-core quota; setup was not rerun for the rebase.

The JSX and native React HIR/SSA dependencies remain unavailable. Partial ports
and their exact limitations are documented in REPORT.md. No shared infrastructure
was edited and no new rules were claimed. Full React findings/fixes/suggestions
parity is not claimed. All implementation files remain .a.

Two validation invocation mistakes are recorded: a combined owned-package and
bridge overlay run hit a test-only import cycle, fixed by running packages
separately; an older 86-case process directory lacked the callee mutant's positive
witnesses, fixed by rerunning the complete 100-case directory. An accidentally
broad fixture-count run was stopped and replaced by the intended one-byte filter.
Those invocations are not passing gates. Final passing runs and compressed raw
streams are in landing-validation/. Scripts are unchanged and reproduce using
the argument forms in the earlier reports, with the rebuilt stage0 and the
complete /workspace/wave08-process-full fixture capture. No full repository Go
gate was run; touched packages plus the filtered independent oracle were run.
