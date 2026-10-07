Built: native post-dominance, unconditional blocks and control frontiers in .a, over supplied graphs for the two unfinished setter rules.
Commits: follows landing-ready 48355ed0 on main e8ba3d5d; the new control-analysis commit is named in the handoff.
Commands: setup 28s, nproc 5; 17796 graph controls including 28 Go-lowered source controls agree in Go, native, Node, emitted JavaScript and sanitizers.
Mutants: including throws as exits changes 3143 answers; dropping switch case tests changes 4575; both compile and exit 0 with empty stderr, caught only by bytes.
Not covered: native source-to-SSA, setter/ref checker integration, capture translation and full three-rule lint comparisons remain blocked; no additional claims.

The branch is still based on current main
`e8ba3d5d81de4d3773c723914fccd4c76248b965`. Before this continuation, local and
origin wave-29 both pointed at `48355ed0a18eb916b7836bd61ff2335440e2cbb4`.
The previous six complete ports and their available oracles remain unchanged
from LANDING_REPORT.md. This continuation touches only this worker's
`wave-29-third/` directory. It adds no claim and edits no shared harness,
registration generator, bridge dispatcher, parser, compiler or submodule.

The unchanged Go `postdominator.go` was read in full. Its runtime counterparts
are now in `post_dominators.a`: reverse postorder computed from the synthetic
return-only exit, Cooper-Harvey-Kennedy immediate post-dominators, the bounded
unconditional chain, the post-dominated predecessor walk, frontier membership,
and If/Branch/Switch test checks including non-default switch cases.
`control_core.a` runs those APIs on structural framed inputs. This is useful
rule infrastructure, not a lint entry or a source-to-SSA implementation.

Go's own no-return controls retain the entry block because its unconditional
walk inserts the entry before looking for an immediate link. Native reproduces
that observed behavior. Return/throw distinction, early exits, disconnected
blocks, cycles, sparse identifiers and absent query identifiers are all checked.

The independent Go oracle calls the unchanged public HIR `UnconditionalBlocks`
and `ControlDominators` directly, with no bridge imports. It enumerates every
predecessor-edge choice on one to three nodes, each nonempty outgoing set or
return/throw terminal, all possible entries and controlling-value subsets.
These are synthetic graph inputs to the algorithm and can include structures
broader than a particular concrete terminal's syntax. Six sparse-ID controls
cover goto and switch subject/case distinctions.

It also extracts 28 concrete source controls from Go's own
`postdominator_test.go`, rejects any source with parser diagnostics, and invokes
unchanged Go `Lower`. Their predecessor lists, terminal kinds and test identifier
IDs are exported as raw structural inputs, with empty, singleton and all-test
controlling sets. Their original Go functions are the oracle's subjects; Go
control results are never serialized as inputs to native. These larger real
lowered graphs supplement the exhaustive small-graph controls. The source graph
lowering in this test is Go, so it does not close the missing native lowering
entry or prove full rule fidelity.

There are 17796 total graph/controlling-set controls, 243119 identical bytes,
SHA-256 `84eb8fe79188f30fe7e1cad895b8c2f2fbde93cf5d6483746c2a2f7d6eaacd77`.
Go, native, the Adamic source on Node and emitted JavaScript on Node agree.
Native under AddressSanitizer, UndefinedBehaviorSanitizer and LeakSanitizer
also agrees, with halt-on-error enabled and empty stderr. All normal runs exit 0.
Two native implementation mutants compile, execute normally and have no stderr:

- Count Throw as well as Return as synthetic exits: comparison catches 3143
  changed graph results.
- Read only the first terminal test, dropping switch cases: comparison catches
  4575 changed graph results.

Only their independent answer bytes kill these mutations. These are control
analysis mutants, not full-rule mutants. The kernel uses no checker handle or C
bridge; new released-handle coverage is not claimed.

Stage 0 initially refused Array.from(Set) and uncontextualized empty fallback
arrays. The local code uses explicit typed empty arrays and unique ordered ID
arrays for returned memberships. All final checks run that version; no shared
compiler changes were made. The array membership scans can cost more than hash
membership on large graphs; the reported inputs do not establish compiler-corpus
lint performance or an asymptotic performance claim.

One observation: native framed-driver process 0.113135747s versus Go
oracle process 0.183838823s. Go's pure control calls total 47.990367ms in
this run. The process workloads differ: native reads frames, while Go builds
fixtures and writes frames plus expected output. These observations cannot stand
in for the requested native-versus-Go whole-lint timing. Full rule timings remain
unavailable until the missing source pipeline exists.

Fresh setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
cache warm 28s, total 28s, nproc 5, CPU quota 400000/100000 and 17.6 GB memory.
No full gate or completed-six-rule repeat was run; their sources and dependencies
were unchanged. This continuation runs the new control kernel's meaningful
comparison, all backends, mutants and sanitizer checks. Complete commands,
compressed input/output streams and empty error streams are under
`validation/control-core/`.

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-third/check_control_core.py \
  /workspace/wave29-control-final > /tmp/wave29-control-final.log 2>&1
```

All three full claimed React rules remain unfinished. Set-state-in-render still
needs native HIR lowering, capture translation, Dispatch alias facts and
compilation-unit/useMemo recognition wired to its validator. Set-state-in-effect
also needs memoization erasure/inlining and ref-value propagation wired to the
new control analysis. Static-components has its tested supplied-graph taint
kernel but no source graph entry. Published JSX parsing was previously verified
in scratch; this branch's shared parser is unchanged. Full findings, fixes,
suggestions, compiler/repository populations, full-rule mutants and lint timings
for these three rules are not claimed. No additional rules were taken.
