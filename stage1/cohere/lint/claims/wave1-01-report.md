Built: foundation merges, the pushed pre-code claim and reproducible blocker probes; no new rule port.
Commits: registration merge 19c6554; helper merge d135f63; claim fe63109; evidence commit follows.
Commands and outputs: setup retry succeeded in 41s, nproc 5; registry and upstream tests passed; probes confirmed both blockers.
Mutants: no assigned-rule semantic mutant was run because no assigned rule was implemented; foundation checks are reported below.
Not covered: assigned-rule Node/emitted-JS/native parity, TypeScript compiler and stage1 corpus comparisons, or native/Node/Go findings per second.

## Assignment and duplicate ports

The 46-rule list is in HELPERS.md, referenced by helpers/REPORT.md, rather than
being printed in REPORT.md itself. Its first three names are consistent-type-assertions,
consistent-type-definitions and init-declarations, all under @typescript-eslint.

The claim was pushed at fe63109 before any implementation work. Definitions and
init-declarations are skipped because origin/codex/stage1-lint-batch4-typescript,
63782c53678711717f9fd4f12762ce3d103b9a3d, contains their implementations. The latter
also exists on origin/codex/stage1-lint-batch4, d486b03a15f3202acc317dc81c40cc21818e21f7.
These are observed source implementations, not a certification of their parity.

## Observations and scope blockers

1. The directory registration foundation only reads rule.ts, generates imports
   ending in rule.ts and admits mutant files ending in .ts. The copied no-debugger
   directory with its module renamed to rule.a exits 1 at registration:
   `open .../rules/no-debugger/rule.ts: no such file or directory`.
   This prevents registering an implementation confined to its own directory
   while honoring the explicit requirement to write Adamic files only as .a.
2. Finding and RuleContext represent one replacement and one suggestion. The
   shared Go oracle rejects more than one suggestion or a suggestion with more
   than one edit. A temporary Go overlay selects the real, unmodified upstream
   ConsistentTypeAssertions rule and decoder. The shared oracle builds with exit 0,
   then exits 2 with `panic: unexpected suggestion shape` for
   `const x = {} as Foo;` and
   `{"assertionStyle":"as","objectLiteralTypeAssertions":"never"}`.
   The upstream rule emits both annotation and satisfies suggestions; annotation
   inserts the binding type and replaces the assertion, two separate edits.
3. docs/parallel-work.md is absent from main and both requested foundation tips.
   docs/lint-registration.md and the merged CLAUDE.md provide the available
   directory ownership and registration contract.
4. The foundations do not merge cleanly at main d090af5: six conflicts occurred
   in lint README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go.
   The helper branch inherits an older scanner-based driver. The merge retains
   the directory-registration driver and brings in helper and inventory additions,
   excluding incompatible historical scanner-driver additions. This is an explicit
   merge resolution, not a claim that the scanner's extra rules were integrated.

Inference: complete parity needs shared registration, module discovery and mutant
support for .a, plus complete fix/suggestion arrays and serialization in the
shared context, finding model and comparison oracle. Those changes are outside
this unit's rule directories and the instruction against changing shared lists.
No placeholder registration or reduced findings-only port is presented as complete.
The assignment remains blocked pending those shared contracts.

## Environment and reproducibility

Source /workspace/adamic-tools/env.sh in every shell. Toolchain: Go 1.27.1,
clang 20.1.8 and Node 24.19.0. nproc: 5.

The initial fetch refspec selected only main. Fetching all heads supplied the
foundation and duplicate-port refs. Its redundant recursive submodule fetch was
stopped after the branch refs were available; setup separately completed the
required pinned submodules.

