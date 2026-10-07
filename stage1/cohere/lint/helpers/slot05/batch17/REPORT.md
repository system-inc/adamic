Built visitUnknown, forkOptionalChain and react.IsHookCall as three separate .a helper files.
SHAs: published claim 07dd8fe8 before source; landing base 902255da; implementation/publication commit reported in final response.
Commands and outputs: 18,960 Go/source/emitted/native cases PASS 21.641s; repository vet and format PASS; six-fixture Node input oracle PASS 0.917s; setup 38s, nproc 5.
Mutants: all sixteen compiling semantic variants caught, with every first Go divergence archived in evidence/mutants.json.
Not covered: full dependent rules, external dependency implementations, invalid parser/state representations, full repository gate and its seventeen required external-input correctness checks.

# Ownership and readiness

All 44 retained helpers were complete, tested and pushed before this claim. Current main 39638d9e and area d65a8f93 remain ancestors and unchanged since the earlier full retained gate. All twenty helper branches were fetched and every claim inspected. Shared comments bundle retains all higher fan-out leaves. These three concrete symbols tie the highest unclaimed count at four consumers each. Their reservation was pushed before source. The final refreshed twenty-branch scan finds no competing claim. No fourth helper is claimed.

The two CFG helpers each serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. IsHookCall serves structure/consistency-require-matching-file-name, structure/react-component-no-destructuring, structure/react-component-no-separate-named-export and structure/react-component-require-properties-parameter. Twelve prerequisite occurrences removed across eight rules, zero final blocker removed. readiness.json records the calculated cumulative inventory. This is dependency readiness under the frozen inventory's common adapter assumption, not implemented rule findings.

# Validation observations

The comparator invokes actual Go helpers at pin 715ba94f3608a6500086b1076ce5cb7e51b836db, with dependency calls renamed only in temporary overlays. The real dependency executes before its wrapper restores the outer callback trace. Go supplies parser classifications and external leaf verdicts; Adamic owns dispatch, ordering, identity and the listed state transitions. README.md defines the boundary. No compiler, production cohere, shared harness or rule registry was edited.

All four consumers of each helper have their Go test string literals inspected. CFG corpus: 166 array-callback-return occurrences, 272 consistent-return, 71 no-unreachable-loop and 555 rules-of-hooks; 801 distinct strings after controls, 12,452 nodes including nil. visitUnknown observes every node. forkOptionalChain covers actual optional roots and source-file negatives, stack depths zero through three and both reachability values: 6,472 states. Hook corpus: 221, 93, 105 and 73 occurrences respectively in the four structure consumers, 355 distinct strings, 4,943 parsed nodes and 36 calls including two absence controls. Coverage JSON and generated corpus SHA-256 files are archived.

Final suite: 18,960 rows PASS across Go, source Node, emitted JavaScript and sanitized native. Native must exit zero with empty stderr. Semantic mutants are credited only after normal compilation and execution; no refusal or crash is a semantic catch. All tests in this bounded gate executed with no SKIP output. The six input probes have zero cache hits and six misses.

Sixteen mutants were caught: visitUnknown ignores nil, ignores statement classification, sends expressions to statement processing, or continues into expr after a statement; optional chain reverses guard short-circuit order, rejects one-item stacks, ignores the root predicate, selects the first join, omits continuation linking, or omits entering; hook call accepts absent input, accepts unsupported callees, ignores identifiers, ignores property accesses, negates the name result, or negates the namespace result. Every exact mutation and first divergence is retained in mutants.json.

Superseded runs were not credited: the original generic N | undefined signature was refused even with explicit type arguments; nil detection was represented through the opaque-node adapter within the owned helper. The next run refused a private assertion closure; its identity checks were inlined. The unchanged compiler then accepted all helpers and driver code. Initial parity passed thirteen variants; final focused guard witnesses brought the count to sixteen. All superseded logs are retained. No shared code was changed to work around either refusal.

# Commands

```
bash cloud/setup.sh > /tmp/lint05-batch17-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH17_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch17/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch17 -count=1 -v -timeout=20m > /tmp/lint05-batch17-final.log 2>&1
go vet ./... > /tmp/lint05-batch17-final-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch17 > /tmp/lint05-batch17-format.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch17-oracle.log 2>&1
```

Setup timing: Go, clang, Node and submodules ready 0s; cache warm 38s; done 38s on five processors, cpu.max 400000 100000, 17.6 GB. Versions Go 1.27.1, clang 20.1.8, Node 24.19.0. Repository vet and formatting logs are empty. Logs were written directly, never piped. This bounded worker gate does not claim the full stage1 gate or its seventeen required external-input comparisons. No skipped correctness check was treated as green or altered.
