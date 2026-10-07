Built: the ruled per-member discriminant-write check, shared by both backends, with exact narrowing and existing confinement proofs.
Commits: initial blanket implementation f0d5de7, main merge 6ad16bb (e8ba3d5); final feature SHA is reported with the branch push.
Commands and outputs: setup done 140s, nproc 5; touched packages, filtered uncached oracle, counts and go vet passed.
Mutants: skip, first member only, every holder fresh, missed compound, missed base view, missed constructor update all caught by refusal assertions.
Not covered: full gate completion, runtime enum syntax, whole-program construction classification of the additional 71 census events.

No push or merge to main or any area branch was performed. Main was merged into
codex/discriminant-writes. The final fetch found origin/main unchanged at e8ba3d5.

Setup: bash cloud/setup.sh logged Go ready 0s, clang ready 1s, Node ready 1s,
submodules 1s, compiler warm 140s, done 140s. nproc reported 5; the cgroup CPU
quota is 4. Environment: /workspace/adamic-tools/env.sh; Node 24.19.0,
clang 20.1.8, Go 1.27.1.

Observed checks:

- `go test ./internal/lower ./internal/fresh ./internal/ir -count=1`:
  PASS, 48.298s / 59.051s / 27.616s. Log /tmp/discriminant-ruled-packages.log.
- `go vet ./...`: exit 0, empty log /tmp/discriminant-ruled-vet.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestDiscriminant|TestFreshWriteProbesStayRefused|TestCountsAreRecorded' -count=1 -v`:
  PASS 48.901s. Log /tmp/discriminant-ruled-final-oracle.log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/discriminant_construction' -count=1 -v`:
  PASS 1.035s. Log /tmp/discriminant-ruled-final-fixture.log.
- `gofmt -l cmd internal` and `git diff --check`: empty logs.
- Stock checker census: TypeScript 6.0.3, 77 compiler files, zero diagnostics,
  437 supplied events adjudicated. Exact reproducible command is in the census report.

The original witness is run on Node and observed to print `branch undefined`.
Its exact discriminant refusal wins before main's separate type-predicate
refusal. Every accepted executable witness is compared byte for byte with Node
using release native, sanitized native and the JavaScript backend, and is leak
checked. Required enum domains come from an ambient declaration file written as
.a and served through the existing virtual .a.ts alias; runtime enum declarations
remain independently unsupported. No missing-field runtime code was edited.

The reproducible mutant driver is `python3 notes/discriminant-writes/mutants.py`.
It writes Go overlays, never modifies the live compiler, and validates that the
killer is the test's missing-refusal assertion, not a compilation failure.

| Mutant | Killer | Observed result |
| --- | --- | --- |
| skip both admission checks | TestDiscriminantWrites/union | exit 1; expected refusal, got nil |
| examine only first possible member | TestDiscriminantWrites/union | exit 1; expected refusal, got nil |
| treat provisional holders as fresh | TestDiscriminantWrites/this_escapes_constructor | exit 1; expected refusal, got nil |
| omit wider result types of compound updates | TestDiscriminantWrites/compound | exit 1; expected refusal, got nil |
| omit non-union base views | TestDiscriminantWrites/base_interface | exit 1; expected refusal, got nil |
| skip constructor declared-domain update check | TestDiscriminantCompoundConstruction | exit 1; expected refusal, got nil |

The across-variant probe calls change on a Leaf and writes branch. The checker
orders the accepted branch member first in that probe; the first-only mutant
makes it compile. The predicate-bearing original witness still has main's
independent predicate refusal if the entire discriminant check is removed, so
that version cannot honestly be claimed to compile on current main. The
predicate-free write probe isolates the exact admission defect.

The earlier full uncached gate was stopped after about 640 seconds while
bridge/tsgo was still running. Its log reports signal interrupt; it is not a
passing gate. The allowed worker fallback is the touched packages and filtered
oracle above. Integration still owns the full uncached gate before main moves.

Census observations and limits: 430 accepted / 7 refused among the supplied 437;
40 accepted / 7 refused among the 47 outside local construction; plain blanket
rule refuses 32 of those 47 when restricted to the exact narrowing predicate.
All 79 kind events are inside construction. Two events of uncertain construction
are assignability-safe. The independent JSON provides candidates and reviewed
local construction, not final ruled verdicts. Every supplied event and every old
30-site event is explained in typescript-writes.md. An independent traversal of
all explicit assignment targets additionally finds 71 structural-view events,
45 safe and 26 requiring freshness or refusal. The larger site sets therefore do
not agree. The stock checker cannot establish their whole-program escape status,
and the compiler source is outside Adamic 0.1's accepted syntax; we have not
pretended to prove their confinement or silently dropped these differences.
