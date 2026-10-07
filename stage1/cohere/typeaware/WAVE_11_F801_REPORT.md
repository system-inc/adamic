Rebased the fifteen implemented wave-11 ports onto current main f8013f0b and added numeric rule.json manifests.
Previous pushed tip 86d5249f was replaced by rebased code tip 4db59750; the evidence/manifest commit follows.
All six worker oracle suites, bridge checks, filtered compiler oracle and package vet passed.
All 34 existing mutant observations passed again; a valid-JSON missing-kinds mutant was caught by the production Go comparison.
Uncovered: three React native source-to-HIR ports, handed-node numeric dispatch, full gate and rule emitted-JavaScript comparison.

## Landing and scope

The only branch this worker pushed is codex/typeaware-wave-11. All 23 worker
commits rebased cleanly onto origin/main f8013f0baac41ddc340d76f83bddde38536a8f07.
The saved range-diff preserves the earlier patches. No shared generator,
shared parser, shared harness, protected compiler files or submodule pin changed
in this unit. Publication is only to the worker branch, using an exact
force-with-lease for its previous pushed SHA 86d5249f07a6a6afa023f5ecd62554c83db7341f.
The explicit user landing instruction authorizes this rebase and republishing.
No main/area branch push, merge or PR is performed.

## Numeric manifests

Fifteen owned wave-11-rules/<rule>/rule.json files contain name and numeric kinds.
They use the schema published by wave 29. The live Go production registration
oracle checks those JSON files and the existing .a listenerKinds declarations;
original-source Node, normal native and sanitized native remain identical to Go.
The manifest mutant writes actual JSON inputs in scratch, deleting the first
rule's kind list while preserving all fifteen names and valid JSON. Parsing
succeeds; only the comparison with unchanged production Go registrations rejects
its missing listener. This tests the manifest consumer rather than mirroring
copied expected numbers. Two existing native declaration mutants run again.

These are metadata declarations for the pending shared driver. ParseNode still
has only kind:string, and existing rule run() loops have not migrated to a
handed-node numeric API. No dispatch speed improvement or full compliance with
the new execution contract is claimed. No new rules were reserved or created.

## React dependency boundary

The three React reservations remain unfinished. Fresh all-heads fetch found
codex/shared-ssa at afa0cb0b7: its document is design only, for Go compiler flow
adapters. Wave 29 now has supplied-graph post-dominance/control kernels and a
static-components taint kernel. Its current CONTROL_REPORT still explicitly
requires native HIR lowering, capture translation, Dispatch alias facts,
compilation-unit/useMemo recognition, ref propagation and memo erasure/inlining.
These supplied-graph kernels do not provide a source graph entry for our ports.
The corresponding reports are retained in the evidence directory. No Go-derived
lint decisions were moved into the bridge and no no-op validators were written.

## Commands and observations

Toolchain setup reused the prior successful run: ready 0s, cache warm 86s,
total 86s; nproc 5. Source /workspace/adamic-tools/env.sh before commands.
All test output went to logs, never through pipes.

Each of six worker tests used go test ./stage1/cohere/typeaware -run <exact-filter>
-count=1 -timeout 15m -v. Full argv vectors and zero exits are in f801-results.json.
Filters: TestWave11AgreementAndMutants, TestWave11NextAgreementAndMutants,
TestWave11ThirdAgreementAndMutants, TestWave11FourthAgreementAndMutants,
TestWave11FifthAgreementAndMutants and TestWave11ListenerDeclarationsAndMutants,
all anchored with ^ and $. Each used its own ADAMIC_WAVE_11_*_ARTIFACTS directory
and ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript. The listener test ran
after the five concurrent rule tests began and included the final JSON checks.

| Suite | Wall seconds |
| --- | ---: |
| first | 247.165 |
| next | 404.467 |
| third | 236.227 |
| fourth | 217.890 |
| fifth | 267.306 |
| listeners | 45.788 |

The five rule batches again compared all finding/fix/suggestion bytes on controls,
the frozen 287 repository roots and 77 compiler roots, normal and ASAN/UBSAN/LSAN.
Controls produced 53/67/191/71/259 findings, totaling 641. All default suggestion
fields are empty. Fifth compares 374 of 376 controls and retains the two logged
strict-parser octal exclusions. Every batch reran released-handle panic-70 checks
and a released-registry mutant. This is not a new sixth-batch React validation.

Additional commands:

```
go test -v -count=1 -timeout 10m ./bridge/tsgo/checker ./bridge/tsgo
go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(sorting|string_index|functions|closures|devirtualize|call_targets_(closure|element|region|reuse|sort)|library_map_set_iterator_.*|047cb0d_n_.*)\.a$'
go vet ./stage1/cohere/typeaware ./bridge/tsgo/...
git diff --check
```

Bridge tests pass, including ownership/sanitizer mutants. The filtered compiler
oracle passes with original-source Node, native and emitted JavaScript, including
new main's iterator and narrowed-element fixtures and the one-byte comparison
mutant. This is compiler coverage, not JavaScript execution of the lint rules.
Vet and whitespace checks produce empty output. The full repository gate and
inherited baseline rule tests were not rerun in this landing unit.

## Every worker mutant observation

