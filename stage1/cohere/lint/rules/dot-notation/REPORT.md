Built and pushed an owned .a compiler-gap probe; earlier rule work remains pushed.
Claim commit be10d93b reserves dot-notation, grouped-accessor-pairs and id-length.
Node prints true then false; native build and emitted-JavaScript lowering both refuse the runtime regex pattern.
No new rule mutant or findings/s measurement was run because no new rule implementation is certified.
The three new ports are pending; work stops at a compiler blocker under Ahra's instruction, without shared edits.

Selection fetched every origin head, checked 350 refs and 53 claim Markdown
files, and read the syntax-only inventory. Main was ef3d907e. All 46 original
helper-ready entries were taken. The reservation was pushed before the probe.
Earlier implementation and parity/mutant evidence through 6de3b402 was already
pushed; git push reported Everything up-to-date.

The minimal probe is gaps/dynamic-pattern.a. Go dot-notation compiles the
configured allowPattern through regexp.Compile at runtime. Go id-length does
the same for exceptionPatterns. Adamic's lowerer currently permits only
constant RegExp patterns. This observation establishes a compiler gap for
these option surfaces, not a shared reporting-harness gap. It does not claim
that grouped-accessor-pairs itself requires regexes. That rule is unstarted
because Ahra directed workers to stop on other blockers without editing shared
files. No placeholder rule.json or incomplete registered rule was added.

Commands, with output written directly to evidence logs:

- source /workspace/adamic-tools/env.sh
- GOFLAGS=-overlay=/tmp/wave02-fifth-final/overlay.json bash cloud/setup.sh
- node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/lint/rules/dot-notation/gaps/dynamic-pattern.a
- go run ./cmd/adamic build stage1/cohere/lint/rules/dot-notation/gaps/dynamic-pattern.a -o /tmp/wave02-sixth-pattern
- go run ./cmd/adamic js stage1/cohere/lint/rules/dot-notation/gaps/dynamic-pattern.a

Setup succeeded: Go, clang, Node and submodules each ready in 0s; cache warm
16s; total 16s; nproc 5; cgroup CPU quota 4. The existing owned scratch overlay
handles the previously documented .a registry limitation, without shared edits.
Node exits 0 with exactly true and false on separate lines. Both compiler
commands exit 1 and report at line 3 column 23:

    stage 0 can't lower RegExp with a nonconstant pattern yet

The initial shell attempted an unavailable adamic executable; the repository
CLI was used instead. The initial probe also passed booleans to console.log,
which Adamic's library rejects; the final probe uses string output and reaches
the actual lowering refusal. Only the corrected final evidence is attached.
No full gate, upstream rule comparison, source-corpus comparison, sanitizer
execution, emitted-JavaScript execution or new semantic mutant was performed.
Those requirements remain outstanding for the new reservations. Earlier owned
reports retain their explicit CFG, JSX, parser and reporting limitations.
