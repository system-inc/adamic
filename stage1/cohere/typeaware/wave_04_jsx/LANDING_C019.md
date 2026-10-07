Built: rebased all 28 owned wave-04 patches onto current main c01907a7; no new rules claimed.
Commits: tested rebased code tip 08671580 replaces remote ef171d86; this landing-evidence commit follows.
Commands: original oracle PASS 220.695 s; continuation, all React/JSX partial validators, checker tests, scoped vet and filtered Node oracle PASS.
Mutants: all six completed-rule finding mutants, both released-registry mutants, and all existing partial kernel/report/refusal/listener mutants caught again.
Uncovered: full React/JSX source parity, numeric shared JSX integration, remaining binding/stability work and full repository gate.

The 28-patch range-diff records every owned patch unchanged. Fresh stage 0 was
built at `/workspace/typeaware-wave-04-landing-c019/adamic`. Only wave-04 is
published, guarded by an exact lease on old remote ef171d86. No main or area
branch is modified. This main update adds stage-3 work and an oracle hook;
the announced stage1/cohere macOS leak-helper diff has not landed here yet.
Nothing from that work was fought or reverted.

Original `TestWave04AgreementAndMutants` ran with both frozen manifests and
ADAMIC_TYPESCRIPT_SOURCE configured. Controls: 39 findings / 17,166 bytes;
repository: 14 / 23,598; compiler: 200 / 77,043. Native and sanitizer outputs
match Go. Casing end+1, matching-message and void end+1 mutants compile, exit
zero and differ only through Go comparison at bytes 5768, 123 and 1592.
Released handle exits 70; registry-retention mutant exits zero and fails the
required panic assertion.

Continuation `validate.py` used fresh --adamic, both frozen manifests and the
same process/blocking controls. PASS full normal and sanitizer findings/fixes/
suggestions protocol. Timer end+1, process end+1 and blocking end+1 mutants
compile and exit zero, caught only by Go bytes; the retained-registry mutant
is caught by the missing required exit-70 panic. Native compiler process
1.922383 s / Go 0.340923 s; repository 0.295035 s / Go 0.147670 s. Builds and
other validation ran concurrently, so these are single-run observations.

Fresh-compiler partial validators all PASS:

- React reporting: 11 / 5,948 bytes; six reporting/refusal mutants.
- Refs lattice: 1,297 / 55,984 bytes; nine comparison mutants.
- Refs environment: 451 / 9,943; six comparison mutants.
- Refs predicates: 180 / 3,136; four comparison mutants.
- Listener declarations: 426 bytes; nine numeric comparison mutants; whole
  module JavaScript emission retains its documented checker-link blocker.
- rule.json declarations: nine manifests / nine in-memory comparison mutants.
- JSX reporting/predicates: 56 / 4,941 bytes; three comparison mutants and
  three removed-source-refusal mutants.
- JSX numeric tag references: 42 nodes / 152 records / 2,045 bytes; five
  comparison mutants (receiver, module, template kind, root case and unwrap).

Kernel runs hold native, ASan/UBSan/leaks, source Node and emitted JavaScript
to independent production Go, with empty backend stderr. These are partial
kernel results, not complete source lint verdicts. Checker tests PASS 0.214 s;
scoped vet has empty output. Filtered Node oracle selects the one-byte check
and closures/method_closures/generic_functions fixtures, PASS 1.173 s.

Reproduction uses the prior landing commands with output prefix
`/workspace/typeaware-wave-04-landing-c019`, fresh --adamic for every Python
validator, and the same frozen corpus/control inputs. Exact logs, provenance,
range-diff and compressed backend/control/mutant streams are under
`evidence/landing-c019`. No full repository gate was run.

No rule regular expression was ported this round and no hand-rolled matcher
was added. Existing numeric rule.json listeners remain intact. Parked HIR
React claims remain reserved; the three JSX claims remain unfinished, with
source run(node) explicitly refused until numeric parser integration and
remaining binding/component/stability judgments are complete.