| Batch | Observation |
| --- | --- |
| first | regexp mutant: exit 0, empty stderr, independent Go bytes catch byte 461 |
| first | array mutant: exit 0, empty stderr, independent Go bytes catch byte 13795 |
| first | branches mutant: exit 0, empty stderr, independent Go bytes catch byte 16505 |
| first | signature mutant: exit 0, empty stderr, independent Go bytes catch byte 20504 |
| first | index mutant: exit 0, empty stderr, independent Go bytes catch byte 13792 |
| first | syntax mutant: exit 0, empty stderr, independent Go bytes catch byte 1320 |
| first | released-registry mutant exits 0, caught by required panic 70 |
| next | collection mutant: exit 0, empty stderr, independent Go bytes catch byte 63 |
| next | outcome mutant: exit 0, empty stderr, independent Go bytes catch byte 24413 |
| next | pure mutant: exit 0, empty stderr, independent Go bytes catch byte 14629 |
| next | lineage mutant: exit 0, empty stderr, independent Go bytes catch byte 13263 |
| next | awaited mutant: exit 0, empty stderr, independent Go bytes catch byte 25310 |
| next | released-registry mutant exits 0, caught by required panic 70 |
| third | eval mutant: exit 0, empty stderr, independent Go bytes catch byte 60 |
| third | extend mutant: exit 0, empty stderr, independent Go bytes catch byte 19666 |
| third | assign mutant: exit 0, empty stderr, independent Go bytes catch byte 27917 |
| third | global mutant: exit 0, empty stderr, independent Go bytes catch byte 1750 |
| third | anchor mutant: exit 0, empty stderr, independent Go bytes catch byte 35301 |
| third | released-registry mutant exits 0, caught by required panic 70 |
| fourth | func mutant: exit 0, empty stderr, independent Go bytes catch byte 62 |
| fourth | nonconstructor mutant: exit 0, empty stderr, independent Go bytes catch byte 6511 |
| fourth | wrappers mutant: exit 0, empty stderr, independent Go bytes catch byte 7660 |
| fourth | global mutant: exit 0, empty stderr, independent Go bytes catch byte 59 |
| fourth | released-registry mutant exits 0, caught by required panic 70 |
| fifth | throw mutant: exit 0, empty stderr, independent Go bytes catch byte 723 |
| fifth | backreference mutant: exit 0, empty stderr, independent Go bytes catch byte 120 |
| fifth | arrow-fix mutant: exit 0, empty stderr, independent Go bytes catch byte 10702 |
| fifth | provenance mutant: exit 0, empty stderr, independent Go bytes catch byte 116 |
| fifth | regex-path mutant: exit 0, empty stderr, independent Go bytes catch byte 4578 |
| fifth | self-resolution mutant: exit 0, empty stderr, independent Go bytes catch byte 10418 |
| fifth | symbol provenance question mutant: exit 0, empty stderr, independent Go bytes catch byte 116 |
| fifth | released-registry mutant exits 0, caught by required panic 70 |
| listeners | manifest-missing-kinds: valid JSON, production Go comparison catches byte 120 |
| listeners | wrong-kind: exit 0, empty stderr, Go listener comparison catches byte 79 |
| listeners | missing-listener: exit 0, empty stderr, Go listener comparison catches byte 406 |

## Native time against Go

Concurrent process timings include contention; they do not establish a quiet
speed ratio or improvement. Previous implementation reports retain quiet runs.

```
first: repository go process=563.640761ms cohere: load_ns=318616709 rule_ns=1402381 run_ns=182043715
first: repository native process=697.68445ms tsgo: load_ns=234620331 query_ns=0 queries=0 first_query_ns=0 run_ns=446995682
first: compiler go process=761.582514ms cohere: load_ns=566851570 rule_ns=27300978 run_ns=161639734
first: compiler native process=3.537554363s tsgo: load_ns=739262224 query_ns=51070223 queries=74 first_query_ns=7647017 run_ns=2780409271
next: repository go process=319.040924ms cohere: load_ns=133778072 rule_ns=90282832 run_ns=163129817
next: repository native process=817.156529ms tsgo: load_ns=164471692 query_ns=182678696 queries=8128 first_query_ns=4381200 run_ns=642806562
next: compiler go process=803.712366ms cohere: load_ns=238494709 rule_ns=488540021 run_ns=542838692
next: compiler native process=38.848694263s tsgo: load_ns=247055604 query_ns=28318636579 queries=91563 first_query_ns=5500538 run_ns=38557850900
third: repository go process=263.034ms cohere: load_ns=121654828 rule_ns=18952944 run_ns=118125677
third: repository native process=792.369521ms tsgo: load_ns=163737772 query_ns=129028327 queries=643 first_query_ns=141023 run_ns=623703608
third: compiler go process=1.737123682s cohere: load_ns=1172458213 rule_ns=293815296 run_ns=525064237
third: compiler native process=5.928665458s tsgo: load_ns=550238982 query_ns=666002331 queries=9190 first_query_ns=7443813 run_ns=5302049291
fourth: repository go process=247.226443ms cohere: load_ns=160078761 rule_ns=1033941 run_ns=68626468
fourth: repository native process=746.293795ms tsgo: load_ns=243413289 query_ns=0 queries=0 first_query_ns=0 run_ns=471198105
fourth: compiler go process=648.189996ms cohere: load_ns=459930369 rule_ns=7146143 run_ns=149759859
fourth: compiler native process=3.850887896s tsgo: load_ns=654214332 query_ns=213809079 queries=2 first_query_ns=3304246 run_ns=3159970157
fifth: repository go process=386.172731ms cohere: load_ns=204598665 rule_ns=9262428 run_ns=161701074
fifth: repository native process=907.689161ms tsgo: load_ns=218009793 query_ns=13423951 queries=25 first_query_ns=604642 run_ns=668372018
fifth: compiler go process=447.235037ms cohere: load_ns=296249717 rule_ns=72917436 run_ns=112359323
fifth: compiler native process=2.591649645s tsgo: load_ns=243665306 query_ns=301490068 queries=460 first_query_ns=5884369 run_ns=2310369565
```

Exact streams, generated controls/manifests, mutant JSON, logs, source hashes,
commands/results and range-diff are in validation-wave-11-f801. Paths and bytes
are not normalized. The final worker response records the pushed commit SHA.
