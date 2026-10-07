Built: resumed 84f5416, merged main b6b1538, and extended runtime generic coverage; historical census generic 123 to 123, structural 110 to 0.
Commits: original generic 9642bb4, structural 1d5ea34, evidence 84f5416; resume merge 9211260; follow-up commit is the commit containing this report.
Commands/results: setup 124s, nproc 5; focused uncached oracle and four mutants pass; counts regeneration and vet pass; final package gate recorded below.
Mutants: undefined number becomes zero, receiver evaluated twice, instance selects static body, callable lookup after arguments; all caught solely by Node stdout with clean sanitized execution.
Limits: no fresh cumulative census on current main, no full repository gate, original forEach/find remain checker-rejected; optional structural constructor calls and incompatible native signatures remain NotYet.

## What was recovered

Read CLAUDE.md, README.md, docs/0.1.md and docs/memory.md. The report was not lost:
84f5416 committed CENSUS_GENERIC_RETURNS.md and CENSUS_GENERIC_RETURNS.json.
The original implementation files and tests were read in full. Main b6b1538 was
merged into this branch with no conflict. No protected production file was edited.

The prior report records generic counts 48 T returns, 62 T-or-undefined returns,
and 13 U-or-undefined returns, before and after: 123 to 123. Structural calls
beside statics changed 110 to 0. I independently recounted each recorded run's
filtered findings and matched every family count in the JSON. This is verification
of the committed historical ledger, not a new observation of current main.
The old /tmp scratch integrations, raw final census runs and artifact-audit logs
are not present in this resumed environment, and those audits were not rerun.

The generic mapper hook solves concrete call instantiations, including explicit
arguments when a missing input cannot reveal T. A latent census that attempts
uninstantiated generic declarations cannot assign a representation to T; the
historical 123 remaining diagnostics therefore do not measure these runtime
instantiations. The structural blanket refusal was replaced by receiver-dependent
dispatch, with a snapshot before argument evaluation. The existing fixture holds
instance/static identity, inherited static this, single receiver evaluation and
an own callable field replaced during argument evaluation.

The previous worker had not run the full repository gate. Its report also
explicitly excluded original forEach and find acceptance, optional structural
calls on possible constructor objects, and different native signatures. These
limits remain. forEach's unchecked indexed argument and generic truthiness need
proof or a language decision. find's optional negative startIndex prevents its
upper-bound condition alone from proving an element exists. No unchecked cast,
boolean-rule relaxation, or source-body rewrite was added to hide either gap.

## Additional runtime coverage

Extended the existing .a generic fixture, preserving the original firstOrUndefined
and lastOrUndefined bodies. Added a number-or-undefined element instantiation,
plain T identity returns, and a separate project<T,U> helper whose callback returns
U or undefined. The project helper is a new focused probe, not copied tsc source.
It covers number zero and missing, dynamic empty/present strings and missing, and
an object containing a dynamic string. No cohere implementation was copied.

Two exploratory probes received explicit NotYet: a mixed string/number array,
and a callback inferred to return only undefined. The former was removed from
this accepted fixture; the latter was given an explicit number-or-undefined or
string-or-undefined return annotation. Logs of these initial failures are
/tmp/generic-resume-focused.log and the first run of
/tmp/generic-resume-focused-final.log (the final run overwrote that second log).
Neither limitation is claimed fixed.

Counts for the expanded fixture changed from allocations/frees 14/14,
retains/releases 23/41, peak 10, regions 0 to 29/29, 28/64, peak 15, regions 0.
No other count row changed. This is the cost of added input, not a regression
measurement on unchanged input. Corrected the generic.go comment to describe
reading the resolved mapper before fallback signature inference.

## Commands and observed output

Each command sourced /workspace/adamic-tools/env.sh. Test output went directly
to logs and was read afterward. Setup printed Go 1.27.1, clang 20.1.8 and
Node 24.19.0; Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 124s, done 124s. nproc was 5; cpu.max was 400000 100000.
Setup log: /tmp/generic-setup.log.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/census_|TestCensus' -v -count=1 -timeout 30m > /tmp/generic-resume-focused-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/generic-resume-counts-final.log 2>&1
go vet ./... > /tmp/generic-resume-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/generic-resume-gate.log 2>&1
```

Focused oracle exited 0 (2.204s), counts exited 0 (19.601s), vet exited 0 with
no diagnostics, gofmt -l cmd internal and git diff --check printed nothing.
All four committed mutants were rerun. The number mutant printed zero where
Node printed undefined; the receiver mutant printed 4 reads instead of 2;
the static-body mutant printed static-body:instance:instance instead of
instance:instance; the lookup-order mutant printed new:argument instead of
old:argument. Each compiled, exited zero, emitted no sanitizer diagnostics,
and was caught specifically by stdout differs.

Final merged package gate exited 0: internal/lower 24.155s, internal/oracle
100.691s. This includes the complete oracle rather than only the focused
fixtures. The full repository go test ./... gate was not run under the worker
gate exception. No performance claim is made.
