# Refusal probes

`go run ./cmd/adamic-refusals -seed 1 -count 2000 -out /tmp/refusal-findings`
checks each program's repaired neighbor with `load.Load`, `lower.Lower` and `native.C`,
then requires the refused program to produce its catalog's particular `lower.Refused`
diagnostic. Exit 0 means all generated probes passed; 1 means findings; 2 means an
invalid command or incomplete syntax refusal catalog. Findings contain the complete
source, neighbor, seed, index, surrounding and expected diagnostic. `-out` also saves
finding programs as `.a`. Timing goes to stderr, independently of deterministic JSON
answers on stdout. Temporary diagnostic paths are normalized to `main.a`.

The compiler tested is the one linked into the command. Build the command inside a
target worktree to test that compiler; `-root` supplies only the source for the catalog
audit, and must refer to the same target. `-only definite-local,definite-field` and
`-only ts-ignore,ts-expect-error` restrict regression measurements to the affected
constructs. `-stop` ends at the first finding. `-catalog` prints the whole catalog.

No result cache is added. `ADAMIC_GATE_UNCACHED=1` and the default mode take the same
path and produce identical JSON. Go's build cache is separate from probe answers.

The catalog's executable entries cycle before any repeat. Seed and index vary top
level, functions, methods, closures, generic functions, nested blocks and class field
closures, with and without imports and with independent accepted padding. Module-only
syntax stays at module scope. Checking pragmas stay before code, where the checker
recognizes them. Every generated control in the first 200 programs is tested.

The catalog has explicit boundaries, also emitted in every run's summary. This is
not full refusal coverage. `yield` needs a generator, which is refused first; strict
modules reject `with`; `erasableSyntaxOnly` rejects enums and parameter properties
before the lowering pass can return Refused. The older 0.1 document also lists
features subsequently opened (exceptions, inheritance, accessors, collection/library
operations and input). It is impossible to require current main to refuse all those
as otherwise accepted single-construct programs. Multi-module rules, polymorphic
recursion, remaining library omissions and the expanded inheritance override rules
still need dedicated scenes. Those boundaries are not counted as successful probes.

The source audit parses both refusal maps and every direct Refused What expression
in refusals.go, plus its error-returning helper calls. A new map entry, direct refusal
or helper without a catalog owner fails loudly. The delegated helpers' internal
refusals are not exhaustively enumerated by this audit.

The diagnostic test is held by an actual mutant: replacing the match with acceptance
of any Refused diagnostic must fail `TestWrongDiagnosticIsAFinding/wrong` and
`TestStrictDiagnosticWithRealLowering`. This tool measures compiler policy, including
refusals in unused declarations. An accepted program demonstrates missing refusal
coverage; proving a runtime type lie requires a separate execution probe. NotYet
findings demonstrate a diagnostic-contract gap and are not silent compilations.
