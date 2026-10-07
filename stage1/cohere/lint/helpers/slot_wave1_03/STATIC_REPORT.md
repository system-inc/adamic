Built: FrameworkStaticReading in one .a helper file with explicit declaration-body dependency.
Commits: pre-code reservation 994ff553, pushed after Theme.Add implementation e3924eb5; final implementation SHA appears in the final response.
Commands and outputs: 6,239 cases and 1,085,222 identical Go/Node/emitted-JavaScript/sanitized-native bytes; owned test PASS 60.420s.
Mutants: missing registration marked present, count incremented, order dropped; all compile/run normally and are caught only by comparison on all three backends.
Not covered: production declaration converter/property-sort integration, whole-rule findings/fixes, absent external repository corpora and the full repository gate.

# Static utility reading handoff

The second reservation was pushed before code after a wildcard fetch of every
origin head and inspection of seventeen distinct helper claim blobs. This helper
ties the largest remaining concrete consumer count at six, excluding the already
implemented comments bundle. Theme.Add and its evidence were pushed first.

`frameworkStaticReading(name, registrations, readBody)` distinguishes missing
registrations from present empty declaration lists. Missing names do not invoke
readBody. Present names invoke it once with the original declaration list and
forward the complete order/count result. `orderNil` explicitly preserves the Go
slice's nil state, including empty-body controls. The typed dependency composes
the separately owned nodesFromStaticDeclarations and PropertySort helpers. The
registration table is input data, not a table of final readings or guessed values.

The owned Go overlay calls original pinned FrameworkStaticReading, capturing its
actual body-dependency calls, exact arguments and resulting order/count. The
independent dependency input comes from Go PropertySort(nodesFromStaticDeclarations)
called separately; it is not the expected final FrameworkStaticReading result.
The mutable call trace is enabled only in the single-threaded oracle process,
not while parallel Go consumer suites capture names. No cohere source worktree,
shared harness, generator or compiler file is changed.

`python3 .../slot_wave1_03/validate-static.py --scratch /tmp/adamic-helper-wave1-03-static-final > evidence/static-final.log 2>&1`
passes. Six actual consuming rule fixture suites each exercise all 890 framework
registrations through live design-system construction. This produces 5,340
consumer helper cases; 899 controls cover every registered utility again plus
missing, case/whitespace-sensitive, NUL, Unicode and present-empty inputs.
6,239 cases produce 1,085,222 identical bytes on original Go, source Node, emitted
JavaScript and ASan/UBSan native, all successful paths exit 0 with empty stderr.
Canonical hash is b43aeb430c3870f1b237fc4022b6fb61a12721ac10ec8d93dc38cb13c3d86440.
Raw consumer logs, corpus, canonical Go stdout and independent output hashes are
under evidence/static. Tailwind 4.3.3 and the fixture package-root scratch overlays
are the same as the first helper. Go cohere pin is 715ba94f3608a6500086b1076ce5cb7e51b836db.

`go test ./stage1/cohere/lint/helpers/slot_wave1_03 -run '^TestFrameworkStaticReadingMatchesCohere$' -count=1 -v -timeout=20m > evidence/static-package.log 2>&1`
passes in 60.420s. The finalized trace guard is additionally validated by the
standalone final run. First helper package evidence remains PASS 31.983s. The
repository-wide vet and filtered uncached Node input oracle remain recorded in
the first report; a final vet run covers the new owned test as well. Setup 74s,
nproc 5, Go 1.27.1, clang 20.1.8, Node 24.19.0.

All mutants compile and run successfully before comparison: missing-found first
differs at line 6231; reading-count and reading-order first differ at line 1.
Each difference is caught on source Node, emitted JavaScript and sanitized native,
with no compiler failure, panic, stderr or sanitizer finding credited as a catch.

## Rules losing one dependency each

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Six dependency entries removed by this helper, twelve with Theme.Add across these
same six distinct rules. Zero final helper blockers removed. No rule is claimed
ported or fully integrated. static-readiness.json retains every residual dependency;
no shared readiness ledger is changed.

No arbitrary malformed registration table, invalid integer order/count, raw invalid
UTF-8, isolated UTF-16 surrogates, mutation of registration lists during callbacks,
full cross-slot callback integration, unavailable external repository fixtures,
whole-rule findings/fixes or complete repository go test gate is claimed. Nil
versus empty list state is observed, but Go backing-slice capacity is not modeled.
