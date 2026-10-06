Built: ownership claim and reproducible foundation-blocker evidence; zero rules ported.
Commits: registration merge 0c9db2c, helpers merge 208185c, pushed claim f2ecf28; report commit follows.
Commands: registry tests PASS; registration generation exit 0; lint tests and setup exit 1; filtered uncached input oracle PASS (10.295s).
Mutants: no rule mutants run; controlled identical-source .ts-to-.a probe accepted baseline and rejected .a registration.
Not covered: all three implementations, findings/fixes parity corpora, emitted JavaScript parity, per-rule mutants and throughput; full gate not run.

# Lint wave 1 slot 15: blocked foundation handoff

The branch is codex/lint-wave1-15, based on origin/main d090af5.
The claim was pushed before any implementation. No pull request was opened.
The three rules are positions 43 through 45 in HELPERS.md, the list referenced
by helpers/REPORT.md:

43. nexus/import-require-node-namespace
44. structure/network-no-invalidate-cache-literal-key
45. structure/network-no-string-literal-query

No corresponding implementation filename was found in stage1 on any fetched
origin branch. This is a filename observation, not an assertion of complete
semantic duplicate detection. No rule was skipped as already ported.

## Observed blockers

The requested .a rule cannot be registered by the supplied foundation.
registry/registry.go reads rule.ts at line 95, generates rule.ts imports at
line 187, and rejects a mutant file unless its extension is .ts at line 151.
The test harness copies only .ts modules at lint_test.go line 47, rewrites only
.ts imports at line 403, and selects only .ts corpus files at line 365.

The controlled probe copies the existing registered rules to scratch, runs the
actual generator successfully, renames the existing no-debugger/rule.ts to
rule.a without changing its contents, then runs the same generator. The second
run exits 1 with rule.ts: no such file or directory. This probe has an accepted
positive control; no zero-finding result is inferred from an absent measurement.
Reproduce with:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/claims/wave1-15-evidence/probe-a.py > /tmp/lint-wave1-15-a.log 2>&1
```

This is a foundation extension probe, not one of the requested semantic rule
mutants. No new Adamic program, .ts workaround, or incomplete rule descriptor
was committed. Fixing the generator and harness requires shared-file edits
outside the rule directories; authorization was requested and was not received
before this report.

The merges were not both clean. Registration merged cleanly. Helpers conflicted
in README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go.
The resolution retained the registration versions of those six files and brought
in the helpers branch's other changes. Those changes include older lint profiling
code: profile_test.go line 32 ranges over portFiles, whereas registration changed
portFiles into a function. Consequently the lint test package fails compilation:

```
profile_test.go:32:23: cannot range over portFiles (value of type func(t *testing.T) []string)
```

Other semantic interactions of these foundations have not been validated.
The merge is a blocked integration artifact, not a claim of a green baseline.

The requested docs/parallel-work.md does not exist on origin/main or either
foundation branch. The available directory contract is docs/lint-registration.md.

## Toolchain and bounded checks

nproc printed 5. Go is 1.27.1, clang is 20.1.8 and Node is 24.19.0.
The environment script is /workspace/adamic-tools/env.sh, not /opt/adamic-tools/env.sh.
An initial attempt to source the latter failed; the bounded checks were repeated
with the actual script. Both setup attempts exited 1 during test-cache warming.
The initial run overlapped the merges and is preserved in setup.log; the stable
rerun is setup-stable.log and establishes the profile_test.go failure above.
Its complete timing lines before failure were:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
```

No build-cache-warm or done timing line was printed. The workaround sourced the
already installed tools and ran unaffected checks. Test output went directly to
log files and was read afterward:

- go test ./stage1/cohere/lint/registry -count=1 -v: PASS, 0.009s, including descriptor rejection controls.
- go run ./cmd/lint-registry: exit 0, listed five existing rules.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v: exit 1, package compilation failed before the test ran.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 5m: PASS, 10.295s, six input fixtures, zero cache hits, six probe misses.
- Controlled extension probe: exit 0 overall, .ts baseline exit 0, identical .a source generator exit 1 as expected.

The initial checkout fetched only main. An explicit all-heads fetch populated
remote branches but continued fetching unrelated historical TypeScript submodule
commits. That recursive fetch was stopped after the parent refs were available;
a non-recursive parent fetch was then run. Setup successfully initialized the
pinned submodules.

No findings-per-second measurement was run for native, Node or Go. There is no
ported rule artifact to measure; reporting zero would misrepresent missing data.
No rule mutant was compiled or run. No findings or fixes were compared for the
assigned rules. The filtered input oracle validates existing compiler behavior,
not these three rules. No forbidden compiler file was edited.

To resume, the foundations need .a discovery/import/copy/mutant/corpus support
and a compatible lint profiling harness, either delivered by their owners or
with explicit expanded territory for this unit. Then the three .a rule ports
and all requested comparisons and semantic mutants remain to be done.
