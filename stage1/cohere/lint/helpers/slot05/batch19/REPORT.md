# Batch nineteen report

Built whileStatement, doStatement and labeledStatement in separate .a files. Each serves array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Twelve prerequisite occurrences removed, zero final blockers; no rule is declared ported. Exact mappings and cumulative counts are in readiness.json and CONSUMERS.md.

## Claim and landing

Original claim a37f0d47 was pushed before any new source; rebased claim 15d5ae60. Original source 614bf5f1, rebased source cc632e36. All twenty origin codex/lint-helpers* branches were fetched and every claim inspected. The larger comments bundle is already reserved in shared HELPERS.md; these helpers tie the highest unclaimed concrete count, four consumers each. A final refresh found no competing claim.

Main c7991b90 remains current. The area advanced during validation from b84a9d93 to b4691483, integrating legacy rule migration. The branch was rebased cleanly onto that area, accepting its shared changes. No main or area push was made. Compiler, runtime, Go module, pinned cohere, Node oracle, options parser, readiness, inventory and all retained slot05 helper trees are byte-identical by Git object identity; evidence/landing-input-identity.json records them. All changes to helper oracle inputs are the new owned batch19 directory. The prior eighteen-package gate and its 179 compiling semantic mutant witnesses remain in batch18/evidence/landing-helpers.log; they were not rerun in this unit. The new trio and filtered input oracle were rerun uncached after rebasing.

## Observed comparisons

The overlay executes the actual original Go method bodies at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, renaming only calls to dependencies for observation. Each dependency still executes real Go. Nested dependency instrumentation is muted to keep the helper's own call trace separate. Condition/body/expression effects are exported as adapter data; source Node, emitted JavaScript and ASan/UBSan native use those effects while executing the port's own allocations, links, entries, jump stack and ordering. Comparisons include final current block, reachability, incoming flags, ordered successor edges, remaining jump prefix, exact label/destination/hook flags and call trace. These are helper comparisons, not complete rule diagnostic comparisons.

All four consumers' test files supply 166, 272, 71 and 555 captured Go string occurrences respectively. With targeted controls, the corpus contains 801 distinct sources and 12,482 parsed nodes. Only relevant statement kinds enter each typed helper: 32 while nodes, 17 do nodes and 21 label nodes. Each is run with reachable/unreachable entry and zero, one or two existing jumps: 192 while, 102 do and 126 label rows, **420 total**. Controls include constant and dynamic tests, short circuit and ternary expressions, return/throw/break/continue bodies, nested labels, switches and Unicode labels. Consumer counts mean their full captured source set was parsed; not every captured string contains each statement kind.

## Commands and results

All test output went directly to logs, never through a pipe. bash cloud/setup.sh succeeded: Go 0s, clang 0s, Node 1s, submodules 1s, build cache 39s, done 39s on nproc 5 (four-core cgroup quota, 17.6 GB). Sourced /workspace/adamic-tools/env.sh. Versions Go 1.27.1, clang 20.1.8, Node 24.19.0.

```sh
ADAMIC_SLOT05_BATCH19_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch19/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch19 -count=1 -v -timeout=20m > /tmp/lint05-batch19-landing-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch19-landing-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch19-landing-oracle.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch19 > /tmp/lint05-batch19-format.log
```

Initial strengthened trio PASS 34.088s; landing trio PASS 29.625s, including all eighteen mutants. Vet and formatting pass with empty logs. Six external input fixtures pass uncached, zero cache hits; exact timing in evidence/oracle.log. git diff --check passes. The source and emitted-JavaScript baselines are byte-for-byte equal to the pinned Go answers, as is sanitized native. Successful baseline and mutant runs must exit zero with empty stderr.

## Every mutant

Each mutant is built independently from a temporary copy, compiles successfully and runs with zero exit and empty stderr before its output is compared. Refusal, sanitizer error or clang warnings do not count. All eighteen were caught by the Go comparator; evidence/mutants.json gives every exact first differing line, byte and pair of outputs.

| Helper | Six independent mutations caught |
|---|---|
| whileStatement | invert truthy routing; omit iteration hook; backedge to body; continue destination to body; condition true destination to after; iteration hook before body |
| doStatement | body destination and node hook instead of test and absent hook; invert exit routing; omit test expression; omit iteration hook; backedge to test; iteration hook before test expression |
| labeledStatement | invert breakable delegation; mark label target breakable; change exact label text; omit after edge; omit after entry; pop target before body |

## Findings and limits

The first driver used a captured restoration closure that the ownership checker refused as cycle-capable; moving it into a function declaration satisfied the check. One preliminary do mutant anchor matched both the initial body link and the backedge; it was replaced with a unique surrounding anchor. Those failed runs are retained as superseded logs, not credited as passes. No shared compiler or harness was changed.

Not covered: complete findings, spans, fixes or suggestions for the four rules; integration into the shared rule registry; external dependency implementations; arbitrary forged/malformed parser nodes or inconsistent adapters; arbitrary user hooks changing the graph; exhaustive nested AST/stack combinations. The full repository gate, including the seventeen required stage1 external-input checks, was not run. This is the authorized bounded touched-package and filtered-oracle gate. No skip was counted as green, and no check was relaxed or removed.
