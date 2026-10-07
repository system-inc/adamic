Built: six earlier complete ports; three React claims partially ported, with explicit production blockers.
Commits: earlier completion cfa7115d and evidence a116e599; React claim baa3228b; this report accompanies the partial implementation commit.
Checks: Globals 35 controls/34 findings, compiler 77 and repository 287 match Go bytes, normal and ASan/UBSan; 52 kernel results match Go.
Mutants: Globals declaration ancestry, immutability join and derivation dependency count compile and are caught by byte comparison; the latter two test kernels only.
Uncovered: JSX parsing and native React HIR/SSA pipelines block full parity for these three claims; no further claims taken.

The six completed ports and their full rule mutants, bridge ownership checks,
released-handle checks and timing are documented in [the original report](../WAVE08_REPORT.md)
and [the continuation report](../wave08-next/REPORT.md). All six were pushed
before selecting these three. The claim scan fetched 356 origin refs and checked
all claim Markdown blobs and full/shorthand registrations on main and the bridge
branch; the frozen selection and foundation inventory are in validation/.

## Globals

`globals.a` makes native AST/scope decisions using the previously implemented raw
symbol question. It handles compiler-name/evidence/parameter gates, reachable
root positions, assignment operator narrowing, lexical and abrupt boundaries,
shorthand symbol resolution, declaration ancestry, report ranges and messages.
No Go lint verdict is exposed. `suite.a` is an owned profile; shared harness,
registration and dispatcher sources were not edited.

`validate.py` compares the complete output stream against unmodified production
Go cohere, including the findings/fix/suggestion protocol. The 35 controls have
34 findings (22,860 bytes); compiler 77 has zero (5,780 bytes), repository 287
has zero (18,485 bytes). Both normal and ASan/UBSan output are identical. The
ancestry mutant changes the outside-declaration verdict, compiles, exits zero
with empty stderr, and differs first at byte 9171. The raw symbol question rejects
a released handle with exit 70 and no findings under the sanitized archive.

The source `let g=0;function Component(){g=1;return <div/>;}` produces a Go
finding, but the native shared parser refuses before findings: `parser slice
expected GreaterThanToken, got SlashToken at 44`. The parser treats this syntax
as a type assertion and lacks JSX expression parsing. Full Globals parity is
therefore **blocked**, not complete. The user prohibits shared parser edits.

One validation round measured native versus Go: controls 0.040012/0.062518 s,
compiler 3.421698/0.530553 s, repository 0.395202/0.197722 s. These are single
observations, not quiet alternating medians. Earlier six-rule measurements are
in the linked continuation report. Setup previously passed in 132 s; nproc 5,
CPU quota four cores.

## HIR-dependent rules

`immutability.a` ports the six-kind lattice merge, reason union and mutation
policy. It does **not** implement source lowering, SSA, alias/instruction
transfer, the inference sweep, reference exemptions or reporting.

`no_deriving_state_in_effects.a` ports a validation kernel over classified HIR
events: capture/dependency validation, inner-ID seeds, backedge exclusion,
allowed-instruction taint transfer, phi unions, setter argument/spread/count
checks, terminal abandonment and deferred report locations. It does **not**
implement source lowering, candidate gathering, hook/type binding, constant
folding, manual-memoization erasure or the reporting pipeline.

Neither has a production lint port. Their run entry points deliberately refuse
with exit 70 before emitting findings, identifying the unavailable pipeline.
An inventory across all 356 fetched origin refs found no native React HIR/SSA
foundation. Implementing that shared infrastructure exceeds the permitted rule
directories; exposing Go lint verdicts would not be a native port. Work stops here.

`validate_cores.py` overlays an owned Go test into the production React package
without modifying shared source. It calls unmodified private Go kernel functions
and a real typed Dispatch/setter witness. The sanitized native probe exactly
matches 36 lattice joins, six mutation decisions and ten derived-effect cases.
These include missing/extra captures, loops, unsupported stores, terminal reads,
argument counts, zero-taint abandonment, binary joins and phi unions. The join
mutant changes a MaybeFrozen merge to Frozen; the dependency-count mutant
reverses its guard. Both compile and exit zero with empty stderr, and the Go
kernel byte oracle catches both. These are **kernel mutants**, not the required
complete production-rule mutants; full findings/fixes/suggestions parity and
production-rule performance remain unverified for these two rules.

## Reproduction and evidence

Run `validate.py --help` and `validate_cores.py --help` for artifact, stage0,
archive and compiler-root arguments. All test output was captured to files.
`validation/` contains validation logs, single-round timing/results, source
hashes, frozen controls, claim/foundation inventories and compressed full output
streams. Native sources are `.a`. No shared production source was changed in
this continuation. The full bridge sanitizer/ownership gate passed for the
preceding completed ports; this continuation adds sanitized Globals/kernel
execution and the raw-symbol released-handle check, not another full gate run.

## Dependency refresh after the next continuation request

Fetched all origin heads again: 389 refs. A token-boundary filename inventory
finds only this branch's two partial kernels for React HIR/SSA/immutability/
derivation candidates. Main, the bridge branch and the harness branch still
have zero `parseJsx` entries in their shared parser. The previously built
Globals JSX probe and both production refusal probes were rerun: all exit 70
with empty output and the same documented reasons. This refresh does not
rebuild against newer branches or claim exhaustive semantic source inspection.
The dependencies remain unavailable; no new rules are claimed. Frozen branch
commits, inventory and refusal results are in validation/dependency-refresh.json.
