# k031ey0 census overlay evidence

Linux, Go 1.27.1. Implementation: 40b65423202a09d42aa5242eb0618b8ff07d9d49; parent: e64bc123c51fc333133684f0abb3bee8e522acf3. Exactly one implementation commit; all 15 changed paths are under stage3/census. Scratch branch was pushed immediately after the overlay build and guards passed.

## Successful checks (exit 0)

- `go test ./stage3/census/latent/... ./stage3/census/predicates/... -count=1 -v -timeout=10m` (`tests.log`). Predicates has no Go tests.
- Generated latent overlay (`overlay-generate.log`) and `go build -buildvcs=false -overlay=/workspace/scratch/k031ey0-overlay/overlay.json -o /workspace/scratch/k031ey0-census ./stage3/census/latent/tool` (`overlay-build.log`).
- `go test -tags=census_overlay -overlay=/workspace/scratch/k031ey0-overlay/overlay.json ./stage3/census/latent/... -count=1 -v -timeout=10m` (`overlay-tests.log`). Whole latent package tree, including tagged census tests and tool overlay tests.
- Standalone guard suite (`guards.log`), including the three in-memory red mutants.
- `python3 -m unittest discover -s stage3/census/latent -p '*test.py'` (5 tests; `python-latent.log`).
- `python3 -m unittest discover -s stage3/census/pair -p test_pair.py` (22 tests; `python-pair.log`).
- Baseline overlay generation and build (`base-overlay-generate.log`, `base-overlay-build.log`).
- Six fixture invocations: `python3 run_fixture.py BINARY MODE LABEL` with modes `first-error`, `full`, `speculative`; labels `base-generic` and `head-generic`. All exit 0. Each child has a 60-second timeout, asserts no usable production output, and enables scoped tail measurement.
- `python3 check_fixture.py` (`fixture-check.log`): both revisions have zero checker diagnostics. Baseline has 0 matching NotYets in all modes. Head has 4 / 4 / 6 in first-error / full / speculative, all phase prepass. Every affected attempted unit is excluded from body-lowered coverage and executes no tail pass.
- `git diff --check` and parent distance check (one commit).

## Fixture

`fixture/main.ts` retains the exact Grow, N and Id definitions from the production test's “in itself” row. It uses a declaration-only generic predicate and an independent scalar unit. This exercises the same growing view target, while keeping the baseline from entering unrelated predicate-body and boundary-schema recursion after silently dropping the prepass error. The generic declaration is deferred by census; the new prepass still sees its view.

`base-generic-*.jsonl` and `head-generic-*.jsonl` are the complete streams. Summary logs print the exact finding:

> a view of a type that reaches Grow<Grow<Grow<Grow<Grow<Grow<number>>>>>>, which instantiates itself deeper at each step, so no walk over it ends

## Direct red mutants (each guard child and wrapper exit 1)

`python3 run_mutant.py discarded`: reverts units.go.txt to a bare `base.collectViewedElements(modules)` call. `mutant-discarded.log` names the file and collectViewedElements.

`python3 run_mutant.py omission`: removes noteOmittedOptionals from latent/lower.go.txt. `mutant-omission.log` names the file and prepass.

`python3 run_mutant.py added`: adds fakeCensusPrepass to production's prepass window. `mutant-added.log` reports all four templates. The runner restores each source in finally and returns its child's exact exit code.

## Additional checks and limitations

Predicates overlay generation exits 0, but its build exits 1 with syntax errors in generated internal_lower_refusals.go at lines 479 and 556. The identical build on e64bc123 also exits 1 with the same errors (`predicates-overlay-build.log`, `base-predicates-overlay-build.log`). This existing generator problem is outside this change; the requested predicates package tests completed with no test files.

An additional direct `go test -overlay=... ./internal/lower -run '^TestLatent'` exits 1 because the census generator removes lowerChecked, which unrelated production test files require (`internal-overlay-tests.log`). The supported latent tagged suite above passes and the overlay binary builds.

The original complete production row (saved as original-row.ts.txt) times out at 60 seconds on both revisions in census refusal scanning, before the prepass. An attempted generic cast fixture also times out on both revisions (head-full/base-full logs, exits 124). A declaration-only nongeneric predicate finishes on head (exit 0) but baseline times out in boundary schema construction (exit 124). Bounded SIGQUIT captures (`diagnose.log`, `base-declared-diagnose.log`, exits 124) identify these earlier/later walks. The final generic-declaration fixture avoids both and completes on both revisions. These failed exploratory runs are not the before/after proof.

The first test invocation before submodule initialization exited 1 because cohere/TypeScript/tsc/go.mod was absent. The standalone GOWORK=off overlay attempt exited 1 due to a missing workspace dependency sum; normal workspace generation passed after initialization. npm ci exited 0. No production source or generated dependency metadata is included in the implementation commit.

Build/test processes were bounded at 20 minutes, Go test binaries at 10 minutes, fixture children at 60 seconds, and direct mutant children at two minutes. All wrappers return the child exit code. No force pushes were used.