The first setup ran while merge resolution removed obsolete scanner files and
failed its warm build with `open stage1/cohere/lint/profile_test.go: no such file or directory`.
This was this worker's avoidable sequencing error. After merge resolution,
`bash cloud/setup.sh > /tmp/lint-wave1-01-setup-retry.log 2>&1` succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (41s)
setup: done in 41s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`python3 stage1/cohere/lint/claims/wave1-01-evidence/probes.py > /tmp/lint-wave1-01-probes.log 2>&1`
exits 0 after asserting extension exit 1, Go build exit 0, suggestion exit 2 and
the exact panic. The probes only copy baseline modules into scratch directories
and build an overlay; production files are untouched.

## Validation

All test output was written directly to log files. Evidence logs accompany this
report. The full repository gate was not run.

- `go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-01-registry.log 2>&1`:
  PASS, 0.039s, including existing descriptor rejection mutants.
- From cohere: `go test ./internal/lint/rules/typescript -run '^TestConsistentTypeAssertions' -count=1 -v -timeout=10m > /tmp/lint-wave1-01-upstream.log 2>&1`:
  PASS, 0.268s. This independently verifies Go's own corpus, including suggestions;
  it does not compare an Adamic implementation.
- `go vet ./... > /tmp/lint-wave1-01-vet.log 2>&1`: exit 0, empty log.
- `git diff --check > /tmp/lint-wave1-01-diffcheck.log 2>&1`: exit 0, empty log.
- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint ./stage1/cohere/lint/helpers ./stage1/cohere/lint/registry -count=1 -v -timeout=20m > /tmp/lint-wave1-01-packages.log 2>&1`:
  exit 0. Lint PASS 250.948s, helpers PASS 151.435s, registry PASS 0.086s.
  The existing five-rule corpus includes 217 upstream combinations and produces
  128,446 identical bytes across Go, source Node and sanitized native. Owned
  witnesses and decoded options also match. This is foundation validation only.
  Compiler/stage1 corpus and throughput tests explicitly skipped because their
  opt-in environment variables were not supplied; no assigned rule exists to measure.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint-wave1-01-oracle.log 2>&1`:
  exit 0, PASS 5.808s, six fixtures, six probe misses and zero cache hits.

## Every foundation mutant observed in this run

No mutant in this table certifies the unimplemented assigned rule. The blocker
probes are contract failures, not comparison-only semantic mutants.

| Existing mutant | What caught it |
|---|---|
| no-debugger fix suppressed | Node and native output comparison to Go |
| no-empty reports empty function bodies | Node and native output comparison to Go |
| eqeqeq applies a suggestion as a fix | Node and native output comparison to Go |
| no-var suppresses var declarations | Node and native output comparison to Go |
| no-duplicate-case inverts the seen-case test | Node and native output comparison to Go |
| registration subscribes no-debugger to the wrong kind | Node and native output comparison to Go |
| omit a factory finish hook | Node and native output comparison |
| rewrite an outside import relative to the root | Node and native module loading; not credited as a semantic comparison mutant |
| ignore decoded allowemptycatch | Node and native output comparison to Go |
| OptionsJson admits raw controls | Native output comparison to Go, first difference at line 15157 |
| OptionSchema permits overlapping oneOf matches | Native output comparison to Go, line 15166 |
| PolicyMessage omits value interpolation | Native output comparison to Go, line 22166 |
| StrictOptions skips unknown fields | Native output comparison to Go, line 15178 |
| duplicate oracle adapter | registry rejection |
| duplicate public name | registry rejection |
| invalid node kind | registry rejection |
| unknown descriptor field | registry rejection |
| missing visit listener | registry rejection |
| missing finish hook | registry rejection |
| missing named hook | registry rejection |
| missing factory | registry rejection |
| missing class | registry rejection |
| unsafe public name | registry rejection |
| missing oracle export | registry rejection |

The last eleven are structural rejection probes, not runtime semantic mutants.
The package log preserves the exact comparisons and rejection reasons.

## Not covered

No .a rule implementation, emitted JavaScript comparison for an assigned rule,
compiler/stage1 corpus run, assigned-rule mutant, findings-per-second timing,
full repository gate or independent revalidation of the two skipped ports.
No compiler files were edited. No pull request was opened. The branch was pushed.
